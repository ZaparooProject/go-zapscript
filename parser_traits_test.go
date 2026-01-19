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

func TestParseTraitsShorthand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr    error
		wantTraits map[string]any
		name       string
		input      string
		wantCmds   int
	}{
		// Type inference - integers
		{
			name:       "integer value",
			input:      "#level=5",
			wantTraits: map[string]any{"level": int64(5)},
			wantCmds:   0,
		},
		{
			name:       "negative integer",
			input:      "#offset=-10",
			wantTraits: map[string]any{"offset": int64(-10)},
			wantCmds:   0,
		},

		// Type inference - floats
		{
			name:       "float value",
			input:      "#score=99.5",
			wantTraits: map[string]any{"score": float64(99.5)},
			wantCmds:   0,
		},
		{
			name:       "negative float",
			input:      "#temp=-0.5",
			wantTraits: map[string]any{"temp": float64(-0.5)},
			wantCmds:   0,
		},

		// Type inference - booleans
		{
			name:       "boolean true",
			input:      "#active=true",
			wantTraits: map[string]any{"active": true},
			wantCmds:   0,
		},
		{
			name:       "boolean false",
			input:      "#enabled=false",
			wantTraits: map[string]any{"enabled": false},
			wantCmds:   0,
		},
		{
			name:       "boolean shorthand",
			input:      "#active",
			wantTraits: map[string]any{"active": true},
			wantCmds:   0,
		},

		// Type inference - strings
		{
			name:       "string value",
			input:      "#name=mario",
			wantTraits: map[string]any{"name": "mario"},
			wantCmds:   0,
		},
		{
			name:       "force string with quotes",
			input:      `#id="123"`,
			wantTraits: map[string]any{"id": "123"},
			wantCmds:   0,
		},
		{
			name:       "empty value",
			input:      "#key=",
			wantTraits: map[string]any{"key": ""},
			wantCmds:   0,
		},

		// Multiple traits
		{
			name:       "multiple traits",
			input:      "#a=1 #b=2 #c=3",
			wantTraits: map[string]any{"a": int64(1), "b": int64(2), "c": int64(3)},
			wantCmds:   0,
		},
		{
			name:       "multiple boolean shorthands",
			input:      "#x #y #z",
			wantTraits: map[string]any{"x": true, "y": true, "z": true},
			wantCmds:   0,
		},
		{
			name:  "mixed trait types",
			input: "#character=mario #level=5 #active",
			wantTraits: map[string]any{
				"character": "mario",
				"level":     int64(5),
				"active":    true,
			},
			wantCmds: 0,
		},

		// Key normalization
		{
			name:       "key lowercased",
			input:      "#Character=Mario",
			wantTraits: map[string]any{"character": "Mario"},
			wantCmds:   0,
		},
		{
			name:       "uppercase key",
			input:      "#NAME=test",
			wantTraits: map[string]any{"name": "test"},
			wantCmds:   0,
		},

		// Quoted values
		{
			name:       "quoted value with space",
			input:      `#msg="hello world"`,
			wantTraits: map[string]any{"msg": "hello world"},
			wantCmds:   0,
		},
		{
			name:       "single quoted value",
			input:      `#expr='a=b'`,
			wantTraits: map[string]any{"expr": "a=b"},
			wantCmds:   0,
		},
		{
			name:       "quoted hash in value",
			input:      `#tag="#hashtag"`,
			wantTraits: map[string]any{"tag": "#hashtag"},
			wantCmds:   0,
		},

		// Escape sequences
		{
			name:       "escape sequence newline",
			input:      "#text=hello^nworld",
			wantTraits: map[string]any{"text": "hello\nworld"},
			wantCmds:   0,
		},
		{
			name:       "escape sequence tab",
			input:      "#text=hello^tworld",
			wantTraits: map[string]any{"text": "hello\tworld"},
			wantCmds:   0,
		},

		// Merging across chain
		{
			name:       "traits merged across chain",
			input:      "#a=1||#b=2",
			wantTraits: map[string]any{"a": int64(1), "b": int64(2)},
			wantCmds:   0,
		},
		{
			name:       "last wins for duplicate keys",
			input:      "#a=1||#a=2",
			wantTraits: map[string]any{"a": int64(2)},
			wantCmds:   0,
		},

		// UTF-8
		{
			name:       "utf8 value",
			input:      "#name=Jose",
			wantTraits: map[string]any{"name": "Jose"},
			wantCmds:   0,
		},

		// Key with underscore
		{
			name:       "key with underscore",
			input:      "#player_name=mario",
			wantTraits: map[string]any{"player_name": "mario"},
			wantCmds:   0,
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
				t.Errorf("ParseScript() got %d commands, want %d", len(got.Cmds), tt.wantCmds)
				return
			}

			if diff := cmp.Diff(tt.wantTraits, got.Traits); diff != "" {
				t.Errorf("traits mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseTraitsInvalidKeyError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid key starts with number",
			input: "#123=value",
		},
		{
			name:  "invalid key with dash",
			input: "#my-trait=x",
		},
		{
			name:  "invalid key with dot",
			input: "#game.rom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			_, err := p.ParseScript()
			if !errors.Is(err, zapscript.ErrInvalidTraitKey) {
				t.Errorf("ParseScript() error = %v, want %v", err, zapscript.ErrInvalidTraitKey)
			}
		})
	}
}

func TestParseTraitsSilentFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantTraits map[string]any
		name       string
		input      string
		wantCmds   []string
	}{
		{
			name:       "invalid trait with other command - silent fallback",
			input:      "@snes/Mario||#my-trait=x",
			wantCmds:   []string{zapscript.ZapScriptCmdLaunchTitle},
			wantTraits: nil,
		},
		{
			name:       "invalid trait number key with other command",
			input:      "**echo:hello||#123=value",
			wantCmds:   []string{"echo"},
			wantTraits: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if err != nil {
				t.Fatalf("ParseScript() unexpected error: %v", err)
			}

			if len(got.Cmds) != len(tt.wantCmds) {
				t.Fatalf("ParseScript() got %d commands, want %d", len(got.Cmds), len(tt.wantCmds))
			}

			for i, wantCmd := range tt.wantCmds {
				if got.Cmds[i].Name != wantCmd {
					t.Errorf("command[%d] = %q, want %q", i, got.Cmds[i].Name, wantCmd)
				}
			}

			if diff := cmp.Diff(tt.wantTraits, got.Traits); diff != "" {
				t.Errorf("traits mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseTraitsWithOtherCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantTraits map[string]any
		name       string
		input      string
		wantCmds   []string
	}{
		{
			name:       "traits with launch.title",
			input:      "@snes/Super Mario||#character=mario",
			wantCmds:   []string{zapscript.ZapScriptCmdLaunchTitle},
			wantTraits: map[string]any{"character": "mario"},
		},
		{
			name:       "traits before command",
			input:      "#level=5||**echo:hello",
			wantCmds:   []string{"echo"},
			wantTraits: map[string]any{"level": int64(5)},
		},
		{
			name:       "traits in middle",
			input:      "**delay:100||#active||**echo:done",
			wantCmds:   []string{"delay", "echo"},
			wantTraits: map[string]any{"active": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if err != nil {
				t.Fatalf("ParseScript() unexpected error: %v", err)
			}

			if len(got.Cmds) != len(tt.wantCmds) {
				t.Fatalf("ParseScript() got %d commands, want %d", len(got.Cmds), len(tt.wantCmds))
			}

			for i, wantCmd := range tt.wantCmds {
				if got.Cmds[i].Name != wantCmd {
					t.Errorf("command[%d] = %q, want %q", i, got.Cmds[i].Name, wantCmd)
				}
			}

			if diff := cmp.Diff(tt.wantTraits, got.Traits); diff != "" {
				t.Errorf("traits mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseTraitsFullSyntax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantTraits map[string]any
		name       string
		input      string
	}{
		{
			name:       "full syntax simple",
			input:      `**traits:{"a":1}`,
			wantTraits: map[string]any{"a": float64(1)},
		},
		{
			name:       "full syntax complex",
			input:      `**traits:{"name":"mario","level":5}`,
			wantTraits: map[string]any{"name": "mario", "level": float64(5)},
		},
		{
			name:       "full syntax nested",
			input:      `**traits:{"data":{"x":1,"y":2}}`,
			wantTraits: map[string]any{"data": map[string]any{"x": float64(1), "y": float64(2)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if err != nil {
				t.Fatalf("ParseScript() unexpected error: %v", err)
			}

			// Full syntax should NOT create a command
			if len(got.Cmds) != 0 {
				t.Fatalf("ParseScript() got %d commands, want 0", len(got.Cmds))
			}

			if diff := cmp.Diff(tt.wantTraits, got.Traits); diff != "" {
				t.Errorf("traits mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseTraitsUnmatchedQuote(t *testing.T) {
	t.Parallel()

	p := zapscript.NewParser(`#msg="unterminated`)
	_, err := p.ParseScript()
	if !errors.Is(err, zapscript.ErrUnmatchedQuote) {
		t.Errorf("ParseScript() error = %v, want %v", err, zapscript.ErrUnmatchedQuote)
	}
}

func TestParseTraitsArray(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr    error
		wantTraits map[string]any
		name       string
		input      string
	}{
		{
			name:       "empty array",
			input:      "#tags=[]",
			wantTraits: map[string]any{"tags": []any{}},
		},
		{
			name:       "single element",
			input:      "#tags=[action]",
			wantTraits: map[string]any{"tags": []any{"action"}},
		},
		{
			name:       "multiple string elements",
			input:      "#tags=[action,rpg,adventure]",
			wantTraits: map[string]any{"tags": []any{"action", "rpg", "adventure"}},
		},
		{
			name:       "integer elements",
			input:      "#nums=[1,2,3]",
			wantTraits: map[string]any{"nums": []any{int64(1), int64(2), int64(3)}},
		},
		{
			name:       "mixed types",
			input:      "#mixed=[1,hello,true]",
			wantTraits: map[string]any{"mixed": []any{int64(1), "hello", true}},
		},
		{
			name:       "quoted elements with comma",
			input:      `#items=["a,b","c"]`,
			wantTraits: map[string]any{"items": []any{"a,b", "c"}},
		},
		{
			name:       "whitespace around elements",
			input:      "#tags=[ action , rpg ]",
			wantTraits: map[string]any{"tags": []any{"action", "rpg"}},
		},
		{
			name:    "unmatched bracket",
			input:   "#tags=[action,rpg",
			wantErr: zapscript.ErrUnmatchedArrayBracket,
		},
		// Additional coverage tests
		{
			name:       "single quoted elements",
			input:      `#items=['hello','world']`,
			wantTraits: map[string]any{"items": []any{"hello", "world"}},
		},
		{
			name:       "quoted element with escape newline",
			input:      `#items=["hello^nworld"]`,
			wantTraits: map[string]any{"items": []any{"hello\nworld"}},
		},
		{
			name:       "quoted element with escape tab",
			input:      `#items=["hello^tworld"]`,
			wantTraits: map[string]any{"items": []any{"hello\tworld"}},
		},
		{
			name:       "unquoted element with escape",
			input:      `#items=[hello^nworld]`,
			wantTraits: map[string]any{"items": []any{"hello\nworld"}},
		},
		{
			name:       "quoted element with escape caret",
			input:      `#items=["hello^^world"]`,
			wantTraits: map[string]any{"items": []any{"hello^world"}},
		},
		{
			name:       "float elements",
			input:      "#nums=[1.5,2.5]",
			wantTraits: map[string]any{"nums": []any{float64(1.5), float64(2.5)}},
		},
		{
			name:       "boolean elements",
			input:      "#flags=[true,false]",
			wantTraits: map[string]any{"flags": []any{true, false}},
		},
		{
			name:       "negative numbers",
			input:      "#nums=[-1,-2.5]",
			wantTraits: map[string]any{"nums": []any{int64(-1), float64(-2.5)}},
		},
		{
			name:    "unmatched quote in array element",
			input:   `#items=["unterminated]`,
			wantErr: zapscript.ErrUnmatchedQuote,
		},
		{
			name:       "empty string element",
			input:      `#items=[""]`,
			wantTraits: map[string]any{"items": []any{""}},
		},
		{
			name:       "array with trailing comma style whitespace",
			input:      "#tags=[a, b, c]",
			wantTraits: map[string]any{"tags": []any{"a", "b", "c"}},
		},
		{
			name:       "single quoted element with escape",
			input:      `#items=['hello^nworld']`,
			wantTraits: map[string]any{"items": []any{"hello\nworld"}},
		},
		{
			name:       "array element with escape return",
			input:      `#items=[hello^rworld]`,
			wantTraits: map[string]any{"items": []any{"hello\rworld"}},
		},
		{
			name:       "single quoted escape caret",
			input:      `#items=['hello^^world']`,
			wantTraits: map[string]any{"items": []any{"hello^world"}},
		},
		{
			name:       "unquoted element with caret escape",
			input:      `#items=[hello^^world]`,
			wantTraits: map[string]any{"items": []any{"hello^world"}},
		},
		{
			name:       "quoted element with tab escape",
			input:      `#items=["hello^tworld"]`,
			wantTraits: map[string]any{"items": []any{"hello\tworld"}},
		},
		{
			name:       "single quoted tab escape",
			input:      `#items=['a^tb']`,
			wantTraits: map[string]any{"items": []any{"a\tb"}},
		},
		{
			name:       "multiple escaped chars in quoted element",
			input:      `#items=["a^nb^tc"]`,
			wantTraits: map[string]any{"items": []any{"a\nb\tc"}},
		},
		{
			name:       "unquoted multiple escapes",
			input:      `#items=[a^nb^tc]`,
			wantTraits: map[string]any{"items": []any{"a\nb\tc"}},
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

func TestParseTraitsNotInCmds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "shorthand traits only",
			input: "#key=value",
		},
		{
			name:  "full syntax traits only",
			input: `**traits:{"key":"value"}`,
		},
		{
			name:  "traits with other commands",
			input: "#key=value||**echo:hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if err != nil {
				t.Fatalf("ParseScript() unexpected error: %v", err)
			}

			// Verify no traits command in Cmds
			for _, cmd := range got.Cmds {
				if cmd.Name == zapscript.ZapScriptCmdTraits {
					t.Error("found traits command in Cmds, should be in Script.Traits")
				}
			}
		})
	}
}

func TestInferType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		want   any
		name   string
		value  string
		quoted bool
	}{
		{int64(5), "integer", "5", false},
		{int64(-10), "negative integer", "-10", false},
		{float64(3.14), "float", "3.14", false},
		{float64(-0.5), "negative float", "-0.5", false},
		{true, "boolean true", "true", false},
		{false, "boolean false", "false", false},
		{"hello", "string", "hello", false},
		{"123", "quoted number", "123", true},
		{"", "empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// We test via the parser since inferType is not exported
			var input string
			if tt.quoted {
				input = `#test="` + tt.value + `"`
			} else {
				input = "#test=" + tt.value
			}
			p := zapscript.NewParser(input)
			got, err := p.ParseScript()
			if err != nil {
				t.Fatalf("ParseScript() unexpected error: %v", err)
			}

			gotValue := got.Traits["test"]
			if diff := cmp.Diff(tt.want, gotValue); diff != "" {
				t.Errorf("value mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseTraitsValueEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr    error
		wantTraits map[string]any
		name       string
		input      string
	}{
		// Single quoted values
		{
			name:       "single quoted value",
			input:      "#key='value'",
			wantTraits: map[string]any{"key": "value"},
		},
		{
			name:       "single quoted with double quote inside",
			input:      `#key='say "hello"'`,
			wantTraits: map[string]any{"key": `say "hello"`},
		},
		// Escape sequences in quoted values
		{
			name:       "quoted escape carriage return",
			input:      `#key="hello^rworld"`,
			wantTraits: map[string]any{"key": "hello\rworld"},
		},
		{
			name:       "quoted escape caret",
			input:      `#key="hello^^world"`,
			wantTraits: map[string]any{"key": "hello^world"},
		},
		{
			name:       "quoted escape quote",
			input:      `#key="hello^"world"`,
			wantTraits: map[string]any{"key": `hello"world`},
		},
		// Escape sequences in unquoted values
		{
			name:       "unquoted escape carriage return",
			input:      "#key=hello^rworld",
			wantTraits: map[string]any{"key": "hello\rworld"},
		},
		{
			name:       "unquoted escape at end",
			input:      "#key=hello^",
			wantTraits: map[string]any{"key": "hello^"},
		},
		// Value termination
		{
			name:       "value ends at command separator",
			input:      "#key=value||#key2=value2",
			wantTraits: map[string]any{"key": "value", "key2": "value2"},
		},
		// Empty and whitespace
		{
			name:       "trait with only whitespace after",
			input:      "#key=value   ",
			wantTraits: map[string]any{"key": "value"},
		},
		{
			name:       "multiple traits with varied whitespace",
			input:      "#a=1  #b=2   #c=3",
			wantTraits: map[string]any{"a": int64(1), "b": int64(2), "c": int64(3)},
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

func TestParseTraitsSyntaxEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr    error
		wantTraits map[string]any
		name       string
		input      string
		wantCmds   int
	}{
		// Boolean shorthand edge cases
		{
			name:       "boolean shorthand at EOF",
			input:      "#flag",
			wantTraits: map[string]any{"flag": true},
			wantCmds:   0,
		},
		{
			name:       "boolean shorthand before command separator",
			input:      "#flag||**echo:hi",
			wantTraits: map[string]any{"flag": true},
			wantCmds:   1,
		},
		{
			name:       "boolean shorthand with whitespace after",
			input:      "#flag   ",
			wantTraits: map[string]any{"flag": true},
			wantCmds:   0,
		},
		// Key variations
		{
			name:       "key with numbers",
			input:      "#player1=mario",
			wantTraits: map[string]any{"player1": "mario"},
			wantCmds:   0,
		},
		{
			name:       "key all uppercase",
			input:      "#LEVEL=5",
			wantTraits: map[string]any{"level": int64(5)},
			wantCmds:   0,
		},
		{
			name:       "key mixed case",
			input:      "#PlayerName=mario",
			wantTraits: map[string]any{"playername": "mario"},
			wantCmds:   0,
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
				t.Errorf("ParseScript() got %d commands, want %d", len(got.Cmds), tt.wantCmds)
				return
			}

			if diff := cmp.Diff(tt.wantTraits, got.Traits); diff != "" {
				t.Errorf("traits mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseTraitsConsumeToEndOfCmd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr  error
		name     string
		input    string
		wantCmds []string
	}{
		{
			name:     "invalid trait consumed to command separator",
			input:    "#my-invalid-trait=value||**echo:hello",
			wantCmds: []string{"echo"},
		},
		{
			name:     "invalid trait consumed to EOF",
			input:    "**echo:hi||#123invalid",
			wantCmds: []string{"echo"},
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

			if len(got.Cmds) != len(tt.wantCmds) {
				t.Fatalf("ParseScript() got %d commands, want %d", len(got.Cmds), len(tt.wantCmds))
			}

			for i, wantCmd := range tt.wantCmds {
				if got.Cmds[i].Name != wantCmd {
					t.Errorf("command[%d] = %q, want %q", i, got.Cmds[i].Name, wantCmd)
				}
			}
		})
	}
}

func TestParseTraitsEmptyHash(t *testing.T) {
	t.Parallel()

	// Test edge case: # followed by space (invalid trait, should error or fallback)
	p := zapscript.NewParser("# ")
	_, err := p.ParseScript()
	if !errors.Is(err, zapscript.ErrInvalidTraitKey) {
		t.Errorf("ParseScript() error = %v, want %v", err, zapscript.ErrInvalidTraitKey)
	}
}

func TestParseTraitsMixedValidInvalid(t *testing.T) {
	t.Parallel()

	// Valid traits followed by invalid should still capture valid ones
	// and silently ignore invalid when mixed with commands
	tests := []struct {
		wantTraits map[string]any
		name       string
		input      string
		wantCmds   int
	}{
		{
			name:       "valid trait then invalid trait then command",
			input:      "#valid=1||#123invalid||**echo:hi",
			wantTraits: map[string]any{"valid": int64(1)},
			wantCmds:   1,
		},
		{
			name:       "command then valid trait then invalid trait",
			input:      "**echo:hi||#valid=1||#my-bad=x",
			wantTraits: map[string]any{"valid": int64(1)},
			wantCmds:   1,
		},
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
				t.Errorf("ParseScript() got %d commands, want %d", len(got.Cmds), tt.wantCmds)
			}

			if diff := cmp.Diff(tt.wantTraits, got.Traits); diff != "" {
				t.Errorf("traits mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
