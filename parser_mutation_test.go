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

package zapscript_test

import (
	"errors"
	"testing"

	"github.com/ZaparooProject/go-zapscript"
	"github.com/google/go-cmp/cmp"
)

// ============================================================================
// advargs.go mutations - IsActionRun, IsActionDetails, IsModeShuffle
// ============================================================================

func TestIsActionRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action string
		want   bool
	}{
		// Empty string should return true (the mutation target)
		{name: "empty string returns true", action: "", want: true},
		// "run" should return true (case-insensitive)
		{name: "run lowercase", action: "run", want: true},
		{name: "RUN uppercase", action: "RUN", want: true},
		{name: "Run mixed case", action: "Run", want: true},
		// Other values should return false
		{name: "details returns false", action: "details", want: false},
		{name: "stop returns false", action: "stop", want: false},
		{name: "random string", action: "xyz", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := zapscript.IsActionRun(tt.action)
			if got != tt.want {
				t.Errorf("IsActionRun(%q) = %v, want %v", tt.action, got, tt.want)
			}
		})
	}
}

func TestIsActionDetails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action string
		want   bool
	}{
		{name: "details lowercase", action: "details", want: true},
		{name: "DETAILS uppercase", action: "DETAILS", want: true},
		{name: "Details mixed case", action: "Details", want: true},
		{name: "empty string returns false", action: "", want: false},
		{name: "run returns false", action: "run", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := zapscript.IsActionDetails(tt.action)
			if got != tt.want {
				t.Errorf("IsActionDetails(%q) = %v, want %v", tt.action, got, tt.want)
			}
		})
	}
}

func TestIsModeShuffle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode string
		want bool
	}{
		{name: "shuffle lowercase", mode: "shuffle", want: true},
		{name: "SHUFFLE uppercase", mode: "SHUFFLE", want: true},
		{name: "Shuffle mixed case", mode: "Shuffle", want: true},
		{name: "empty string returns false", mode: "", want: false},
		{name: "random returns false", mode: "random", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := zapscript.IsModeShuffle(tt.mode)
			if got != tt.want {
				t.Errorf("IsModeShuffle(%q) = %v, want %v", tt.mode, got, tt.want)
			}
		})
	}
}

// ============================================================================
// reader.go mutations - ScriptReader methods
// ============================================================================

func TestParseEscapeSequences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Test all escape sequences
		{
			name:  "escape newline",
			input: `**cmd:a^nb`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"a\nb"}},
				},
			},
		},
		{
			name:  "escape carriage return",
			input: `**cmd:a^rb`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"a\rb"}},
				},
			},
		},
		{
			name:  "escape tab",
			input: `**cmd:a^tb`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"a\tb"}},
				},
			},
		},
		{
			name:  "escape caret",
			input: `**cmd:a^^b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"a^b"}},
				},
			},
		},
		{
			name:  "escape double quote",
			input: `**cmd:a^"b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{`a"b`}},
				},
			},
		},
		{
			name:  "escape single quote",
			input: `**cmd:a^'b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"a'b"}},
				},
			},
		},
		// Escape at end of input (EOF after caret)
		{
			name:  "caret at end of input",
			input: `**cmd:test^`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"test^"}},
				},
			},
		},
		// Unknown escape sequence (just outputs the char)
		{
			name:  "unknown escape sequence",
			input: `**cmd:a^xb`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"axb"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseScript() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCheckEndOfCmd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantCmds int
	}{
		// Double pipe ends command
		{name: "double pipe separates commands", input: "**cmd1||**cmd2", wantCmds: 2},
		// Single pipe does not end command (becomes part of arg)
		{name: "single pipe in arg", input: "**cmd:a|b", wantCmds: 1},
		// Triple pipe (double + single)
		{name: "triple pipe", input: "**cmd1|||**cmd2", wantCmds: 2},
		// Pipe at end
		{name: "trailing pipe", input: "**cmd:test|", wantCmds: 1},
		// Double pipe at end
		{name: "trailing double pipe", input: "**cmd:test||", wantCmds: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if err != nil {
				t.Fatalf("ParseScript() unexpected error: %v", err)
			}
			if len(got.Cmds) != tt.wantCmds {
				t.Errorf("got %d commands, want %d", len(got.Cmds), tt.wantCmds)
			}
		})
	}
}

func TestParseQuotedArgEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Escape in quoted arg - ^" escapes to just closing quote
		{
			name:  "escaped quote in quoted arg",
			input: `**cmd:"test^""`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{`test"`}},
				},
			},
		},
		// Expression in quoted arg
		{
			name:  "expression in quoted arg",
			input: `**cmd:"[[var]]"`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{zapscript.TokExpStart + "var" + zapscript.TokExprEnd}},
				},
			},
		},
		// Unmatched expression in quoted arg
		{
			name:    "unmatched expression in quoted arg",
			input:   `**cmd:"[[var"`,
			wantErr: zapscript.ErrUnmatchedExpression,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseScript() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// ============================================================================
// expressions.go mutations - parseExpression, ParseExpressions, EvalExpressions
// ============================================================================

func TestParseExpressionEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		name    string
		input   string
		want    string
	}{
		// Single bracket is not expression start
		{
			name:  "single bracket is literal",
			input: "[hello",
			want:  "[hello",
		},
		// Closing bracket without opening
		{
			name:  "single closing bracket",
			input: "]hello",
			want:  "]hello",
		},
		// Expression followed by text
		{
			name:  "expression then text",
			input: "[[var]]end",
			want:  zapscript.TokExpStart + "var" + zapscript.TokExprEnd + "end",
		},
		// Text then expression
		{
			name:  "text then expression",
			input: "start[[var]]",
			want:  "start" + zapscript.TokExpStart + "var" + zapscript.TokExprEnd,
		},
		// Multiple expressions
		{
			name:  "multiple expressions",
			input: "[[a]][[b]][[c]]",
			want: zapscript.TokExpStart + "a" + zapscript.TokExprEnd +
				zapscript.TokExpStart + "b" + zapscript.TokExprEnd +
				zapscript.TokExpStart + "c" + zapscript.TokExprEnd,
		},
		// Expression with bracket inside (first ]] ends it)
		{
			name:  "bracket inside expression ends at first close",
			input: "[[a[0]]]",
			want:  zapscript.TokExpStart + "a[0" + zapscript.TokExprEnd + "]",
		},
		// Unmatched opening
		{
			name:    "unmatched opening bracket",
			input:   "[[unclosed",
			wantErr: zapscript.ErrUnmatchedExpression,
		},
		// Escaped expression
		{
			name:  "escaped expression",
			input: "^[[notexpr]]",
			want:  "[[notexpr]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseExpressions()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseExpressions() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if got != tt.want {
				t.Errorf("ParseExpressions() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEvalExpressionsEdgeCases(t *testing.T) {
	t.Parallel()

	env := zapscript.ArgExprEnv{
		Platform: "test",
		Version:  "1.0.0",
	}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		// String only
		{name: "string only", input: "hello", want: "hello"},
		// Expression only
		{name: "expression only", input: zapscript.TokExpStart + "2+2" + zapscript.TokExprEnd, want: "4"},
		// Mixed string and expression
		{name: "mixed", input: "result: " + zapscript.TokExpStart + "10*5" + zapscript.TokExprEnd, want: "result: 50"},
		// Multiple expressions
		{
			name: "multiple expressions",
			input: zapscript.TokExpStart + "1+1" + zapscript.TokExprEnd +
				" and " + zapscript.TokExpStart + "2+2" + zapscript.TokExprEnd,
			want: "2 and 4",
		},
		// Boolean expression
		{name: "boolean true", input: zapscript.TokExpStart + "true" + zapscript.TokExprEnd, want: "true"},
		{name: "boolean false", input: zapscript.TokExpStart + "false" + zapscript.TokExprEnd, want: "false"},
		// Float expression
		{name: "float", input: zapscript.TokExpStart + "3.14" + zapscript.TokExprEnd, want: "3.14"},
		// Environment variable
		{name: "env platform", input: zapscript.TokExpStart + "platform" + zapscript.TokExprEnd, want: "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.EvalExpressions(env)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("EvalExpressions() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ============================================================================
// arguments.go mutations - parseJSONArg, parseInputMacroArg, parseAdvArgs, parseArgs
// ============================================================================

func TestParseJSONArgMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Nested JSON
		{
			name:  "deeply nested JSON",
			input: `**cmd:{"a":{"b":{"c":1}}}`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{`{"a":{"b":{"c":1}}}`}},
				},
			},
		},
		// JSON with escaped quotes
		{
			name:  "JSON with escaped quotes",
			input: `**cmd:{"msg":"say \"hi\""}`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{`{"msg":"say \"hi\""}`}},
				},
			},
		},
		// JSON with braces in string
		{
			name:  "JSON with braces in string",
			input: `**cmd:{"text":"{not json}"}`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{`{"text":"{not json}"}`}},
				},
			},
		},
		// Invalid JSON (missing closing brace)
		{
			name:    "invalid JSON missing brace",
			input:   `**cmd:{"key":"value"`,
			wantErr: zapscript.ErrInvalidJSON,
		},
		// JSON with array
		{
			name:  "JSON with array",
			input: `**cmd:{"items":[1,2,3]}`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{`{"items":[1,2,3]}`}},
				},
			},
		},
		// Empty JSON object
		{
			name:  "empty JSON object",
			input: `**cmd:{}`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{`{}`}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseScript() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseInputMacroArgMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Escape sequence followed by EOF
		{
			name:  "backslash at very end",
			input: `**input.keyboard:\`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.keyboard", Args: []string{"\\"}},
				},
			},
		},
		// Mixed extended macros and chars
		{
			name:  "mixed macros and chars",
			input: `**input.keyboard:a{enter}b{tab}c`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.keyboard", Args: []string{"a", "{enter}", "b", "{tab}", "c"}},
				},
			},
		},
		// Command separator in input
		{
			name:  "double pipe ends input",
			input: `**input.keyboard:abc||**cmd:x`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.keyboard", Args: []string{"a", "b", "c"}},
					{Name: "cmd", Args: []string{"x"}},
				},
			},
		},
		// Invalid adv arg in input macro (fallback to chars) - key must start with letter
		{
			name:  "invalid adv arg key starting with symbol becomes input chars",
			input: `**input.keyboard:x?@b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.keyboard", Args: []string{"x", "?", "@", "b"}},
				},
			},
		},
		// Numeric key is valid in adv args (unlike traits)
		{
			name:  "numeric adv arg key is valid",
			input: `**input.keyboard:x?123=val`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "input.keyboard",
						Args:    []string{"x"},
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"123": "val"}),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseScript() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseAdvArgsMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Quote not at value start (should be literal)
		{
			name:  "quote not at value start",
			input: `**cmd?key=a"b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", AdvArgs: zapscript.NewAdvArgs(map[string]string{"key": `a"b`})},
				},
			},
		},
		// JSON not at value start (should be literal)
		{
			name:  "brace not at value start",
			input: `**cmd?key=a{b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", AdvArgs: zapscript.NewAdvArgs(map[string]string{"key": `a{b`})},
				},
			},
		},
		// Escape sequence in value
		{
			name:  "escape tab in value",
			input: `**cmd?key=a^tb`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", AdvArgs: zapscript.NewAdvArgs(map[string]string{"key": "a\tb"})},
				},
			},
		},
		// Expression in value
		{
			name:  "expression in adv arg value",
			input: `**cmd?key=[[var]]`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name: "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{
							"key": zapscript.TokExpStart + "var" + zapscript.TokExprEnd,
						}),
					},
				},
			},
		},
		// Multiple values with expressions
		{
			name:  "multiple expressions in values",
			input: `**cmd?a=[[x]]&b=[[y]]`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", AdvArgs: zapscript.NewAdvArgs(map[string]string{
						"a": zapscript.TokExpStart + "x" + zapscript.TokExprEnd,
						"b": zapscript.TokExpStart + "y" + zapscript.TokExprEnd,
					})},
				},
			},
		},
		// Adv arg ends at command separator
		{
			name:  "adv args end at separator",
			input: `**cmd1?key=val||**cmd2`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd1", AdvArgs: zapscript.NewAdvArgs(map[string]string{"key": "val"})},
					{Name: "cmd2"},
				},
			},
		},
		// Quoted value with ampersand
		{
			name:  "quoted value containing ampersand",
			input: `**cmd?key="a&b"`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", AdvArgs: zapscript.NewAdvArgs(map[string]string{"key": "a&b"})},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseScript() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseArgsMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Quote at arg start vs middle
		{
			name:  "quote at arg start",
			input: `**cmd:"quoted"`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"quoted"}},
				},
			},
		},
		{
			name:  "quote in middle of arg is literal",
			input: `**cmd:a"b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{`a"b`}},
				},
			},
		},
		// JSON at arg start vs middle
		{
			name:  "JSON at arg start",
			input: `**cmd:{"key":"val"}`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{`{"key":"val"}`}},
				},
			},
		},
		{
			name:  "brace in middle of arg is literal",
			input: `**cmd:a{b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{`a{b`}},
				},
			},
		},
		// Expression in arg
		{
			name:  "expression in arg",
			input: `**cmd:[[var]]`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{zapscript.TokExpStart + "var" + zapscript.TokExprEnd}},
				},
			},
		},
		// Arg separator creates new arg
		{
			name:  "comma separates args",
			input: `**cmd:a,b,c`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"a", "b", "c"}},
				},
			},
		},
		// Whitespace trimming
		{
			name:  "whitespace around args trimmed",
			input: `**cmd:  a  ,  b  `,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"a", "b"}},
				},
			},
		},
		// Empty arg
		{
			name:  "empty arg from colon only",
			input: `**cmd:`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{""}},
				},
			},
		},
		// Escape sequence before comma
		{
			name:  "escaped comma in arg",
			input: `**cmd:a^,b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"a,b"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseScript() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// ============================================================================
// parser.go mutations - parseCommand, ParseScript
// ============================================================================

func TestParseCommandMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Command name with dot
		{
			name:  "command with dot in name",
			input: `**launch.title:game`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"game"}},
				},
			},
		},
		// Command name uppercase converted to lowercase
		{
			name:  "uppercase command name",
			input: `**ECHO:test`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "echo", Args: []string{"test"}},
				},
			},
		},
		// Invalid character in command name (treated as auto-launch)
		{
			name:  "invalid char in cmd name becomes auto-launch",
			input: `**he@llo`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{`**he@llo`}},
				},
			},
		},
		// Empty command name
		{
			name:    "empty command name error",
			input:   `**:arg`,
			wantErr: zapscript.ErrEmptyCmdName,
		},
		// Just asterisks
		{
			name:    "just double asterisk",
			input:   `**`,
			wantErr: zapscript.ErrEmptyCmdName,
		},
		// Single asterisk at EOF
		{
			name:    "single asterisk EOF",
			input:   `*`,
			wantErr: zapscript.ErrUnexpectedEOF,
		},
		// Adv args directly after command name
		{
			name:  "adv args no colon",
			input: `**cmd?key=val`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", AdvArgs: zapscript.NewAdvArgs(map[string]string{"key": "val"})},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseScript() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseScriptMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Whitespace handling
		{
			name:  "leading whitespace",
			input: `   **cmd:arg`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"arg"}},
				},
			},
		},
		// Auto-launch (no ** prefix)
		{
			name:  "auto-launch path",
			input: `/path/to/game.rom`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{`/path/to/game.rom`}},
				},
			},
		},
		// Media title syntax
		{
			name:  "media title syntax",
			input: `@snes/Super Mario World`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{`snes/Super Mario World`}},
				},
			},
		},
		// Media title without slash (becomes auto-launch)
		{
			name:  "media title no slash",
			input: `@noseparator`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{`@noseparator`}},
				},
			},
		},
		// Single asterisk treated as auto-launch prefix
		{
			name:  "single asterisk auto-launch",
			input: `*notacommand`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{`*notacommand`}},
				},
			},
		},
		// Starting with { reserved for JSON (error)
		{
			name:    "starting brace error",
			input:   `{"key":"value"}`,
			wantErr: zapscript.ErrInvalidJSON,
		},
		// Traits command merges into script.Traits
		{
			name:  "traits command",
			input: `**traits:{"a":1}`,
			want: zapscript.Script{
				Traits: map[string]any{"a": float64(1)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseScript() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// ============================================================================
// traits.go mutations - parseTraitsSyntax, parseTraitValue, parseTraitArray
// ============================================================================

func TestParseTraitsMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr    error
		wantTraits map[string]any
		name       string
		input      string
		wantCmds   int
	}{
		// Quoted value with escaped quote inside
		{
			name:       "quoted value with escaped quote",
			input:      `#key="test^"end"`,
			wantTraits: map[string]any{"key": `test"end`},
		},
		// Unquoted value escape at end
		{
			name:       "unquoted escape at end",
			input:      `#key=val^`,
			wantTraits: map[string]any{"key": "val^"},
		},
		// Multiple traits with tabs between
		{
			name:       "traits with tab separator",
			input:      "#a=1\t#b=2",
			wantTraits: map[string]any{"a": int64(1), "b": int64(2)},
		},
		// Trait key with numbers in middle
		{
			name:       "key with numbers",
			input:      `#key123=val`,
			wantTraits: map[string]any{"key123": "val"},
		},
		// Boolean shorthand before another trait
		{
			name:       "boolean before value trait",
			input:      `#flag #val=123`,
			wantTraits: map[string]any{"flag": true, "val": int64(123)},
		},
		// Value terminated by #
		{
			name:       "value ends at hash",
			input:      `#a=1#b=2`,
			wantTraits: map[string]any{"a": int64(1), "b": int64(2)},
		},
		// Single pipe ends trait value (becomes part of next command parsing)
		{
			name:       "single pipe ends trait value",
			input:      `#key=a||#b=2`,
			wantTraits: map[string]any{"key": "a", "b": int64(2)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseScript() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if len(got.Cmds) != tt.wantCmds {
				t.Errorf("got %d commands, want %d", len(got.Cmds), tt.wantCmds)
			}
			if diff := cmp.Diff(tt.wantTraits, got.Traits); diff != "" {
				t.Errorf("traits mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseTraitsArrayMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr    error
		wantTraits map[string]any
		name       string
		input      string
	}{
		// Array with whitespace
		{
			name:       "array with newline in whitespace",
			input:      "#arr=[\n1,\n2]",
			wantTraits: map[string]any{"arr": []any{int64(1), int64(2)}},
		},
		// Array with mixed whitespace
		{
			name:       "array with tab whitespace",
			input:      "#arr=[\t1,\t2\t]",
			wantTraits: map[string]any{"arr": []any{int64(1), int64(2)}},
		},
		// Quoted element with escaped quote inside
		{
			name:       "quoted array element with escaped quote",
			input:      `#arr=["test^"end"]`,
			wantTraits: map[string]any{"arr": []any{`test"end`}},
		},
		// Unquoted array element with escape
		{
			name:       "unquoted array escape",
			input:      `#arr=[a^nb]`,
			wantTraits: map[string]any{"arr": []any{"a\nb"}},
		},
		// Array EOF without close bracket
		{
			name:    "array EOF error",
			input:   `#arr=[1,2`,
			wantErr: zapscript.ErrUnmatchedArrayBracket,
		},
		// Array element quoted EOF error
		{
			name:    "quoted array element EOF",
			input:   `#arr=["test`,
			wantErr: zapscript.ErrUnmatchedQuote,
		},
		// Array with trailing comma style
		{
			name:       "array elements no spaces",
			input:      `#arr=[a,b,c]`,
			wantTraits: map[string]any{"arr": []any{"a", "b", "c"}},
		},
		// Escape in unquoted element (caret followed by ] escapes the ])
		{
			name:    "caret before close bracket escapes it",
			input:   `#arr=[test^]`,
			wantErr: zapscript.ErrUnmatchedArrayBracket,
		},
		// Escape followed by more content in unquoted element
		{
			name:       "unquoted array element escape newline in middle",
			input:      `#arr=[a^nb]`,
			wantTraits: map[string]any{"arr": []any{"a\nb"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseScript() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if diff := cmp.Diff(tt.wantTraits, got.Traits); diff != "" {
				t.Errorf("traits mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestInferTypeMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		want       any
		wantTraits map[string]any
		name       string
		input      string
	}{
		// Integer
		{name: "zero", input: "#n=0", wantTraits: map[string]any{"n": int64(0)}},
		{name: "positive int", input: "#n=42", wantTraits: map[string]any{"n": int64(42)}},
		{name: "negative int", input: "#n=-5", wantTraits: map[string]any{"n": int64(-5)}},
		// Float
		{name: "zero float", input: "#n=0.0", wantTraits: map[string]any{"n": float64(0)}},
		{name: "positive float", input: "#n=3.14", wantTraits: map[string]any{"n": float64(3.14)}},
		// Boolean
		{name: "true", input: "#b=true", wantTraits: map[string]any{"b": true}},
		{name: "false", input: "#b=false", wantTraits: map[string]any{"b": false}},
		// String (quoted forces string even for numbers)
		{name: "quoted number string", input: `#s="123"`, wantTraits: map[string]any{"s": "123"}},
		{name: "quoted bool string", input: `#s="true"`, wantTraits: map[string]any{"s": "true"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.wantTraits, got.Traits); diff != "" {
				t.Errorf("traits mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// ============================================================================
// Media title mutations
// ============================================================================

func TestMediaTitleMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Valid media title
		{
			name:  "basic media title",
			input: `@snes/Mario`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Mario"}},
				},
			},
		},
		// Media title with advanced args
		{
			name:  "media title with adv args",
			input: `@snes/Mario?action=details`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "launch.title",
						Args:    []string{"snes/Mario"},
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"action": "details"}),
					},
				},
			},
		},
		// Media title with escape
		{
			name:  "media title with escape",
			input: `@snes/Mario^nWorld`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Mario\nWorld"}},
				},
			},
		},
		// Empty system (fallback to auto-launch)
		{
			name:  "empty system fallback",
			input: `@/game`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{"@/game"}},
				},
			},
		},
		// Empty title (fallback to auto-launch)
		{
			name:  "empty title fallback",
			input: `@system/`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{"@system/"}},
				},
			},
		},
		// Media title ends at command separator
		{
			name:  "media title chain",
			input: `@snes/Game||**echo:done`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Game"}},
					{Name: "echo", Args: []string{"done"}},
				},
			},
		},
		// Invalid adv arg in media title (fallback)
		{
			name:  "media title invalid adv arg",
			input: `@snes/Game?-invalid`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Game?-invalid"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseScript() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
