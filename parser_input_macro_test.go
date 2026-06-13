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
	"strings"
	"testing"

	zapscript "github.com/ZaparooProject/go-zapscript"
	"github.com/google/go-cmp/cmp"
)

// kbd builds the expected Script for a single **input.keyboard command.
func kbd(args ...string) zapscript.Script {
	return zapscript.Script{Cmds: []zapscript.Command{{Name: "input.keyboard", Args: args}}}
}

// txt builds the expected Script for a single **input.text command.
func txt(args ...string) zapscript.Script {
	return zapscript.Script{Cmds: []zapscript.Command{{Name: "input.text", Args: args}}}
}

// diffOpts is the cmp option used consistently with parser_coverage_test.go.
var diffOpts = cmp.AllowUnexported(zapscript.AdvArgs{})

// ─── Regression test for issue #939 ──────────────────────────────────────────

// TestInputMacro_Issue939_LeadingSpace confirms that the leading space in
// " *MENU" is preserved as a literal arg — the original bug produced only
// "*" because the shift-toggling corrupted the sequence on MiSTer.
func TestInputMacro_Issue939_LeadingSpace(t *testing.T) {
	t.Parallel()
	p := zapscript.NewParser("**input.keyboard: *MENU")
	got, err := p.ParseScript()
	if err != nil {
		t.Fatalf("ParseScript() unexpected error: %v", err)
	}
	want := kbd(" ", "*", "M", "E", "N", "U")
	if diff := cmp.Diff(want, got, diffOpts); diff != "" {
		t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
	}
}

// TestInputMacro_EscapeAtEOF verifies that a trailing backslash at end of
// input is appended as a literal backslash, not silently dropped.
func TestInputMacro_EscapeAtEOF(t *testing.T) {
	t.Parallel()
	p := zapscript.NewParser(`**input.keyboard:hello\`)
	got, err := p.ParseScript()
	if err != nil {
		t.Fatalf("ParseScript() unexpected error: %v", err)
	}
	want := kbd("h", "e", "l", "l", "o", `\`)
	if diff := cmp.Diff(want, got, diffOpts); diff != "" {
		t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
	}
}

// ─── New grammar: repeat, single-char brace, specials ────────────────────────

func TestInputMacroGrammar(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Single-char brace expands without braces; multi-char keeps braces.
		{
			name:  "braced single char expands to bare char",
			input: "**input.keyboard:{a}",
			want:  kbd("a"),
		},
		{
			name:  "braced special keeps braces",
			input: "**input.keyboard:{f1}",
			want:  kbd("{f1}"),
		},
		{
			name:  "braced combo keeps braces",
			input: "**input.keyboard:{ctrl+c}",
			want:  kbd("{ctrl+c}"),
		},
		// Repeat *N.
		{
			name:  "single char repeated 5 times",
			input: "**input.keyboard:{a*5}",
			want:  kbd("a", "a", "a", "a", "a"),
		},
		{
			name:  "special key repeated 3 times",
			input: "**input.keyboard:{enter*3}",
			want:  kbd("{enter}", "{enter}", "{enter}"),
		},
		{
			name:  "combo repeated 3 times",
			input: "**input.keyboard:{ctrl+c*3}",
			want:  kbd("{ctrl+c}", "{ctrl+c}", "{ctrl+c}"),
		},
		{
			name:  "repeat of 1 is a no-op",
			input: "**input.keyboard:{a*1}",
			want:  kbd("a"),
		},
		{
			name:  "asterisk after non-integer is literal",
			input: "**input.keyboard:{a*b}",
			want:  kbd("{a*b}"),
		},
		{
			name:  "repeat in mixed sequence",
			input: "**input.keyboard:a{enter*2}b",
			want:  kbd("a", "{enter}", "{enter}", "b"),
		},
		// Quoted literal {"text"[*N]}.
		{
			name:  "quoted literal basic",
			input: `**input.keyboard:{"hello"}`,
			want:  kbd("h", "e", "l", "l", "o"),
		},
		{
			name:  "quoted literal with repeat",
			input: `**input.keyboard:{"hi"*3}`,
			want:  kbd("h", "i", "h", "i", "h", "i"),
		},
		{
			name:  "quoted literal empty produces no tokens",
			input: `**input.keyboard:{""}`,
			want:  kbd(),
		},
		{
			name:  "quoted literal asterisk is literal char",
			input: `**input.keyboard:{"a*b"}`,
			want:  kbd("a", "*", "b"),
		},
		{
			name:  "quoted literal plus is literal char",
			input: `**input.keyboard:{"a+b"}`,
			want:  kbd("a", "+", "b"),
		},
		{
			name:  "quoted literal escaped quote",
			input: `**input.keyboard:{"say \"hi\""}`,
			want:  kbd("s", "a", "y", " ", `"`, "h", "i", `"`),
		},
		{
			name:  "asterisk typed via quoted literal",
			input: `**input.keyboard:{"*"*5}`,
			want:  kbd("*", "*", "*", "*", "*"),
		},
		// text: verb — equivalent to quoted literal.
		{
			name:  "text verb basic",
			input: "**input.keyboard:{text:hi}",
			want:  kbd("h", "i"),
		},
		{
			name:  "text verb with repeat",
			input: "**input.keyboard:{text:ab*3}",
			want:  kbd("a", "b", "a", "b", "a", "b"),
		},
		{
			name:  "text verb asterisk followed by non-integer is literal",
			input: "**input.keyboard:{text:a*b}",
			want:  kbd("a", "*", "b"),
		},
		{
			name:  "quoted and text verb are equivalent",
			input: `**input.keyboard:{"hello"*2}`,
			want:  kbd("h", "e", "l", "l", "o", "h", "e", "l", "l", "o"),
		},
		// delay pass-through.
		{
			name:  "delay integer ms passes through",
			input: "**input.keyboard:{delay:500}",
			want:  kbd("{delay:500}"),
		},
		{
			name:  "delay human duration passes through",
			input: "**input.keyboard:{delay:1s}",
			want:  kbd("{delay:1s}"),
		},
		{
			name:  "delay in sequence",
			input: "**input.keyboard:a{delay:100}b",
			want:  kbd("a", "{delay:100}", "b"),
		},
		// Hold verbs and sigils.
		{
			name:  "press verb passes through",
			input: "**input.keyboard:{press:a}",
			want:  kbd("{press:a}"),
		},
		{
			name:  "release verb passes through",
			input: "**input.keyboard:{release:a}",
			want:  kbd("{release:a}"),
		},
		{
			name:  "hold verb passes through",
			input: "**input.keyboard:{hold:a:500}",
			want:  kbd("{hold:a:500}"),
		},
		{
			name:  "press sigil passes through",
			input: "**input.keyboard:{_a}",
			want:  kbd("{_a}"),
		},
		{
			name:  "release sigil passes through",
			input: "**input.keyboard:{^a}",
			want:  kbd("{^a}"),
		},
		{
			name:  "hold sigil passes through",
			input: "**input.keyboard:{~a:500}",
			want:  kbd("{~a:500}"),
		},
		{
			name:  "hold-while sequence",
			input: "**input.keyboard:{_right}bbb{^right}",
			want:  kbd("{_right}", "b", "b", "b", "{^right}"),
		},
		// Command chaining: macro ends at ||.
		{
			name:  "command terminator ends macro",
			input: "**input.keyboard:ab||**stop",
			want: zapscript.Script{Cmds: []zapscript.Command{
				{Name: "input.keyboard", Args: []string{"a", "b"}},
				{Name: "stop"},
			}},
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
			if diff := cmp.Diff(tt.want, got, diffOpts); diff != "" {
				t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// ─── Safety caps ─────────────────────────────────────────────────────────────

func TestInputMacroCaps(t *testing.T) {
	t.Parallel()
	// wantErr: specific exported error checked with errors.Is.
	// wantAnyErr: true when an error is expected but is an unexported fmt.Errorf.
	tests := []struct {
		wantErr    error
		name       string
		input      string
		wantAnyErr bool
	}{
		// Per-repeat cap.
		{
			name:    "key repeat exceeds cap",
			input:   "**input.keyboard:{a*1001}",
			wantErr: zapscript.ErrInputMacroRepeatTooLarge,
		},
		{
			name:    "quoted literal repeat exceeds cap",
			input:   `**input.keyboard:{"a"*1001}`,
			wantErr: zapscript.ErrInputMacroRepeatTooLarge,
		},
		// Quoted literal parse errors.
		{
			name:    "single opening quote only",
			input:   `**input.keyboard:{"}`,
			wantErr: zapscript.ErrUnmatchedQuote,
		},
		{
			name:       "trailing non-asterisk after closing quote",
			input:      `**input.keyboard:{"hello"x}`,
			wantAnyErr: true,
		},
		{
			name:       "zero repeat in quoted literal",
			input:      `**input.keyboard:{"hello"*0}`,
			wantAnyErr: true,
		},
		{
			name:       "non-numeric repeat in quoted literal",
			input:      `**input.keyboard:{"hello"*abc}`,
			wantAnyErr: true,
		},
		// Total keys cap.
		{
			name:    "total keys exceeded via key repeats",
			input:   "**input.keyboard:" + strings.Repeat("{a*1000}", 6),
			wantErr: zapscript.ErrInputMacroTooLong,
		},
		{
			name:    "total keys exceeded via bare chars",
			input:   "**input.keyboard:" + strings.Repeat("a", 5001),
			wantErr: zapscript.ErrInputMacroTooLong,
		},
		{
			name:    "total keys exceeded via pass-through token after full cap",
			input:   "**input.keyboard:" + strings.Repeat("{a*1000}", 5) + "{delay:1}",
			wantErr: zapscript.ErrInputMacroTooLong,
		},
		{
			name:    "total keys exceeded via sigil token after full cap",
			input:   "**input.keyboard:" + strings.Repeat("{a*1000}", 5) + "{_shift}",
			wantErr: zapscript.ErrInputMacroTooLong,
		},
		{
			name:    "total keys exceeded via literal expansion after full cap",
			input:   "**input.keyboard:" + strings.Repeat("{a*1000}", 5) + `{"b"}`,
			wantErr: zapscript.ErrInputMacroTooLong,
		},
		// Other error cases.
		{
			name:    "empty braces",
			input:   "**input.keyboard:{}",
			wantErr: zapscript.ErrUnmatchedInputMacroExt,
		},
		{
			name:    "empty key after repeat suffix removal",
			input:   "**input.keyboard:{*5}",
			wantErr: zapscript.ErrInputMacroEmptyKey,
		},
		{
			name:    "unclosed quoted literal at EOF",
			input:   `**input.keyboard:{"unclosed`,
			wantErr: zapscript.ErrUnmatchedInputMacroExt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			_, err := p.ParseScript()
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseScript() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantAnyErr && err == nil {
				t.Error("ParseScript() expected an error, got nil")
			}
		})
	}
}

// TestInputMacro_RepeatAtMax verifies that exactly 1000 tokens are produced at
// the cap boundary (not an error).
func TestInputMacro_RepeatAtMax(t *testing.T) {
	t.Parallel()
	p := zapscript.NewParser("**input.keyboard:{a*1000}")
	script, err := p.ParseScript()
	if err != nil {
		t.Fatalf("ParseScript() unexpected error: %v", err)
	}
	if got := len(script.Cmds[0].Args); got != 1000 {
		t.Errorf("len(Args) = %d, want 1000", got)
	}
}

// TestInputMacro_TotalKeysAtMax verifies that exactly 5000 tokens are accepted
// without error.
func TestInputMacro_TotalKeysAtMax(t *testing.T) {
	t.Parallel()
	p := zapscript.NewParser("**input.keyboard:" + strings.Repeat("{a*1000}", 5))
	script, err := p.ParseScript()
	if err != nil {
		t.Fatalf("ParseScript() unexpected error: %v", err)
	}
	if got := len(script.Cmds[0].Args); got != 5000 {
		t.Errorf("len(Args) = %d, want 5000", got)
	}
}

// ─── input.text raw mode ─────────────────────────────────────────────────────

func TestInputTextGrammar(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		{
			name:  "bare text produces individual char args",
			input: "**input.text:hello",
			want:  txt("h", "e", "l", "l", "o"),
		},
		{
			name:  "braces are literal chars",
			input: "**input.text:{enter}",
			want:  txt("{", "e", "n", "t", "e", "r", "}"),
		},
		{
			name:  "asterisk is a literal char",
			input: "**input.text:a*5",
			want:  txt("a", "*", "5"),
		},
		{
			name:  "question mark is literal — no adv-arg parsing",
			input: "**input.text:what?",
			want:  txt("w", "h", "a", "t", "?"),
		},
		{
			name:  "full URL is literal",
			input: "**input.text:https://x.com?q=foo",
			want: func() zapscript.Script {
				chars := make([]string, 0, len("https://x.com?q=foo"))
				for _, r := range "https://x.com?q=foo" {
					chars = append(chars, string(r))
				}
				return txt(chars...)
			}(),
		},
		{
			name:  "newline maps to {enter}",
			input: "**input.text:a\nb",
			want:  txt("a", "{enter}", "b"),
		},
		{
			name:  "tab maps to {tab}",
			input: "**input.text:a\tb",
			want:  txt("a", "{tab}", "b"),
		},
		{
			name:  "empty arg produces no tokens",
			input: "**input.text:",
			want:  txt(),
		},
		{
			name:  "speed arg is typed literally",
			input: "**input.text:hello?speed=50",
			want: func() zapscript.Script {
				chars := make([]string, 0, len("hello?speed=50"))
				for _, r := range "hello?speed=50" {
					chars = append(chars, string(r))
				}
				return txt(chars...)
			}(),
		},
		{
			name:  "command terminator ends raw text",
			input: "**input.text:hello||**stop",
			want: zapscript.Script{Cmds: []zapscript.Command{
				{Name: "input.text", Args: []string{"h", "e", "l", "l", "o"}},
				{Name: "stop"},
			}},
		},
		{
			name:    "total keys cap enforced in raw mode",
			input:   "**input.text:" + strings.Repeat("a", 5001),
			wantErr: zapscript.ErrInputMacroTooLong,
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
			if diff := cmp.Diff(tt.want, got, diffOpts); diff != "" {
				t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
