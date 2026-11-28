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
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/ZaparooProject/go-zapscript"
	"github.com/google/go-cmp/cmp"
)

func TestParseMemory(t *testing.T) {
	t.Parallel()

	coin := zapscript.Command{Name: zapscript.ZapScriptCmdInputCoinP1}
	launch := zapscript.Command{Name: zapscript.ZapScriptCmdLaunch, Args: []string{"/roms/game.sfc"}}

	tests := []struct {
		name  string
		input string
		want  zapscript.Script
	}{
		{
			name:  "credits and a quoted owner",
			input: `**input.coinp1||$credits=10 $owner="Callan B"`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{coin},
				Memory: zapscript.Memory{
					{Key: "credits", Value: int64(10)},
					{Key: "owner", Value: "Callan B"},
				},
			},
		},
		{
			name:  "every value type",
			input: `/roms/game.sfc||$i=-3 $f=2.5 $b=false $s='it^'s' $l=[1,"two",3.5] $e=`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{launch},
				Memory: zapscript.Memory{
					{Key: "i", Value: int64(-3)},
					{Key: "f", Value: 2.5},
					{Key: "b", Value: false},
					{Key: "s", Value: "it's"},
					{Key: "l", Value: []any{int64(1), "two", 3.5}},
					{Key: "e", Value: ""},
				},
			},
		},
		{
			name:  "pairs without spaces between them",
			input: `/roms/game.sfc||$a=1$b="x"$c=[2]`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{launch},
				Memory: zapscript.Memory{
					{Key: "a", Value: int64(1)},
					{Key: "b", Value: "x"},
					{Key: "c", Value: []any{int64(2)}},
				},
			},
		},
		{
			name:  "keys are lowercased and a repeated key keeps its first place",
			input: `/roms/game.sfc||$Credits=1 $other=2 $CREDITS=3`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{launch},
				Memory: zapscript.Memory{
					{Key: "credits", Value: int64(3)},
					{Key: "other", Value: int64(2)},
				},
			},
		},
		{
			name:  "a trailing separator and whitespace",
			input: "/roms/game.sfc|| $a=1 || \n",
			want: zapscript.Script{
				Cmds:   []zapscript.Command{launch},
				Memory: zapscript.Memory{{Key: "a", Value: int64(1)}},
			},
		},
		{
			name:  "memory on its own",
			input: `$credits=3`,
			want:  zapscript.Script{Memory: zapscript.Memory{{Key: "credits", Value: int64(3)}}},
		},
		{
			name:  "a quoted value may hold separators and hashes",
			input: `/roms/game.sfc||$note="a || b #c $d"`,
			want: zapscript.Script{
				Cmds:   []zapscript.Command{launch},
				Memory: zapscript.Memory{{Key: "note", Value: "a || b #c $d"}},
			},
		},
		{
			name:  "an unquoted value may hold a hash",
			input: `/roms/game.sfc||$tag=a#b`,
			want: zapscript.Script{
				Cmds:   []zapscript.Command{launch},
				Memory: zapscript.Memory{{Key: "tag", Value: "a#b"}},
			},
		},
		{
			name:  "traits and memory together",
			input: `#arcade||/roms/game.sfc||$a=1`,
			want: zapscript.Script{
				Traits: map[string]any{"arcade": true},
				Cmds:   []zapscript.Command{launch},
				Memory: zapscript.Memory{{Key: "a", Value: int64(1)}},
			},
		},
		{
			name:  "a $ not followed by a letter is a launch",
			input: `$1/game.sfc`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{{Name: zapscript.ZapScriptCmdLaunch, Args: []string{"$1/game.sfc"}}},
			},
		},
		{
			name:  "a $ inside a launch path is text",
			input: `/roms/$game.sfc`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{{Name: zapscript.ZapScriptCmdLaunch, Args: []string{"/roms/$game.sfc"}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := zapscript.NewParser(tt.input).ParseScript()
			if err != nil {
				t.Fatalf("ParseScript() error = %v", err)
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseMemoryErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		want  error
		name  string
		input string
	}{
		{name: "a key without a value", input: `**stop||$credits`, want: zapscript.ErrInvalidMemory},
		{name: "a key with a bad character", input: `**stop||$my-key=1`, want: zapscript.ErrInvalidMemory},
		{name: "text after a quoted value", input: `**stop||$a="x"y`, want: zapscript.ErrInvalidMemory},
		{name: "a lone pipe after a value", input: `**stop||$a=1 | $b=2`, want: zapscript.ErrInvalidMemory},
		{name: "a second $ without a key", input: `**stop||$a=1 $`, want: zapscript.ErrInvalidMemory},
		{name: "a second key starting with a digit", input: `**stop||$a=1 $2=b`, want: zapscript.ErrInvalidMemory},
		{name: "a command after memory", input: `$a=1||**stop`, want: zapscript.ErrMemoryNotLast},
		{name: "a second memory item", input: `**stop||$a=1||$b=2`, want: zapscript.ErrMemoryNotLast},
		{name: "an unmatched quote", input: `**stop||$a="x`, want: zapscript.ErrUnmatchedQuote},
		{name: "an unmatched bracket", input: `**stop||$a=[1,2`, want: zapscript.ErrUnmatchedArrayBracket},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := zapscript.NewParser(tt.input).ParseScript()
			if !errors.Is(err, tt.want) {
				t.Errorf("ParseScript(%q) error = %v, want %v", tt.input, err, tt.want)
			}
		})
	}
}

func TestFormatMemory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
		mem  zapscript.Memory
	}{
		{name: "empty", mem: nil, want: ""},
		{
			name: "ints of any width are written the same",
			mem: zapscript.Memory{
				{Key: "a", Value: 1},
				{Key: "b", Value: int8(-2)},
				{Key: "c", Value: uint32(3)},
				{Key: "d", Value: int64(4)},
			},
			want: `$a=1 $b=-2 $c=3 $d=4`,
		},
		{
			name: "whole floats keep a point",
			mem: zapscript.Memory{
				{Key: "a", Value: 1.0},
				{Key: "b", Value: float32(0.5)},
				{Key: "c", Value: 1e21},
				{Key: "d", Value: math.Copysign(0, -1)},
			},
			want: `$a=1.0 $b=0.5 $c=1e+21 $d=-0.0`,
		},
		{
			name: "strings are always quoted and escaped",
			mem: zapscript.Memory{
				{Key: "n", Value: "10"},
				{Key: "b", Value: "true"},
				{Key: "s", Value: "a \"b\" ^ c||d\n\t\r"},
				{Key: "e", Value: ""},
			},
			want: `$n="10" $b="true" $s="a ^"b^" ^^ c^|^|d^n^t^r" $e=""`,
		},
		{
			name: "lists",
			mem: zapscript.Memory{
				{Key: "l", Value: []any{1, "x", true, 2.0}},
				{Key: "s", Value: []string{"a", "b"}},
				{Key: "e", Value: []any{}},
			},
			want: `$l=[1,"x",true,2.0] $s=["a","b"] $e=[]`,
		},
		{
			name: "keys are lowercased",
			mem:  zapscript.Memory{{Key: "Credits", Value: 1}},
			want: `$credits=1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := zapscript.FormatMemory(tt.mem)
			if err != nil {
				t.Fatalf("FormatMemory() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("FormatMemory() = %q, want %q", got, tt.want)
			}
			if got == "" {
				return
			}

			parsed, err := zapscript.NewParser(got).ParseScript()
			if err != nil {
				t.Fatalf("reparse %q error = %v", got, err)
			}
			want := zapscript.Memory{}
			for _, e := range tt.mem {
				if err := want.Set(e.Key, e.Value); err != nil {
					t.Fatalf("Set(%q) error = %v", e.Key, err)
				}
			}
			if diff := cmp.Diff(want, parsed.Memory); diff != "" {
				t.Errorf("round trip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFormatMemoryErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value any
		want  error
		name  string
		key   string
	}{
		{name: "nil", key: "a", value: nil, want: zapscript.ErrInvalidMemoryValue},
		{name: "NaN", key: "a", value: math.NaN(), want: zapscript.ErrInvalidMemoryValue},
		{name: "infinity", key: "a", value: math.Inf(1), want: zapscript.ErrInvalidMemoryValue},
		{name: "a map", key: "a", value: map[string]any{}, want: zapscript.ErrInvalidMemoryValue},
		{name: "a nested list", key: "a", value: []any{[]any{1}}, want: zapscript.ErrInvalidMemoryValue},
		{name: "a uint64 too large", key: "a", value: uint64(math.MaxUint64), want: zapscript.ErrInvalidMemoryValue},
		{name: "invalid UTF-8", key: "a", value: "\xff", want: zapscript.ErrInvalidMemoryValue},
		{name: "a NUL", key: "a", value: "a\x00b", want: zapscript.ErrInvalidMemoryValue},
		{name: "a bad key", key: "1a", value: 1, want: zapscript.ErrInvalidMemory},
		{name: "an empty key", key: "", value: 1, want: zapscript.ErrInvalidMemory},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := zapscript.FormatMemory(zapscript.Memory{{Key: tt.key, Value: tt.value}})
			if !errors.Is(err, tt.want) {
				t.Errorf("FormatMemory() error = %v, want %v", err, tt.want)
			}

			var m zapscript.Memory
			if err := m.Set(tt.key, tt.value); !errors.Is(err, tt.want) {
				t.Errorf("Set() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestReplaceMemory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want string
		mem  zapscript.Memory
	}{
		{
			name: "replaces the item and keeps the script",
			text: `**input.coinp1?when=[[true]]||**echo:  spaced^,text  ||$credits=10 $owner="Callan B"`,
			mem:  zapscript.Memory{{Key: "credits", Value: 9}, {Key: "owner", Value: "Callan B"}},
			want: `**input.coinp1?when=[[true]]||**echo:  spaced^,text  ||$credits=9 $owner="Callan B"`,
		},
		{
			name: "drops a trailing separator after the item",
			text: `**stop||$a=1||`,
			mem:  zapscript.Memory{{Key: "a", Value: 2}},
			want: `**stop||$a=2`,
		},
		{
			name: "removes memory-only text led by whitespace",
			text: " \n$a=1",
			want: "",
		},
		{
			name: "keeps whitespace before the item",
			text: "**stop|| \n$a=1",
			mem:  zapscript.Memory{{Key: "a", Value: 2}},
			want: "**stop|| \n$a=2",
		},
		{
			name: "appends to a script without memory",
			text: `**stop`,
			mem:  zapscript.Memory{{Key: "a", Value: 1}},
			want: `**stop||$a=1`,
		},
		{
			name: "appends after an existing trailing separator",
			text: `**stop||`,
			mem:  zapscript.Memory{{Key: "a", Value: 1}},
			want: `**stop||$a=1`,
		},
		{
			name: "keeps whitespace that ends the script when appending",
			text: `**stop `,
			mem:  zapscript.Memory{{Key: "a", Value: 1}},
			want: `**stop ||$a=1`,
		},
		{
			name: "removes the item when memory is emptied",
			text: `**stop|| $a=1`,
			mem:  nil,
			want: `**stop`,
		},
		{
			name: "keeps whitespace before the separator when removing",
			text: `**stop ||$a=1`,
			mem:  zapscript.Memory{},
			want: `**stop `,
		},
		{
			name: "leaves a script without memory alone when there is none to write",
			text: "**stop  ",
			mem:  nil,
			want: "**stop  ",
		},
		{
			name: "emptying memory-only text leaves nothing",
			text: `$a=1`,
			mem:  nil,
			want: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := zapscript.ReplaceMemory(tt.text, tt.mem)
			if err != nil {
				t.Fatalf("ReplaceMemory() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("ReplaceMemory() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReplaceMemoryErrors(t *testing.T) {
	t.Parallel()

	mem := zapscript.Memory{{Key: "a", Value: 1}}
	tests := []struct {
		want error
		name string
		text string
		mem  zapscript.Memory
	}{
		{
			name: "a nil value", text: `**stop||$a=`,
			mem: zapscript.Memory{{Key: "a"}}, want: zapscript.ErrInvalidMemoryValue,
		},
		{name: "a parse error", text: `**stop||$a="x`, mem: mem, want: zapscript.ErrUnmatchedQuote},
		{name: "a NUL in the text", text: "**stop\x00", mem: mem, want: zapscript.ErrInvalidMemory},
		{name: "a lone pipe at the end", text: `/roms/game|`, mem: mem, want: zapscript.ErrInvalidMemory},
		{
			name: "a value memory can't hold", text: `**stop`,
			mem: zapscript.Memory{{Key: "a", Value: math.NaN()}}, want: zapscript.ErrInvalidMemoryValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := zapscript.ReplaceMemory(tt.text, tt.mem); !errors.Is(err, tt.want) {
				t.Errorf("ReplaceMemory(%q) error = %v, want %v", tt.text, err, tt.want)
			}
		})
	}
}

func TestMemoryMethods(t *testing.T) {
	t.Parallel()

	var m zapscript.Memory
	if err := m.Set("Credits", 3); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := m.Set("owner", "Ann"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := m.Set("CREDITS", 2); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	want := zapscript.Memory{{Key: "credits", Value: int64(2)}, {Key: "owner", Value: "Ann"}}
	if diff := cmp.Diff(want, m); diff != "" {
		t.Errorf("after Set (-want +got):\n%s", diff)
	}

	if v, ok := m.Get("CrEdItS"); !ok || v != int64(2) {
		t.Errorf("Get(credits) = %v, %v", v, ok)
	}
	if _, ok := m.Get("missing"); ok {
		t.Error("Get(missing) found a value")
	}

	if err := m.Set("list", []any{1}); err != nil {
		t.Fatalf("Set(list) error = %v", err)
	}
	clone := m.Clone()
	list, ok := clone[2].Value.([]any)
	if !ok {
		t.Fatalf("clone list is %T", clone[2].Value)
	}
	list[0] = int64(99)
	if v, _ := m.Get("list"); v.([]any)[0] != int64(1) {
		t.Error("Clone shares a list with the original")
	}

	if diff := cmp.Diff(map[string]any{
		"credits": int64(2), "owner": "Ann", "list": []any{int64(1)},
	}, m.Map()); diff != "" {
		t.Errorf("Map() (-want +got):\n%s", diff)
	}

	if !m.Delete("owner") || m.Delete("owner") {
		t.Error("Delete reported the wrong presence")
	}
	if diff := cmp.Diff(zapscript.Memory{
		{Key: "credits", Value: int64(2)}, {Key: "list", Value: []any{int64(1)}},
	}, m); diff != "" {
		t.Errorf("after Delete (-want +got):\n%s", diff)
	}
}

func TestMemoryJSON(t *testing.T) {
	t.Parallel()

	script, err := zapscript.NewParser(`**stop||$z=1 $a="x" $m=[true]`).ParseScript()
	if err != nil {
		t.Fatalf("ParseScript() error = %v", err)
	}
	got, err := json.Marshal(script.Memory)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if want := `{"z":1,"a":"x","m":[true]}`; string(got) != want {
		t.Errorf("Marshal() = %s, want %s", got, want)
	}

	script.Memory = nil
	got, err = json.Marshal(script)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if strings.Contains(string(got), "memory") {
		t.Errorf("Marshal() without memory = %s, want no memory key", got)
	}
}

func TestMemoryExpressions(t *testing.T) {
	t.Parallel()

	env := zapscript.ArgExprEnv{
		Memory: map[string]any{"credits": int64(2), "owner": "Ann"},
		Now:    1_700_000_000,
	}
	tests := []struct {
		input string
		want  string
	}{
		{input: `[[memory.credits > 0]]`, want: "true"},
		{input: `[[memory.owner]]`, want: "Ann"},
		{input: `[[memory.missing ?? 5]]`, want: "5"},
		{input: `[[memory.missing == nil]]`, want: "true"},
		{input: `[[memory.credits]] left`, want: "2 left"},
		{input: `[[memory.credits - 1]]`, want: "1"},
		{input: `[[now]]`, want: "1700000000"},
		{input: `[[now >= 1700000000]]`, want: "true"},
		// The now field must not hide Expr's now() and date() functions.
		{input: `[[now().Unix() > 0]]`, want: "true"},
		{input: `[[date("2024-01-02").Year()]]`, want: "2024"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			parsed, err := zapscript.NewParser(tt.input).ParseExpressions()
			if err != nil {
				t.Fatalf("ParseExpressions() error = %v", err)
			}
			got, err := zapscript.NewParser(parsed).EvalExpressions(env)
			if err != nil {
				t.Fatalf("EvalExpressions() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("EvalExpressions() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseMemoryValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		want  any
		name  string
		input string
	}{
		{name: "int", input: "10", want: int64(10)},
		{name: "negative int", input: "-3", want: int64(-3)},
		{name: "float", input: "1.5", want: 1.5},
		{name: "bool", input: "true", want: true},
		{name: "unquoted text with spaces", input: "Callan B", want: "Callan B"},
		{name: "surrounding space trimmed", input: "  7 ", want: int64(7)},
		{name: "quoted number stays a string", input: `"5"`, want: "5"},
		{name: "quoted with escape", input: `"a^"b"`, want: `a"b`},
		{name: "single quoted", input: `'x y'`, want: "x y"},
		{name: "list", input: "[1,two,3.5]", want: []any{int64(1), "two", 3.5}},
		{name: "empty", input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := zapscript.ParseMemoryValue(tt.input)
			if err != nil {
				t.Fatalf("ParseMemoryValue(%q): %v", tt.input, err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseMemoryValue(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func TestParseMemoryValueErrors(t *testing.T) {
	t.Parallel()

	for _, input := range []string{
		`"unterminated`,
		`"a" b`,
		"[1,2",
		"[1,2] x",
		"[[1]]",
		"a\x00b",
	} {
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			if _, err := zapscript.ParseMemoryValue(input); err == nil {
				t.Errorf("ParseMemoryValue(%q) succeeded, want error", input)
			}
		})
	}
}
