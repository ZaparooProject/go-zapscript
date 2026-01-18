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
	"errors"
	"fmt"
	"strings"
)

func (sr *ScriptReader) parseMediaTitleSyntax() (*mediaTitleParseResult, error) {
	result := &mediaTitleParseResult{
		advArgs: make(map[string]string),
	}
	rawContent := ""

	var contentBuilder strings.Builder
	for {
		ch, readErr := sr.read()
		if readErr != nil {
			return nil, readErr
		}

		if ch == eof {
			break
		}

		// Check for escape sequences
		if ch == SymEscapeSeq {
			next, escapeErr := sr.parseEscapeSeq()
			if escapeErr != nil {
				return nil, escapeErr
			}
			if next == "" {
				_, _ = contentBuilder.WriteString(string(SymEscapeSeq))
			} else {
				_, _ = contentBuilder.WriteString(next)
			}
			continue
		}

		// Check for end of command
		eoc, checkErr := sr.checkEndOfCmd(ch)
		if checkErr != nil {
			return nil, checkErr
		} else if eoc {
			break
		}

		// Check for advanced args start (?)
		if ch == SymAdvArgStart {
			// Parse advanced args (? already consumed)
			parsedAdvArgs, buf, err := sr.parseAdvArgs()
			if errors.Is(err, ErrInvalidAdvArgName) {
				// Fallback: treat as part of content
				_, _ = contentBuilder.WriteString(string(SymAdvArgStart) + buf)
				continue
			} else if err != nil {
				return nil, err
			}

			result.advArgs = parsedAdvArgs
			break
		}

		_, _ = contentBuilder.WriteString(string(ch))
	}
	rawContent += contentBuilder.String()

	result.rawContent = strings.TrimSpace(rawContent)

	// Validate: must contain at least one / separator for system/title format
	sepIdx := strings.Index(result.rawContent, string(SymMediaTitleSep))
	if sepIdx == -1 {
		// Not valid media title format, return for auto-launch fallback
		result.valid = false
		return result, nil
	}

	// Validate: both system ID and game name must be non-empty
	systemID := strings.TrimSpace(result.rawContent[:sepIdx])
	gameName := strings.TrimSpace(result.rawContent[sepIdx+1:])
	if systemID == "" || gameName == "" {
		// Empty system or game name, return for auto-launch fallback
		result.valid = false
		return result, nil
	}

	result.valid = true
	return result, nil
}

func (sr *ScriptReader) parseCommand(onlyOneArg bool) (Command, string, error) {
	cmd := Command{}
	var buf []rune

commandLoop:
	for {
		ch, err := sr.read()
		if err != nil {
			return cmd, string(buf), err
		} else if ch == eof {
			break commandLoop
		}

		buf = append(buf, ch)

		eoc, err := sr.checkEndOfCmd(ch)
		if err != nil {
			return cmd, string(buf), err
		} else if eoc {
			break commandLoop
		}

		switch {
		case isCmdName(ch):
			cmd.Name += string(ch)
		case ch == SymArgStart || ch == SymAdvArgStart:
			// parse arguments
			if cmd.Name == "" {
				break commandLoop
			}

			onlyAdvArgs := false
			if ch == SymAdvArgStart {
				// roll it back to trigger adv arg parsing in parseArgs
				err := sr.unread()
				if err != nil {
					return cmd, string(buf), err
				}
				onlyAdvArgs = true
			}

			var args []string
			var advArgs map[string]string
			var err error

			if isInputMacroCmd(cmd.Name) {
				args, advArgs, err = sr.parseInputMacroArg()
				if err != nil {
					return cmd, string(buf), err
				}
			} else {
				args, advArgs, err = sr.parseArgs("", onlyAdvArgs, onlyOneArg)
				if err != nil {
					return cmd, string(buf), err
				}
			}

			if len(args) > 0 {
				cmd.Args = args
			}

			if len(advArgs) > 0 {
				cmd.AdvArgs = NewAdvArgs(advArgs)
			}

			break commandLoop
		default:
			// might be a launch cmd
			return cmd, string(buf), ErrInvalidCmdName
		}
	}

	if cmd.Name == "" {
		return cmd, string(buf), ErrEmptyCmdName
	}

	cmd.Name = strings.ToLower(cmd.Name)

	return cmd, string(buf), nil
}

func (sr *ScriptReader) ParseScript() (Script, error) {
	script := Script{}

	parseErr := func(err error) error {
		return fmt.Errorf("parse error at %d: %w", sr.pos, err)
	}

	parseAutoLaunchCmd := func(prefix string) error {
		args, advArgs, err := sr.parseArgs(prefix, false, true)
		if err != nil {
			return parseErr(err)
		}
		cmd := Command{
			Name: ZapScriptCmdLaunch,
			Args: args,
		}
		if len(advArgs) > 0 {
			cmd.AdvArgs = NewAdvArgs(advArgs)
		}
		script.Cmds = append(script.Cmds, cmd)
		return nil
	}

	for {
		ch, err := sr.read()
		if err != nil {
			return script, err
		} else if ch == eof {
			break
		}

		switch {
		case isWhitespace(ch):
			continue
		case sr.pos == 1 && ch == SymJSONStart:
			// reserve starting { as json script for later
			return Script{}, ErrInvalidJSON
		case ch == SymMediaTitleStart:
			// Media title syntax: @System Name/Game Title (optional tags)?advArgs
			result, err := sr.parseMediaTitleSyntax()
			if err != nil {
				return script, parseErr(err)
			}

			// If not valid media title format (no / found), treat as auto-launch
			if !result.valid {
				if autoErr := parseAutoLaunchCmd(string(SymMediaTitleStart) + result.rawContent); autoErr != nil {
					return script, parseErr(autoErr)
				}
				continue
			}

			// Build launch.title command with raw content
			// The command layer will handle system lookup and tag extraction
			cmd := Command{
				Name: ZapScriptCmdLaunchTitle,
				Args: []string{result.rawContent},
			}

			// Only set AdvArgs if there are any
			if len(result.advArgs) > 0 {
				cmd.AdvArgs = NewAdvArgs(result.advArgs)
			}

			script.Cmds = append(script.Cmds, cmd)
			continue
		case ch == SymCmdStart:
			next, err := sr.peek()
			if err != nil {
				return script, parseErr(err)
			}

			switch next {
			case eof:
				return script, ErrUnexpectedEOF
			case SymCmdStart:
				if skipErr := sr.skip(); skipErr != nil {
					return script, parseErr(skipErr)
				}
			default:
				// assume it's actually an auto launch cmd
				if autoErr := parseAutoLaunchCmd("*"); autoErr != nil {
					return script, parseErr(autoErr)
				}
				continue
			}

			cmd, buf, err := sr.parseCommand(false)
			switch {
			case errors.Is(err, ErrInvalidCmdName):
				// assume it's actually an auto launch cmd
				if autoErr := parseAutoLaunchCmd("**" + buf); autoErr != nil {
					return script, parseErr(autoErr)
				}
				continue
			case err != nil:
				return script, parseErr(err)
			default:
				script.Cmds = append(script.Cmds, cmd)
			}

			continue
		default:
			err := sr.unread()
			if err != nil {
				return script, parseErr(err)
			}

			err = parseAutoLaunchCmd("")
			if err != nil {
				return script, parseErr(err)
			}

			continue
		}
	}

	if len(script.Cmds) == 0 {
		return script, ErrEmptyZapScript
	}

	return script, nil
}
