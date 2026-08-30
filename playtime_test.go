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
	"reflect"
	"testing"

	"github.com/ZaparooProject/go-zapscript"
	"github.com/google/go-cmp/cmp"
)

// TestParsePlaytimeExtend covers both amount forms the command accepts, plus
// the argument combinations a written card is likely to carry.
func TestParsePlaytimeExtend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  zapscript.Script
	}{
		{
			name:  "duration amount",
			input: `**playtime.extend:15m?profile=abc123`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    zapscript.ZapScriptCmdPlaytimeExtend,
						Args:    []string{"15m"},
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"profile": "abc123"}),
					},
				},
			},
		},
		{
			name:  "compound duration amount",
			input: `**playtime.extend:1h30m?profile=abc123`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    zapscript.ZapScriptCmdPlaytimeExtend,
						Args:    []string{"1h30m"},
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"profile": "abc123"}),
					},
				},
			},
		},
		{
			name:  "today amount",
			input: `**playtime.extend:today?profile=abc123`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    zapscript.ZapScriptCmdPlaytimeExtend,
						Args:    []string{zapscript.PlaytimeExtendToday},
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"profile": "abc123"}),
					},
				},
			},
		},
		{
			// The when argument is global, so it has to survive alongside a
			// command-specific one.
			name:  "with global when argument",
			input: `**playtime.extend:15m?profile=abc123&when=true`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name: zapscript.ZapScriptCmdPlaytimeExtend,
						Args: []string{"15m"},
						AdvArgs: zapscript.NewAdvArgs(map[string]string{
							"profile": "abc123",
							"when":    "true",
						}),
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
			if err != nil {
				t.Fatalf("ParseScript() error = %v", err)
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestPlaytimeExtendRoundTrip checks a parsed command survives String() and
// reparsing unchanged, so tooling that rewrites a card cannot alter the
// amount or drop the authorizing profile.
func TestPlaytimeExtendRoundTrip(t *testing.T) {
	t.Parallel()

	inputs := []string{
		`**playtime.extend:15m?profile=abc123`,
		`**playtime.extend:1h30m?profile=abc123`,
		`**playtime.extend:today?profile=abc123`,
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			first, err := zapscript.NewParser(input).ParseScript()
			if err != nil {
				t.Fatalf("ParseScript() error = %v", err)
			}
			if len(first.Cmds) != 1 {
				t.Fatalf("expected 1 command, got %d", len(first.Cmds))
			}

			second, err := zapscript.NewParser(first.Cmds[0].String()).ParseScript()
			if err != nil {
				t.Fatalf("reparse error = %v", err)
			}

			if diff := cmp.Diff(first, second, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("round-trip mismatch (-original +reparsed):\n  string(): %s\n%s",
					first.Cmds[0].String(), diff)
			}
		})
	}
}

// TestPlaytimeExtendArgsContract pins the parts of PlaytimeExtendArgs that
// consumers bind against. This library only declares the tags; the decoding
// lives in the consumer, so a tag that stops matching its key constant would
// otherwise fail silently and far from here.
func TestPlaytimeExtendArgsContract(t *testing.T) {
	t.Parallel()

	argsType := reflect.TypeOf(zapscript.PlaytimeExtendArgs{})

	profile, ok := argsType.FieldByName("Profile")
	if !ok {
		t.Fatal("PlaytimeExtendArgs.Profile is missing")
	}
	if got, want := profile.Tag.Get("advarg"), string(zapscript.KeyProfile); got != want {
		t.Errorf("Profile advarg tag = %q, want %q", got, want)
	}

	// GlobalArgs has to stay embedded or the global when argument silently
	// stops reaching the command.
	if _, ok := argsType.FieldByName("When"); !ok {
		t.Error("PlaytimeExtendArgs does not embed GlobalArgs")
	}
}
