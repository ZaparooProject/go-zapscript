// Copyright 2026 The Zaparoo Project Contributors.
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zapscript

import (
	"strconv"
	"strings"
)

type traitsParseResult struct {
	traits         map[string]any
	fallback       string
	invalidKeyName string
	invalidKey     bool
}

// parseTraitsSyntax parses trait shorthand syntax: #key=value #key2=value2
// Returns traits map and fallback string. If fallback is non-empty, the input
// should be treated as auto-launch content instead of traits.
func (sr *ScriptReader) parseTraitsSyntax() (*traitsParseResult, error) {
	result := &traitsParseResult{
		traits: make(map[string]any),
	}
	var fallbackBuf strings.Builder
	fallbackBuf.WriteRune(SymTraitsStart)

	for {
		ch, readErr := sr.read()
		if readErr != nil {
			return nil, readErr
		}

		if ch == eof {
			break
		}

		fallbackBuf.WriteRune(ch)

		// First char of key must be a letter
		if !isAdvArgNameStart(ch) {
			// Invalid key start - consume rest and return as fallback
			rest, consumeErr := sr.consumeToEndOfCmd()
			if consumeErr != nil {
				return nil, consumeErr
			}
			fallbackBuf.WriteString(rest)
			result.fallback = fallbackBuf.String()
			result.invalidKey = true
			result.invalidKeyName = fallbackBuf.String()
			return result, nil
		}

		// Read the rest of the key
		key := strings.ToLower(string(ch))
		var keySb strings.Builder
		for {
			next, peekErr := sr.peek()
			if peekErr != nil {
				return nil, peekErr
			}
			if next == eof || !isAdvArgName(next) {
				break
			}
			ch, readErr = sr.read()
			if readErr != nil {
				return nil, readErr
			}
			fallbackBuf.WriteRune(ch)
			keySb.WriteString(strings.ToLower(string(ch)))
		}
		key += keySb.String()

		// Check what comes after the key
		next, peekErr := sr.peek()
		if peekErr != nil {
			return nil, peekErr
		}

		var value any = true // Default for boolean shorthand (#flag)

		switch {
		case next == SymAdvArgEq:
			// Consume the =
			ch, readErr = sr.read()
			if readErr != nil {
				return nil, readErr
			}
			fallbackBuf.WriteRune(ch)

			// Parse the value
			parsedValue, valueStr, parseErr := sr.parseTraitValue()
			if parseErr != nil {
				return nil, parseErr
			}
			fallbackBuf.WriteString(valueStr)
			value = parsedValue
		case next == eof || next == SymCmdSep || next == SymTraitsStart || isWhitespace(next):
			// Valid end of boolean shorthand trait
		default:
			// Invalid character after key (e.g., #my-trait) - fallback to auto-launch
			rest, consumeErr := sr.consumeToEndOfCmd()
			if consumeErr != nil {
				return nil, consumeErr
			}
			fallbackBuf.WriteString(rest)
			result.fallback = fallbackBuf.String()
			result.invalidKey = true
			result.invalidKeyName = key
			return result, nil
		}

		result.traits[key] = value

		// Look for next trait, whitespace, or end
		for {
			next, peekErr = sr.peek()
			if peekErr != nil {
				return nil, peekErr
			}

			if next == eof {
				return result, nil
			}

			// Check for end of command
			if next == SymCmdSep {
				ch, readErr = sr.read()
				if readErr != nil {
					return nil, readErr
				}
				eoc, eocErr := sr.checkEndOfCmd(ch)
				if eocErr != nil {
					return nil, eocErr
				}
				if eoc {
					return result, nil
				}
				// Single | is not end of command, continue
				fallbackBuf.WriteRune(ch)
				continue
			}

			if isWhitespace(next) {
				ch, readErr = sr.read()
				if readErr != nil {
					return nil, readErr
				}
				fallbackBuf.WriteRune(ch)
				continue
			}

			if next == SymTraitsStart {
				// Another trait - consume the # and continue parsing
				ch, readErr = sr.read()
				if readErr != nil {
					return nil, readErr
				}
				fallbackBuf.WriteRune(ch)
				break
			}

			// Unexpected character - shouldn't happen for valid trait syntax
			return result, nil
		}
	}

	return result, nil
}

// parseTraitValue parses a trait value with type inference.
// Returns (parsed value, raw string for fallback, error).
func (sr *ScriptReader) parseTraitValue() (parsedValue any, rawStr string, err error) {
	var rawBuf strings.Builder   // tracks raw input for fallback
	var valueBuf strings.Builder // tracks processed value
	quoted := false
	quoteChar := rune(0)

	// Check if value starts with a quote
	first, err := sr.peek()
	if err != nil {
		return "", "", err
	}

	if first == SymArgDoubleQuote || first == SymArgSingleQuote {
		// Consume the opening quote
		ch, readErr := sr.read()
		if readErr != nil {
			return "", "", readErr
		}
		rawBuf.WriteRune(ch)
		quoted = true
		quoteChar = ch

		// Parse quoted value
		for {
			ch, readErr = sr.read()
			if readErr != nil {
				return "", "", readErr
			}
			if ch == eof {
				return "", rawBuf.String(), ErrUnmatchedQuote
			}
			rawBuf.WriteRune(ch)

			if ch == SymEscapeSeq {
				// Peek next char for raw tracking before parseEscapeSeq consumes it
				nextRaw, peekErr := sr.peek()
				if peekErr != nil {
					return "", rawBuf.String(), peekErr
				}

				escaped, escapeErr := sr.parseEscapeSeq()
				if escapeErr != nil {
					return "", rawBuf.String(), escapeErr
				}
				if escaped == "" {
					// EOF after escape
					valueBuf.WriteRune(SymEscapeSeq)
					continue
				}
				rawBuf.WriteRune(nextRaw)
				valueBuf.WriteString(escaped)
				continue
			}

			if ch == quoteChar {
				// End of quoted string
				return valueBuf.String(), rawBuf.String(), nil
			}

			valueBuf.WriteRune(ch)
		}
	}

	// Check if value is an array
	if first == SymArrayStart {
		return sr.parseTraitArray()
	}

	// Unquoted value - read until whitespace, #, or end of command
	for {
		next, peekErr := sr.peek()
		if peekErr != nil {
			return "", rawBuf.String(), peekErr
		}

		if next == eof || next == SymCmdSep || isWhitespace(next) || next == SymTraitsStart {
			break
		}

		ch, readErr := sr.read()
		if readErr != nil {
			return "", rawBuf.String(), readErr
		}
		rawBuf.WriteRune(ch)

		// Handle escape sequences
		if ch == SymEscapeSeq {
			// Peek next char for raw tracking before parseEscapeSeq consumes it
			nextRaw, peekErr := sr.peek()
			if peekErr != nil {
				return "", rawBuf.String(), peekErr
			}

			escaped, escapeErr := sr.parseEscapeSeq()
			if escapeErr != nil {
				return "", rawBuf.String(), escapeErr
			}
			if escaped == "" {
				valueBuf.WriteRune(SymEscapeSeq)
				continue
			}
			rawBuf.WriteRune(nextRaw)
			valueBuf.WriteString(escaped)
			continue
		}

		valueBuf.WriteRune(ch)
	}

	return inferType(valueBuf.String(), quoted), rawBuf.String(), nil
}

// consumeToEndOfCmd reads all characters until end of command or EOF.
func (sr *ScriptReader) consumeToEndOfCmd() (string, error) {
	var buf strings.Builder
	for {
		ch, err := sr.read()
		if err != nil {
			return buf.String(), err
		}
		if ch == eof {
			break
		}

		eoc, err := sr.checkEndOfCmd(ch)
		if err != nil {
			return buf.String(), err
		}
		if eoc {
			break
		}

		buf.WriteRune(ch)
	}
	return buf.String(), nil
}

// parseTraitArray parses an array value: [a,b,c]
// Returns ([]any, raw string for fallback, error).
func (sr *ScriptReader) parseTraitArray() (parsedValue any, rawStr string, err error) {
	var rawBuf strings.Builder
	elements := make([]any, 0)

	// Consume opening bracket
	ch, readErr := sr.read()
	if readErr != nil {
		return nil, "", readErr
	}
	rawBuf.WriteRune(ch)

	// Parse array elements
	for {
		// Skip whitespace
		for {
			next, peekErr := sr.peek()
			if peekErr != nil {
				return nil, rawBuf.String(), peekErr
			}
			if !isWhitespace(next) {
				break
			}
			ch, readErr = sr.read()
			if readErr != nil {
				return nil, rawBuf.String(), readErr
			}
			rawBuf.WriteRune(ch)
		}

		// Check for end of array or empty array
		next, peekErr := sr.peek()
		if peekErr != nil {
			return nil, rawBuf.String(), peekErr
		}
		if next == eof {
			return nil, rawBuf.String(), ErrUnmatchedArrayBracket
		}
		if next == SymArrayEnd {
			ch, readErr = sr.read()
			if readErr != nil {
				return nil, rawBuf.String(), readErr
			}
			rawBuf.WriteRune(ch)
			return elements, rawBuf.String(), nil
		}

		// Parse element
		elemValue, elemRaw, parseErr := sr.parseArrayElement()
		if parseErr != nil {
			return nil, rawBuf.String() + elemRaw, parseErr
		}
		rawBuf.WriteString(elemRaw)
		elements = append(elements, elemValue)

		// Skip whitespace after element
		for {
			next, peekErr = sr.peek()
			if peekErr != nil {
				return nil, rawBuf.String(), peekErr
			}
			if !isWhitespace(next) {
				break
			}
			ch, readErr = sr.read()
			if readErr != nil {
				return nil, rawBuf.String(), readErr
			}
			rawBuf.WriteRune(ch)
		}

		// Check for separator or end
		next, peekErr = sr.peek()
		if peekErr != nil {
			return nil, rawBuf.String(), peekErr
		}

		if next == eof {
			return nil, rawBuf.String(), ErrUnmatchedArrayBracket
		}
		if next == SymArrayEnd {
			ch, readErr = sr.read()
			if readErr != nil {
				return nil, rawBuf.String(), readErr
			}
			rawBuf.WriteRune(ch)
			return elements, rawBuf.String(), nil
		}
		if next == SymArraySep {
			ch, readErr = sr.read()
			if readErr != nil {
				return nil, rawBuf.String(), readErr
			}
			rawBuf.WriteRune(ch)
			continue
		}

		return nil, rawBuf.String(), ErrUnmatchedArrayBracket
	}
}

// parseArrayElement parses a single array element.
func (sr *ScriptReader) parseArrayElement() (parsedValue any, rawStr string, err error) {
	var rawBuf strings.Builder
	var valueBuf strings.Builder

	// Check if value starts with a quote
	first, err := sr.peek()
	if err != nil {
		return "", "", err
	}

	if first == SymArgDoubleQuote || first == SymArgSingleQuote {
		// Consume the opening quote
		ch, readErr := sr.read()
		if readErr != nil {
			return "", "", readErr
		}
		rawBuf.WriteRune(ch)
		quoteChar := ch

		// Parse quoted value
		for {
			ch, readErr = sr.read()
			if readErr != nil {
				return "", rawBuf.String(), readErr
			}
			if ch == eof {
				return "", rawBuf.String(), ErrUnmatchedQuote
			}
			rawBuf.WriteRune(ch)

			if ch == SymEscapeSeq {
				nextRaw, peekErr := sr.peek()
				if peekErr != nil {
					return "", rawBuf.String(), peekErr
				}

				escaped, escapeErr := sr.parseEscapeSeq()
				if escapeErr != nil {
					return "", rawBuf.String(), escapeErr
				}
				if escaped == "" {
					valueBuf.WriteRune(SymEscapeSeq)
					continue
				}
				rawBuf.WriteRune(nextRaw)
				valueBuf.WriteString(escaped)
				continue
			}

			if ch == quoteChar {
				return valueBuf.String(), rawBuf.String(), nil
			}

			valueBuf.WriteRune(ch)
		}
	}

	// Unquoted value - read until separator, closing bracket, or whitespace
	for {
		next, peekErr := sr.peek()
		if peekErr != nil {
			return "", rawBuf.String(), peekErr
		}

		if next == eof || next == SymArraySep || next == SymArrayEnd || isWhitespace(next) {
			break
		}

		ch, readErr := sr.read()
		if readErr != nil {
			return "", rawBuf.String(), readErr
		}
		rawBuf.WriteRune(ch)

		if ch == SymEscapeSeq {
			nextRaw, peekErr := sr.peek()
			if peekErr != nil {
				return "", rawBuf.String(), peekErr
			}

			escaped, escapeErr := sr.parseEscapeSeq()
			if escapeErr != nil {
				return "", rawBuf.String(), escapeErr
			}
			if escaped == "" {
				valueBuf.WriteRune(SymEscapeSeq)
				continue
			}
			rawBuf.WriteRune(nextRaw)
			valueBuf.WriteString(escaped)
			continue
		}

		valueBuf.WriteRune(ch)
	}

	return inferType(strings.TrimSpace(valueBuf.String()), false), rawBuf.String(), nil
}

// inferType infers the Go type from a string value.
func inferType(value string, quoted bool) any {
	if quoted {
		return value
	}

	if value == "" {
		return ""
	}

	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}

	// Try integer first
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return i
	}

	// Try float
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	return value
}
