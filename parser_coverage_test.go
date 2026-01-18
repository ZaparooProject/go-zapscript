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
// parseInputMacroArg Tests (input.* commands)
// ============================================================================

func TestParseInputMacroArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Basic input macro tests
		{
			name:  "simple input characters",
			input: `**input.keyboard:abc`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.keyboard", Args: []string{"a", "b", "c"}},
				},
			},
		},
		// Input macro escapes: \ followed by char outputs that char literally
		{
			name:  "input with backslash n outputs literal n",
			input: `**input.keyboard:hello\nworld`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.keyboard", Args: []string{"h", "e", "l", "l", "o", "n", "w", "o", "r", "l", "d"}},
				},
			},
		},
		{
			name:  "input with backslash t outputs literal t",
			input: `**input.keyboard:a\tb`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.keyboard", Args: []string{"a", "t", "b"}},
				},
			},
		},
		{
			name:  "input with escaped backslash",
			input: `**input.keyboard:a\\b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.keyboard", Args: []string{"a", "\\", "b"}},
				},
			},
		},
		{
			name:  "input with backslash at end",
			input: `**input.keyboard:abc\`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.keyboard", Args: []string{"a", "b", "c", "\\"}},
				},
			},
		},
		// Extended input macros with curly braces
		{
			name:  "input with extended macro enter",
			input: `**input.keyboard:a{enter}b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.keyboard", Args: []string{"a", "{enter}", "b"}},
				},
			},
		},
		{
			name:  "input with extended macro tab",
			input: `**input.keyboard:{tab}`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.keyboard", Args: []string{"{tab}"}},
				},
			},
		},
		{
			name:  "input with multiple extended macros",
			input: `**input.keyboard:{ctrl}{alt}{delete}`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.keyboard", Args: []string{"{ctrl}", "{alt}", "{delete}"}},
				},
			},
		},
		{
			name:    "input with unmatched extended macro",
			input:   `**input.keyboard:{enter`,
			wantErr: zapscript.ErrUnmatchedInputMacroExt,
		},
		// Input macro with advanced args
		{
			name:  "input with advanced args",
			input: `**input.keyboard:abc?delay=100`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "input.keyboard",
						Args:    []string{"a", "b", "c"},
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"delay": "100"}),
					},
				},
			},
		},
		{
			name:  "input with advanced args multiple",
			input: `**input.keyboard:xy?delay=50&repeat=3`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "input.keyboard",
						Args:    []string{"x", "y"},
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"delay": "50", "repeat": "3"}),
					},
				},
			},
		},
		// Fallback when adv arg name is invalid (treated as input chars)
		{
			name:  "input with question mark treated as char when invalid adv arg",
			input: `**input.keyboard:a?@b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.keyboard", Args: []string{"a", "?", "@", "b"}},
				},
			},
		},
		{
			name:  "input gamepad command",
			input: `**input.gamepad:abxy`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.gamepad", Args: []string{"a", "b", "x", "y"}},
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
// parseQuotedArg Tests
// ============================================================================

func TestParseQuotedArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Basic quoted args
		{
			name:  "double quoted arg",
			input: `**cmd:"hello world"`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"hello world"}},
				},
			},
		},
		{
			name:  "single quoted arg",
			input: `**cmd:'hello world'`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"hello world"}},
				},
			},
		},
		{
			name:  "quoted arg with comma inside",
			input: `**cmd:"a,b,c"`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"a,b,c"}},
				},
			},
		},
		// Escape sequences in quoted args
		{
			name:  "quoted arg with escaped newline",
			input: `**cmd:"hello^nworld"`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"hello\nworld"}},
				},
			},
		},
		{
			name:  "quoted arg with escaped tab",
			input: `**cmd:"col1^tcol2"`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"col1\tcol2"}},
				},
			},
		},
		{
			name:  "quoted arg with escaped caret",
			input: `**cmd:"a^^b"`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"a^b"}},
				},
			},
		},
		{
			name:  "quoted arg with escaped quote",
			input: `**cmd:"say ^"hello^""`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{`say "hello"`}},
				},
			},
		},
		// Expressions in quoted args (tokenized internally)
		{
			name:  "quoted arg with expression",
			input: `**cmd:"value is [[var]]"`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"value is \ue000var\ue001"}},
				},
			},
		},
		{
			name:  "quoted arg with multiple expressions",
			input: `**cmd:"[[a]] and [[b]]"`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"\ue000a\ue001 and \ue000b\ue001"}},
				},
			},
		},
		// Unmatched quote errors
		{
			name:    "unmatched double quote",
			input:   `**cmd:"hello`,
			wantErr: zapscript.ErrUnmatchedQuote,
		},
		{
			name:    "unmatched single quote",
			input:   `**cmd:'hello`,
			wantErr: zapscript.ErrUnmatchedQuote,
		},
		// Mixed quoted and unquoted args
		{
			name:  "mixed quoted and unquoted",
			input: `**cmd:plain,"quoted",another`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"plain", "quoted", "another"}},
				},
			},
		},
		{
			name:  "empty quoted arg",
			input: `**cmd:""`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{""}},
				},
			},
		},
		{
			name:  "quoted arg with single quote inside double quotes",
			input: `**cmd:"it's fine"`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"it's fine"}},
				},
			},
		},
		{
			name:  "quoted arg with double quote inside single quotes",
			input: `**cmd:'say "hello"'`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{`say "hello"`}},
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
// parseAdvArgs Tests
// ============================================================================

func TestParseAdvArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Basic advanced args
		{
			name:  "single advanced arg",
			input: `**cmd?key=value`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"key": "value"}),
					},
				},
			},
		},
		{
			name:  "multiple advanced args",
			input: `**cmd?a=1&b=2&c=3`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"a": "1", "b": "2", "c": "3"}),
					},
				},
			},
		},
		{
			name:  "advanced arg with underscore in key",
			input: `**cmd?my_key=my_value`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"my_key": "my_value"}),
					},
				},
			},
		},
		// Quoted values in advanced args
		{
			name:  "advanced arg with double quoted value",
			input: `**cmd?msg="hello world"`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"msg": "hello world"}),
					},
				},
			},
		},
		{
			name:  "advanced arg with single quoted value",
			input: `**cmd?msg='hello world'`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"msg": "hello world"}),
					},
				},
			},
		},
		{
			name:  "advanced arg with quoted value containing ampersand",
			input: `**cmd?query="a&b"&other=val`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"query": "a&b", "other": "val"}),
					},
				},
			},
		},
		// JSON values in advanced args
		{
			name:  "advanced arg with JSON object value",
			input: `**cmd?data={"key":"value"}`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"data": `{"key":"value"}`}),
					},
				},
			},
		},
		{
			name:  "advanced arg with nested JSON",
			input: `**cmd?data={"outer":{"inner":"val"}}`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"data": `{"outer":{"inner":"val"}}`}),
					},
				},
			},
		},
		{
			name:  "advanced arg with JSON array",
			input: `**cmd?arr={"items":[1,2,3]}`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"arr": `{"items":[1,2,3]}`}),
					},
				},
			},
		},
		// Escape sequences in advanced arg values
		{
			name:  "advanced arg value with escaped caret",
			input: `**cmd?val=a^^b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"val": "a^b"}),
					},
				},
			},
		},
		{
			name:  "advanced arg value with escaped newline",
			input: `**cmd?val=line1^nline2`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"val": "line1\nline2"}),
					},
				},
			},
		},
		{
			name:  "advanced arg value with caret at end",
			input: `**cmd?val=test^`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"val": "test^"}),
					},
				},
			},
		},
		// Expressions in advanced arg values (tokenized internally)
		{
			name:  "advanced arg value with expression",
			input: `**cmd?val=[[myvar]]`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"val": "\ue000myvar\ue001"}),
					},
				},
			},
		},
		{
			name:  "advanced arg value with text and expression",
			input: `**cmd?val=prefix_[[var]]_suffix`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"val": "prefix_\ue000var\ue001_suffix"}),
					},
				},
			},
		},
		// Command with args and advanced args
		{
			name:  "command with positional and advanced args",
			input: `**cmd:arg1,arg2?opt=val`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						Args:    []string{"arg1", "arg2"},
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"opt": "val"}),
					},
				},
			},
		},
		// Invalid advanced arg names in positional args context - becomes part of the arg
		{
			name:  "invalid adv arg name becomes positional arg",
			input: `**cmd:arg?my-key=val`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"arg?my-key=val"}},
				},
			},
		},
		{
			name:  "invalid adv arg with at sign becomes positional arg",
			input: `**cmd:arg?@key=val`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"arg?@key=val"}},
				},
			},
		},
		// Invalid adv arg directly after command name - preserved as positional arg
		{
			name:  "invalid adv arg name after command name becomes arg",
			input: `**cmd?my-key=val`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"?my-key=val"}},
				},
			},
		},
		// Empty values
		{
			name:  "advanced arg with empty value",
			input: `**cmd?key=`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"key": ""}),
					},
				},
			},
		},
		{
			name:  "advanced arg key only no equals",
			input: `**cmd?flag`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"flag": ""}),
					},
				},
			},
		},
		// Chained commands with advanced args
		{
			name:  "chained commands both with advanced args",
			input: `**cmd1?a=1||**cmd2?b=2`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd1",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"a": "1"}),
					},
					{
						Name:    "cmd2",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"b": "2"}),
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

// ============================================================================
// Additional edge case tests
// ============================================================================

func TestParseEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Invalid JSON
		{
			name:    "unmatched JSON brace",
			input:   `**cmd?data={"key":"value"`,
			wantErr: zapscript.ErrInvalidJSON,
		},
		// Unmatched expression
		{
			name:    "unmatched expression in arg",
			input:   `**cmd:[[var`,
			wantErr: zapscript.ErrUnmatchedExpression,
		},
		// Complex nested structures
		{
			name:  "JSON with escaped quotes",
			input: `**cmd?data={"msg":"say \"hello\""}`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "cmd",
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"data": `{"msg":"say \"hello\""}`}),
					},
				},
			},
		},
		// Single bracket is not expression
		{
			name:  "single bracket in arg is literal",
			input: `**cmd:a[b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"a[b"}},
				},
			},
		},
		// Closing bracket without opening is literal
		{
			name:  "closing bracket in arg is literal",
			input: `**cmd:a]b`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"a]b"}},
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
