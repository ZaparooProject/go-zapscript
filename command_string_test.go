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
	"testing"

	"github.com/ZaparooProject/go-zapscript"
	"github.com/google/go-cmp/cmp"
)

func TestCommandString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
		cmd  zapscript.Command
	}{
		{
			name: "name only",
			cmd:  zapscript.Command{Name: "stop"},
			want: "**stop",
		},
		{
			name: "single arg",
			cmd:  zapscript.Command{Name: "launch", Args: []string{"/games/snes/mario.sfc"}},
			want: "**launch:/games/snes/mario.sfc",
		},
		{
			name: "multiple args",
			cmd:  zapscript.Command{Name: "greet", Args: []string{"hi", "there"}},
			want: "**greet:hi,there",
		},
		{
			name: "arg with comma needs quoting",
			cmd:  zapscript.Command{Name: "say", Args: []string{"hello, world"}},
			want: `**say:"hello, world"`,
		},
		{
			name: "arg with colon needs quoting",
			cmd:  zapscript.Command{Name: "say", Args: []string{"key:value"}},
			want: `**say:"key:value"`,
		},
		{
			name: "arg with newline re-escapes",
			cmd:  zapscript.Command{Name: "echo", Args: []string{"hello\nworld"}},
			want: `**echo:"hello^nworld"`,
		},
		{
			name: "arg with tab re-escapes",
			cmd:  zapscript.Command{Name: "echo", Args: []string{"one\ttwo"}},
			want: `**echo:"one^ttwo"`,
		},
		{
			name: "arg with caret re-escapes",
			cmd:  zapscript.Command{Name: "echo", Args: []string{"2^3"}},
			want: `**echo:"2^^3"`,
		},
		{
			name: "with advanced args",
			cmd: zapscript.Command{
				Name:    "launch",
				Args:    []string{"game.exe"},
				AdvArgs: zapscript.NewAdvArgs(map[string]string{"platform": "win"}),
			},
			want: "**launch:game.exe?platform=win",
		},
		{
			name: "advanced args sorted",
			cmd: zapscript.Command{
				Name:    "launch",
				Args:    []string{"game.exe"},
				AdvArgs: zapscript.NewAdvArgs(map[string]string{"platform": "win", "fullscreen": "yes", "lang": "en"}),
			},
			want: "**launch:game.exe?fullscreen=yes&lang=en&platform=win",
		},
		{
			name: "advanced args only",
			cmd: zapscript.Command{
				Name:    "example",
				AdvArgs: zapscript.NewAdvArgs(map[string]string{"debug": "true"}),
			},
			want: "**example?debug=true",
		},
		{
			name: "input.keyboard macro",
			cmd:  zapscript.Command{Name: "input.keyboard", Args: []string{"a", "b", "c"}},
			want: "**input.keyboard:abc",
		},
		{
			name: "input.keyboard with extensions",
			cmd:  zapscript.Command{Name: "input.keyboard", Args: []string{"{f1}", "a", "{ctrl+q}"}},
			want: "**input.keyboard:{f1}a{ctrl+q}",
		},
		{
			name: "input.gamepad macro",
			cmd:  zapscript.Command{Name: "input.gamepad", Args: []string{"^", "^", "V", "V", "<", ">"}},
			want: "**input.gamepad:^^VV<>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.cmd.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Command.String() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCommandString_RoundTrip(t *testing.T) {
	t.Parallel()

	// Commands that should round-trip: parse → String() → parse → same Command
	inputs := []string{
		"**stop",
		"**launch:/games/snes/mario.sfc",
		"**greet:hi,there",
		`**say:"hello, world"`,
		"**launch:game.exe?platform=win",
		"**input.keyboard:abc{f1}{enter}",
		"**input.gamepad:^^VV<><>",
		"**delay:500",
		"**launch.random:SNES",
		"**http.get:https://example.com/api",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			// Parse the original input
			reader1 := zapscript.NewParser(input)
			script1, err := reader1.ParseScript()
			if err != nil {
				t.Fatalf("first parse failed: %v", err)
			}
			if len(script1.Cmds) != 1 {
				t.Fatalf("expected 1 command, got %d", len(script1.Cmds))
			}

			// Convert to string
			str := script1.Cmds[0].String()

			// Parse the string output
			reader2 := zapscript.NewParser(str)
			script2, err := reader2.ParseScript()
			if err != nil {
				t.Fatalf("second parse of %q failed: %v", str, err)
			}
			if len(script2.Cmds) != 1 {
				t.Fatalf("expected 1 command from re-parse, got %d", len(script2.Cmds))
			}

			cmd1 := script1.Cmds[0]
			cmd2 := script2.Cmds[0]

			// Compare Name and Args
			if diff := cmp.Diff(cmd1.Name, cmd2.Name); diff != "" {
				t.Errorf("round-trip Name mismatch (-original +reparsed):\n  input: %s\n  string(): %s\n%s",
					input, str, diff)
			}
			if diff := cmp.Diff(cmd1.Args, cmd2.Args); diff != "" {
				t.Errorf("round-trip Args mismatch (-original +reparsed):\n  input: %s\n  string(): %s\n%s",
					input, str, diff)
			}

			// Compare AdvArgs via public Raw() accessor
			if diff := cmp.Diff(cmd1.AdvArgs.Raw(), cmd2.AdvArgs.Raw()); diff != "" {
				t.Errorf("round-trip AdvArgs mismatch (-original +reparsed):\n  input: %s\n  string(): %s\n%s",
					input, str, diff)
			}
		})
	}
}
