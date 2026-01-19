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
	"strings"
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

			args = append(args, string(next))
			continue
		}

		eoc, err := sr.checkEndOfCmd(ch)
		if err != nil {
			return args, advArgs, err
		} else if eoc {
			break
		}

		if ch == SymInputMacroExtStart {
			extName := string(ch)
			var extBuilder strings.Builder
			for {
				next, err := sr.read()
				if err != nil {
					return args, advArgs, err
				} else if next == eof {
					return args, advArgs, ErrUnmatchedInputMacroExt
				}

				_, _ = extBuilder.WriteString(string(next))

				if next == SymInputMacroExtEnd {
					break
				}
			}
			extName += extBuilder.String()
			args = append(args, extName)
			continue
		} else if ch == SymAdvArgStart {
			newAdvArgs, buf, err := sr.parseAdvArgs()
			if errors.Is(err, ErrInvalidAdvArgName) {
				// if an adv arg name is invalid, fallback on treating it
				// as a list of input args
				for _, ch := range string(SymAdvArgStart) + buf {
					args = append(args, string(ch))
				}
				continue
			} else if err != nil {
				return args, advArgs, err
			}

			advArgs = newAdvArgs

			// advanced args are always the last part of a command
			break
		}

		args = append(args, string(ch))
	}

	return args, advArgs, nil
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
			continue argsLoop
		case argStart == sr.pos-1 && ch == SymJSONStart:
			jsonArg, jsonErr := sr.parseJSONArg()
			if jsonErr != nil {
				return args, advArgs, jsonErr
			}
			currentArg = jsonArg
			continue argsLoop
		case ch == SymEscapeSeq:
			// escaping next character
			next, escapeErr := sr.parseEscapeSeq()
			if escapeErr != nil {
				return args, advArgs, escapeErr
			} else if next == "" {
				currentArg += string(SymEscapeSeq)
				continue argsLoop
			}
			currentArg += next
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
			continue argsLoop
		default:
			currentArg += string(ch)
			continue argsLoop
		}
	}

	currentArg = strings.TrimSpace(currentArg)
	if !onlyAdvArgs {
		// if a cmd was called with ":" it will always have at least 1 blank arg
		args = append(args, currentArg)
	} else if currentArg != "" {
		// fallback content from invalid adv args should still be preserved
		args = append(args, currentArg)
	}

	return args, advArgs, nil
}
