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

func TestParseMediaTitleSyntax(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		// Basic format tests
		{
			name:  "basic media title",
			input: `@snes/Super Mario World`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Super Mario World"}},
				},
			},
		},
		{
			name:  "with system name containing spaces",
			input: `@Sega Genesis/Sonic the Hedgehog`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"Sega Genesis/Sonic the Hedgehog"}},
				},
			},
		},
		{
			name:  "with special characters in title",
			input: `@arcade/Ms. Pac-Man`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"arcade/Ms. Pac-Man"}},
				},
			},
		},
		{
			name:  "with ampersand in title",
			input: `@genesis/Sonic & Knuckles`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"genesis/Sonic & Knuckles"}},
				},
			},
		},
		{
			name:  "with multiple slashes in title",
			input: `@ps1/WCW/nWo Thunder`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"ps1/WCW/nWo Thunder"}},
				},
			},
		},

		// Parentheses (filename metadata for tag extraction)
		{
			name:  "with single parenthesis group",
			input: `@snes/Super Mario World (USA)`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Super Mario World (USA)"}},
				},
			},
		},
		{
			name:  "with multiple parenthesis groups",
			input: `@snes/Super Mario World (USA) (Rev 1)`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Super Mario World (USA) (Rev 1)"}},
				},
			},
		},
		{
			name:  "with canonical tag in parenthesis",
			input: `@snes/Game (year:1994)`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Game (year:1994)"}},
				},
			},
		},
		{
			name:  "with canonical tags in multiple parenthesis groups",
			input: `@snes/Game (region:us) (year:1994) (lang:en)`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Game (region:us) (year:1994) (lang:en)"}},
				},
			},
		},
		{
			name:  "with mixed filename and canonical tags",
			input: `@snes/Super Mario World (USA) (year:1991) (Rev A)`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Super Mario World (USA) (year:1991) (Rev A)"}},
				},
			},
		},
		{
			name:  "with tag operators in parentheses",
			input: `@snes/Game (-unfinished:beta) (+region:us)`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Game (-unfinished:beta) (+region:us)"}},
				},
			},
		},

		// Advanced args
		{
			name:  "with single advanced arg",
			input: `@snes/Super Mario World?launcher=custom`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "launch.title",
						Args:    []string{"snes/Super Mario World"},
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"launcher": "custom"}),
					},
				},
			},
		},
		{
			name:  "with multiple advanced args",
			input: `@snes/Game?launcher=custom&tags=region:us`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name: "launch.title",
						Args: []string{"snes/Game"},
						AdvArgs: zapscript.NewAdvArgs(map[string]string{
							"launcher": "custom",
							"tags":     "region:us",
						}),
					},
				},
			},
		},
		{
			name:  "with parentheses and advanced args",
			input: `@snes/Game (USA) (year:1994)?launcher=custom`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name:    "launch.title",
						Args:    []string{"snes/Game (USA) (year:1994)"},
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"launcher": "custom"}),
					},
				},
			},
		},

		// Escape sequences
		{
			name:  "escaped slash in title",
			input: `@snes/Game^/Name`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Game/Name"}},
				},
			},
		},
		{
			name:  "escaped space",
			input: `@snes/Super^ Mario`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Super Mario"}},
				},
			},
		},
		{
			name:  "escaped question mark",
			input: `@snes/What^?`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/What?"}},
				},
			},
		},
		{
			name:  "escaped parenthesis",
			input: `@snes/Game^(2^)`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Game(2)"}},
				},
			},
		},

		// Command chaining
		{
			name:  "chained with delay command",
			input: `@snes/Super Mario World||**delay:1000`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Super Mario World"}},
					{Name: "delay", Args: []string{"1000"}},
				},
			},
		},
		{
			name:  "chained with parentheses and command",
			input: `@snes/Game (USA)||**delay:500`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Game (USA)"}},
					{Name: "delay", Args: []string{"500"}},
				},
			},
		},

		// Whitespace handling
		{
			name:  "trailing space trimmed",
			input: `@snes/Game Name  `,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Game Name"}},
				},
			},
		},
		{
			name:  "leading space after slash",
			input: `@snes/ Game Name`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/ Game Name"}},
				},
			},
		},
		{
			name:  "spaces in parentheses preserved",
			input: `@snes/Game ( USA ) ( Rev 1 )`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Game ( USA ) ( Rev 1 )"}},
				},
			},
		},

		// Invalid format (fallback to auto-launch)
		{
			name:  "no slash separator - fallback to auto-launch",
			input: `@SomeFile`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{"@SomeFile"}},
				},
			},
		},
		{
			name:  "empty after @ - fallback to auto-launch",
			input: `@`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{"@"}},
				},
			},
		},
		{
			name:  "only system no slash - fallback to auto-launch",
			input: `@snes`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{"@snes"}},
				},
			},
		},
		{
			name:  "with parentheses but no slash - fallback to auto-launch",
			input: `@File (USA)`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{"@File (USA)"}},
				},
			},
		},

		// Edge cases - empty system or game name falls back to auto-launch
		{
			name:  "empty system ID - fallback to auto-launch",
			input: `@/Game Name`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{"@/Game Name"}},
				},
			},
		},
		{
			name:  "empty game name - fallback to auto-launch",
			input: `@snes/`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{"@snes/"}},
				},
			},
		},
		{
			name:  "just slash - fallback to auto-launch",
			input: `@/`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{"@/"}},
				},
			},
		},
		{
			name:  "multiple consecutive slashes",
			input: `@snes///Game`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes///Game"}},
				},
			},
		},
		{
			name:  "unicode characters in title",
			input: `@sfc/ドラゴンクエストVII`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"sfc/ドラゴンクエストVII"}},
				},
			},
		},
		{
			name:  "unicode in system and title",
			input: `@スーパーファミコン/ゼルダの伝説`, //nolint:gosmopolitan // Japanese test
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"スーパーファミコン/ゼルダの伝説"}}, //nolint:gosmopolitan // Japanese test
				},
			},
		},
		{
			name:  "numbers in system ID",
			input: `@3do/Road Rash`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"3do/Road Rash"}},
				},
			},
		},
		{
			name:  "hyphens in system ID",
			input: `@sega-cd/Sonic CD`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"sega-cd/Sonic CD"}},
				},
			},
		},

		// Complex real-world examples
		{
			name:  "complex with everything",
			input: `@Sega Genesis/Sonic & Knuckles (USA) (Rev A) (year:1994)?launcher=custom&tags=region:us`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name: "launch.title",
						Args: []string{"Sega Genesis/Sonic & Knuckles (USA) (Rev A) (year:1994)"},
						AdvArgs: zapscript.NewAdvArgs(map[string]string{
							"launcher": "custom",
							"tags":     "region:us",
						}),
					},
				},
			},
		},
		{
			name:  "long title with multiple metadata groups",
			input: `@ps1/Final Fantasy VII (USA) (Disc 1) (Rev 1) (year:1997) (lang:en)`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name: "launch.title",
						Args: []string{"ps1/Final Fantasy VII (USA) (Disc 1) (Rev 1) (year:1997) (lang:en)"},
					},
				},
			},
		},
		{
			name:  "with nested parentheses in title",
			input: `@snes/Game (Prototype (Beta))`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch.title", Args: []string{"snes/Game (Prototype (Beta))"}},
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
