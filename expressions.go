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
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/expr-lang/expr"
)

type ExprEnvDevice struct {
	Hostname string `expr:"hostname" json:"hostname"`
	OS       string `expr:"os" json:"os"`
	Arch     string `expr:"arch" json:"arch"`
}

type ExprEnvLastScanned struct {
	ID    string `expr:"id" json:"id"`
	Value string `expr:"value" json:"value"`
	Data  string `expr:"data" json:"data"`
}

//nolint:tagliatelle // JSON uses snake_case to match expression env naming
type ExprEnvActiveMedia struct {
	LauncherID string `expr:"launcher_id" json:"launcher_id"`
	SystemID   string `expr:"system_id" json:"system_id"`
	SystemName string `expr:"system_name" json:"system_name"`
	Path       string `expr:"path" json:"path"`
	Name       string `expr:"name" json:"name"`
}

// ExprEnvScanned represents the token currently being processed.
type ExprEnvScanned struct {
	ID    string `expr:"id" json:"id"`
	Value string `expr:"value" json:"value"`
	Data  string `expr:"data" json:"data"`
}

// ExprEnvLaunching represents the media about to launch.
//
//nolint:tagliatelle // JSON uses snake_case to match expression env naming
type ExprEnvLaunching struct {
	Path       string `expr:"path" json:"path"`
	SystemID   string `expr:"system_id" json:"system_id"`
	LauncherID string `expr:"launcher_id" json:"launcher_id"`
}

//nolint:tagliatelle // JSON uses snake_case to match expression env naming
type ArgExprEnv struct {
	ActiveMedia  ExprEnvActiveMedia `expr:"active_media" json:"active_media"`
	Device       ExprEnvDevice      `expr:"device" json:"device"`
	LastScanned  ExprEnvLastScanned `expr:"last_scanned" json:"last_scanned"`
	Scanned      ExprEnvScanned     `expr:"scanned" json:"scanned,omitempty"`
	Launching    ExprEnvLaunching   `expr:"launching" json:"launching,omitempty"`
	Platform     string             `expr:"platform" json:"platform"`
	Version      string             `expr:"version" json:"version"`
	ScanMode     string             `expr:"scan_mode" json:"scan_mode"`
	MediaPlaying bool               `expr:"media_playing" json:"media_playing"`
}

//nolint:tagliatelle // JSON uses snake_case to match expression env naming
type CustomLauncherExprEnv struct {
	Platform   string        `expr:"platform" json:"platform"`
	Version    string        `expr:"version" json:"version"`
	Device     ExprEnvDevice `expr:"device" json:"device"`
	MediaPath  string        `expr:"media_path" json:"media_path"`
	Action     string        `expr:"action" json:"action"`
	InstallDir string        `expr:"install_dir" json:"install_dir"`
	ServerURL  string        `expr:"server_url" json:"server_url"`
	SystemID   string        `expr:"system_id" json:"system_id"`
	LauncherID string        `expr:"launcher_id" json:"launcher_id"`
}

func (sr *ScriptReader) parseExpression() (string, error) {
	rawExpr := TokExpStart

	next, err := sr.read()
	if err != nil {
		return rawExpr, err
	} else if next != SymExpressionStart {
		err := sr.unread()
		if err != nil {
			return rawExpr, err
		}
		return string(SymExpressionStart), nil
	}

	for {
		ch, err := sr.read()
		if err != nil {
			return rawExpr, err
		} else if ch == eof {
			return rawExpr, ErrUnmatchedExpression
		}

		if ch == SymExpressionEnd {
			next, err := sr.peek()
			if err != nil {
				return rawExpr, err
			} else if next == SymExpressionEnd {
				rawExpr += TokExprEnd
				err := sr.skip()
				if err != nil {
					return rawExpr, err
				}
				break
			}
		}

		rawExpr += string(ch)
	}

	return rawExpr, nil
}

func (sr *ScriptReader) parsePostExpression() (string, error) {
	rawExpr := ""
	exprEndToken, _ := utf8.DecodeRuneInString(TokExprEnd)

	for {
		ch, err := sr.read()
		if err != nil {
			return rawExpr, err
		} else if ch == eof {
			return rawExpr, ErrUnmatchedExpression
		}

		if ch == exprEndToken {
			break
		}

		rawExpr += string(ch)
	}

	return rawExpr, nil
}

// ParseExpressions parses and converts expressions in the input string from
// [[...]] formatted expression fields to internal expression token delimiters,
// to be evaluated by the EvalExpressions function. This function ONLY parses
// expression symbols and escape sequences, no other ZapScript syntax.
func (sr *ScriptReader) ParseExpressions() (string, error) {
	result := ""

	for {
		ch, err := sr.read()
		if err != nil {
			return result, err
		} else if ch == eof {
			break
		}

		switch ch {
		case SymEscapeSeq:
			next, err := sr.parseEscapeSeq()
			if err != nil {
				return result, err
			}
			result += next
			continue
		case SymExpressionStart:
			exprValue, err := sr.parseExpression()
			if err != nil {
				return result, err
			}
			result += exprValue
			continue
		default:
			result += string(ch)
			continue
		}
	}

	return result, nil
}

func (sr *ScriptReader) EvalExpressions(exprEnv any) (string, error) {
	parts := make([]PostArgPart, 0)
	currentPart := PostArgPart{}

	exprStartToken, _ := utf8.DecodeRuneInString(TokExpStart)

	for {
		ch, err := sr.read()
		if err != nil {
			return "", err
		} else if ch == eof {
			break
		}

		if ch == exprStartToken {
			if currentPart.Type != ArgPartTypeUnknown {
				parts = append(parts, currentPart)
				currentPart = PostArgPart{}
			}

			currentPart.Type = ArgPartTypeExpression
			exprValue, err := sr.parsePostExpression()
			if err != nil {
				return "", err
			}
			currentPart.Value = exprValue

			parts = append(parts, currentPart)
			currentPart = PostArgPart{}

			continue
		}
		currentPart.Type = ArgPartTypeString
		currentPart.Value += string(ch)
		continue
	}

	if currentPart.Type != ArgPartTypeUnknown {
		parts = append(parts, currentPart)
	}

	var result strings.Builder
	for _, part := range parts {
		if part.Type == ArgPartTypeExpression {
			output, err := expr.Eval(part.Value, exprEnv)
			if err != nil {
				return "", fmt.Errorf("failed to evaluate expression %q: %w", part.Value, err)
			}

			switch v := output.(type) {
			case string:
				_, _ = result.WriteString(v)
			case bool:
				_, _ = result.WriteString(strconv.FormatBool(v))
			case int:
				_, _ = result.WriteString(strconv.Itoa(v))
			case float64:
				_, _ = result.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
			default:
				return "", fmt.Errorf("%w: %v (%T)", ErrBadExpressionReturn, v, v)
			}
		} else {
			_, _ = result.WriteString(part.Value)
		}
	}

	return result.String(), nil
}
