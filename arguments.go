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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (sr *ScriptReader) parseJSONArg() (string, error) {
	jsonStr := string(SymJSONStart)
	braceCount := 1
	inString := false
	escaped := false

	var jsonBuilder strings.Builder
	for braceCount > 0 {
		ch, err := sr.read()
		if err != nil {
			return "", err
		} else if ch == eof {
			return "", ErrInvalidJSON
		}

		_, _ = jsonBuilder.WriteString(string(ch))

		if escaped {
			escaped = false
			continue
		}

		if ch == SymJSONEscapeSeq {
			escaped = true
			continue
		}

		if ch == SymJSONString {
			inString = !inString
			continue
		}

		if !inString {
			switch ch {
			case SymJSONStart:
				braceCount++
			case SymJSONEnd:
				braceCount--
			}
		}
	}
	jsonStr += jsonBuilder.String()

	// validate json
	var jsonObj any
	if err := json.Unmarshal([]byte(jsonStr), &jsonObj); err != nil {
		return "", ErrInvalidJSON
	}

	// convert back to string
	normalizedJSON, err := json.Marshal(jsonObj)
	if err != nil {
		return "", ErrInvalidJSON
	}

	return string(normalizedJSON), nil
}

func (sr *ScriptReader) parseInputMacroArg() (args []string, advArgs map[string]string, err error) {
	args = make([]string, 0)
	advArgs = make(map[string]string)
	totalLen := 0

macroLoop:
	for {
		ch, err := sr.read()
		if err != nil {
			return args, advArgs, err
		} else if ch == eof {
			break
		}

		if ch == SymInputMacroEscapeSeq {
			next, readErr := sr.read()
			if readErr != nil {
				return args, advArgs, readErr
			} else if next == eof {
				args = append(args, string(SymInputMacroEscapeSeq))
				break
			}

			totalLen++
			if totalLen > InputMacroMaxKeys {
				return args, advArgs, ErrInputMacroTooLong
			}
			args = append(args, string(next))
			continue
		}

		eoc, err := sr.checkEndOfCmd(ch)
		if err != nil {
			return args, advArgs, err
		} else if eoc {
			break
		}

		switch ch {
		case SymInputMacroExtStart:
			content, readErr := sr.parseInputMacroExtContent()
			if readErr != nil {
				return args, advArgs, readErr
			}
			tokens, expandErr := expandInputMacroExt(content, &totalLen)
			if expandErr != nil {
				return args, advArgs, expandErr
			}
			args = append(args, tokens...)
			continue
		case SymExpressionStart:
			exprValue, exprErr := sr.parseExpression()
			if exprErr != nil {
				return args, advArgs, exprErr
			}
			totalLen++
			if totalLen > InputMacroMaxKeys {
				return args, advArgs, ErrInputMacroTooLong
			}
			args = append(args, exprValue)
			continue
		case SymAdvArgStart:
			newAdvArgs, buf, err := sr.parseAdvArgs()
			if errors.Is(err, ErrInvalidAdvArgName) {
				// if an adv arg name is invalid, fallback on treating it
				// as a list of input args
				for _, r := range string(SymAdvArgStart) + buf {
					totalLen++
					if totalLen > InputMacroMaxKeys {
						return args, advArgs, ErrInputMacroTooLong
					}
					args = append(args, string(r))
				}
				continue
			} else if err != nil {
				return args, advArgs, err
			}

			advArgs = newAdvArgs

			// advanced args are always the last part of a command
			break macroLoop
		default:
			totalLen++
			if totalLen > InputMacroMaxKeys {
				return args, advArgs, ErrInputMacroTooLong
			}
			args = append(args, string(ch))
		}
	}

	return args, advArgs, nil
}

// parseInputMacroExtContent reads characters from the reader until the closing
// SymInputMacroExtEnd ('}') and returns the raw content between the braces.
func (sr *ScriptReader) parseInputMacroExtContent() (string, error) {
	var b strings.Builder
	for {
		ch, err := sr.read()
		if err != nil {
			return "", err
		}
		if ch == eof {
			return "", ErrUnmatchedInputMacroExt
		}
		if ch == SymInputMacroExtEnd {
			break
		}
		_, _ = b.WriteRune(ch)
	}
	return b.String(), nil
}

// expandInputMacroExt parses the raw content between '{' and '}' and returns the
// expanded token slice. totalLen is updated by the number of tokens added so the
// caller can enforce InputMacroMaxKeys across the whole macro.
//
// Grammar inside braces:
//
//	{"text"[*N]}       literal text, optionally repeated N times
//	{text:content[*N]} same using verb form; content is typed literally
//	{delay:dur}        pass through as "{delay:dur}" — interpreted by core emitter
//	{press:key}        pass through as "{press:key}"
//	{release:key}      pass through as "{release:key}"
//	{hold:key[:dur]}   pass through as "{hold:key:dur}"
//	{_key}             sigil sugar for press, passed through
//	{^key}             sigil sugar for release, passed through
//	{~key[:dur]}       sigil sugar for hold, passed through
//	{key[*N]}          key/combo/special, optionally repeated N times
func expandInputMacroExt(content string, totalLen *int) ([]string, error) {
	if content == "" {
		return nil, ErrUnmatchedInputMacroExt
	}

	// Quoted literal: {"text"[*N]}
	if content[0] == '"' {
		text, repeat, err := parseQuotedLiteralWithRepeat(content)
		if err != nil {
			return nil, err
		}
		return expandLiteralChars(text, repeat, totalLen)
	}

	// text: verb — {text:content[*N]}
	if strings.HasPrefix(content, "text:") {
		raw := content[len("text:"):]
		text, repeat, err := parseSuffixRepeat(raw)
		if err != nil {
			return nil, err
		}
		return expandLiteralChars(text, repeat, totalLen)
	}

	// Pass-through verb forms: delay, press, release, hold
	if strings.HasPrefix(content, "delay:") ||
		strings.HasPrefix(content, "press:") ||
		strings.HasPrefix(content, "release:") ||
		strings.HasPrefix(content, "hold:") {
		*totalLen++
		if *totalLen > InputMacroMaxKeys {
			return nil, ErrInputMacroTooLong
		}
		return []string{"{" + content + "}"}, nil
	}

	// Sigil forms: {_key}, {^key}, {~key[:dur]}
	if content != "" {
		switch content[0] {
		case '_', '^', '~':
			*totalLen++
			if *totalLen > InputMacroMaxKeys {
				return nil, ErrInputMacroTooLong
			}
			return []string{"{" + content + "}"}, nil
		}
	}

	// Key / combo / special with optional *N repeat.
	name, repeat, err := parseSuffixRepeat(content)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, ErrInputMacroEmptyKey
	}

	// Single-rune keys are appended without braces (e.g. "a", "*").
	// Multi-rune names need braces so ParseKeyCombo recognises them.
	var token string
	if utf8.RuneCountInString(name) == 1 {
		token = name
	} else {
		token = "{" + name + "}"
	}

	return expandTokenN(token, repeat, totalLen)
}

// parseInputRawArg reads the entire command argument as literal text — no '{}'
// grammar, no '*' repeat, no adv-args. Every rune is a key to type, with
// '\n' mapped to "{enter}" and '\t' mapped to "{tab}". The cap InputMacroMaxKeys
// still applies to prevent runaway sequences.
func (sr *ScriptReader) parseInputRawArg() ([]string, error) {
	args := make([]string, 0)
	totalLen := 0

	for {
		ch, err := sr.read()
		if err != nil {
			return args, err
		}
		if ch == eof {
			break
		}

		eoc, err := sr.checkEndOfCmd(ch)
		if err != nil {
			return args, err
		}
		if eoc {
			break
		}

		totalLen++
		if totalLen > InputMacroMaxKeys {
			return args, ErrInputMacroTooLong
		}

		switch ch {
		case '\n':
			args = append(args, "{enter}")
		case '\t':
			args = append(args, "{tab}")
		default:
			args = append(args, string(ch))
		}
	}

	return args, nil
}

// parseSuffixRepeat splits "content*N" at the LAST '*' followed by a positive
// integer, returning (content, N, nil). If there is no such suffix, it returns
// (s, 1, nil) so callers get a no-op repeat. Returns an error if N > InputMacroMaxRepeat.
func parseSuffixRepeat(s string) (content string, n int, err error) {
	idx := strings.LastIndex(s, "*")
	if idx == -1 {
		return s, 1, nil
	}
	rest := s[idx+1:]
	n64, parseErr := strconv.ParseUint(rest, 10, 64)
	if parseErr != nil || n64 == 0 {
		return s, 1, nil //nolint:nilerr // non-integer after * means * is literal content
	}
	if n64 > uint64(InputMacroMaxRepeat) {
		return "", 0, fmt.Errorf("%w: %d (max %d)", ErrInputMacroRepeatTooLarge, n64, InputMacroMaxRepeat)
	}
	return s[:idx], int(n64), nil
}

// parseQuotedLiteralWithRepeat parses the braces content that starts with '"'.
// Expected form: '"' text '"' ['*' N]. Inside the quotes '"' is escaped as '\"'.
func parseQuotedLiteralWithRepeat(content string) (text string, repeat int, err error) {
	if len(content) < 2 {
		return "", 0, ErrUnmatchedQuote
	}

	// Find the closing quote, honouring \" escapes.
	closeIdx := -1
	for i := 1; i < len(content); i++ {
		if content[i] == '\\' {
			i++ // skip next character — it is escaped
			continue
		}
		if content[i] == '"' {
			closeIdx = i
			break
		}
	}
	if closeIdx == -1 {
		return "", 0, ErrUnmatchedQuote
	}

	rawText := content[1:closeIdx]
	text = strings.ReplaceAll(rawText, `\"`, `"`)

	rest := content[closeIdx+1:]
	if rest == "" {
		return text, 1, nil
	}
	if rest[0] != '*' {
		return "", 0, fmt.Errorf("unexpected content after quoted literal: %q", rest)
	}

	n64, parseErr := strconv.ParseUint(rest[1:], 10, 64)
	if parseErr != nil || n64 == 0 {
		return "", 0, fmt.Errorf("invalid repeat count in quoted literal: %q", rest[1:])
	}
	if n64 > uint64(InputMacroMaxRepeat) {
		return "", 0, fmt.Errorf("%w: %d (max %d)", ErrInputMacroRepeatTooLarge, n64, InputMacroMaxRepeat)
	}
	repeat = int(n64)

	return text, repeat, nil
}

// expandLiteralChars expands text into individual rune tokens, repeated n times.
func expandLiteralChars(text string, n int, totalLen *int) ([]string, error) {
	runes := []rune(text)
	count := len(runes) * n
	*totalLen += count
	if *totalLen > InputMacroMaxKeys {
		return nil, ErrInputMacroTooLong
	}
	result := make([]string, 0, count)
	for range n {
		for _, r := range runes {
			result = append(result, string(r))
		}
	}
	return result, nil
}

// expandTokenN returns a slice of n copies of token.
func expandTokenN(token string, n int, totalLen *int) ([]string, error) {
	*totalLen += n
	if *totalLen > InputMacroMaxKeys {
		return nil, ErrInputMacroTooLong
	}
	result := make([]string, n)
	for i := range result {
		result[i] = token
	}
	return result, nil
}

func (sr *ScriptReader) parseAdvArgs() (advArgs map[string]string, remainingStr string, err error) {
	advArgs = make(map[string]string)
	inValue := false
	currentArg := ""
	currentValue := ""
	valueStart := int64(-1)
	buf := make([]rune, 0, 64)

	storeArg := func() {
		if currentArg != "" {
			currentValue = strings.TrimSpace(currentValue)
			advArgs[currentArg] = currentValue
		}
		currentArg = ""
		currentValue = ""
	}

	for {
		ch, err := sr.read()
		if err != nil {
			return advArgs, string(buf), err
		} else if ch == eof {
			break
		}

		buf = append(buf, ch)

		if inValue {
			switch {
			case valueStart == sr.pos-1 && (ch == SymArgDoubleQuote || ch == SymArgSingleQuote):
				quotedValue, parseErr := sr.parseQuotedArg(ch)
				if parseErr != nil {
					return advArgs, string(buf), parseErr
				}
				currentValue = quotedValue
				continue
			case ch == SymJSONStart && valueStart == sr.pos-1:
				jsonValue, parseErr := sr.parseJSONArg()
				if parseErr != nil {
					return advArgs, string(buf), parseErr
				}
				currentValue = jsonValue
				continue
			case ch == SymEscapeSeq:
				// Peek next char for raw tracking before parseEscapeSeq consumes it
				nextRaw, peekErr := sr.peek()
				if peekErr != nil {
					return advArgs, string(buf), peekErr
				}

				next, escapeErr := sr.parseEscapeSeq()
				if escapeErr != nil {
					return advArgs, string(buf), escapeErr
				} else if next == "" {
					currentValue += string(SymEscapeSeq)
					continue
				}
				buf = append(buf, nextRaw)
				currentValue += next
				continue
			}
		}

		eoc, err := sr.checkEndOfCmd(ch)
		if err != nil {
			return advArgs, string(buf), err
		} else if eoc {
			break
		}

		if ch == SymAdvArgSep {
			storeArg()
			inValue = false
			continue
		} else if ch == SymAdvArgEq && !inValue {
			valueStart = sr.pos
			inValue = true
			continue
		}

		switch {
		case inValue:
			if ch == SymExpressionStart {
				exprValue, err := sr.parseExpression()
				if err != nil {
					return advArgs, string(buf), err
				}
				currentValue += exprValue
			} else {
				currentValue += string(ch)
			}
			continue
		case !isAdvArgName(ch):
			return advArgs, string(buf), ErrInvalidAdvArgName
		default:
			currentArg += string(ch)
		}
	}

	storeArg()

	return advArgs, string(buf), nil
}

func (sr *ScriptReader) parseArgs(
	prefix string,
	onlyAdvArgs bool,
	onlyOneArg bool,
) (args []string, advArgs map[string]string, err error) {
	args = make([]string, 0)
	advArgs = make(map[string]string)
	currentArg := prefix
	argStart := sr.pos
	// tracks whether content was explicitly written, distinguishing
	// "**cmd:" (no content, no arg) from "**cmd:''" (explicit empty arg)
	argWritten := prefix != ""

argsLoop:
	for {
		ch, err := sr.read()
		if err != nil {
			return args, advArgs, err
		} else if ch == eof {
			break argsLoop
		}

		switch {
		case argStart == sr.pos-1 && (ch == SymArgDoubleQuote || ch == SymArgSingleQuote):
			quotedArg, quotedErr := sr.parseQuotedArg(ch)
			if quotedErr != nil {
				return args, advArgs, quotedErr
			}
			currentArg = quotedArg
			argWritten = true
			continue argsLoop
		case argStart == sr.pos-1 && ch == SymJSONStart:
			jsonArg, jsonErr := sr.parseJSONArg()
			if jsonErr != nil {
				return args, advArgs, jsonErr
			}
			currentArg = jsonArg
			argWritten = true
			continue argsLoop
		case ch == SymEscapeSeq:
			// escaping next character
			next, escapeErr := sr.parseEscapeSeq()
			if escapeErr != nil {
				return args, advArgs, escapeErr
			} else if next == "" {
				currentArg += string(SymEscapeSeq)
				argWritten = true
				continue argsLoop
			}
			currentArg += next
			argWritten = true
			continue argsLoop
		}

		eoc, err := sr.checkEndOfCmd(ch)
		if err != nil {
			return args, advArgs, err
		} else if eoc {
			break argsLoop
		}

		switch {
		case !onlyOneArg && ch == SymArgSep:
			// new argument
			currentArg = strings.TrimSpace(currentArg)
			args = append(args, currentArg)
			currentArg = ""
			argStart = sr.pos
			argWritten = false
			continue argsLoop
		case ch == SymAdvArgStart:
			newAdvArgs, buf, err := sr.parseAdvArgs()
			switch {
			case errors.Is(err, ErrInvalidAdvArgName):
				// if an adv arg name is invalid, fallback on treating it
				// as a positional arg with a ? in it
				currentArg += string(SymAdvArgStart) + buf
				continue argsLoop
			case err != nil:
				return args, advArgs, err
			}

			advArgs = newAdvArgs

			// advanced args are always the last part of a command
			break argsLoop
		case ch == SymExpressionStart:
			exprValue, err := sr.parseExpression()
			if err != nil {
				return args, advArgs, err
			}
			currentArg += exprValue
			argWritten = true
			continue argsLoop
		default:
			currentArg += string(ch)
			if !isWhitespace(ch) {
				argWritten = true
			}
			continue argsLoop
		}
	}

	currentArg = strings.TrimSpace(currentArg)
	if !onlyAdvArgs && (currentArg != "" || argWritten) {
		args = append(args, currentArg)
	} else if onlyAdvArgs && currentArg != "" {
		// fallback content from invalid adv args should still be preserved
		args = append(args, currentArg)
	}

	return args, advArgs, nil
}
