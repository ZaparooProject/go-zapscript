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

	"github.com/ZaparooProject/go-zapscript"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTraitKeyValid(t *testing.T) {
	t.Parallel()

	keys := []string{
		"a",
		"tap",
		"hold",
		"scan_mode",
		"trait1",
		"a_1_b_2",
		"MixedCase",
		"UPPER",
		"x_",
	}

	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			if err := zapscript.ValidateTraitKey(key); err != nil {
				t.Errorf("ValidateTraitKey(%q) = %v, want nil", key, err)
			}
		})
	}
}

func TestValidateTraitKeyInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
	}{
		{name: "empty", key: ""},
		{name: "leading digit", key: "1trait"},
		{name: "leading underscore", key: "_trait"},
		{name: "dash", key: "my-trait"},
		{name: "dot", key: "game.rom"},
		{name: "space", key: "my trait"},
		{name: "non ascii", key: "café"},
		{name: "leading non ascii", key: "état"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := zapscript.ValidateTraitKey(tt.key)
			if err == nil {
				t.Fatalf("ValidateTraitKey(%q) = nil, want error", tt.key)
			}
			if !errors.Is(err, zapscript.ErrInvalidTraitKey) {
				t.Errorf("ValidateTraitKey(%q) error = %v, want %v",
					tt.key, err, zapscript.ErrInvalidTraitKey)
			}
		})
	}
}

// The reported position counts runes from 1, so it points at the character a
// reader would count rather than at a byte offset.
func TestValidateTraitKeyErrorPosition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
		want string
	}{
		{name: "second character", key: "a-b", want: "position 2"},
		{name: "third character", key: "ab-c", want: "position 3"},
		{name: "after multi byte prefix", key: "aé-b", want: "position 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := zapscript.ValidateTraitKey(tt.key)
			if err == nil {
				t.Fatalf("ValidateTraitKey(%q) = nil, want error", tt.key)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("ValidateTraitKey(%q) error = %q, want it to contain %q",
					tt.key, err.Error(), tt.want)
			}
		})
	}
}

// A key the validator accepts must also survive the parser, and one it rejects
// must not become a trait. This is what keeps the two rule sets from drifting.
func TestValidateTraitKeyMatchesParser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		key   string
		valid bool
	}{
		{name: "simple", key: "tap", valid: true},
		{name: "underscore", key: "scan_mode", valid: true},
		{name: "digit suffix", key: "player2", valid: true},
		{name: "dash", key: "my-trait", valid: false},
		{name: "dot", key: "game.rom", valid: false},
		{name: "leading digit", key: "1up", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			validErr := zapscript.ValidateTraitKey(tt.key)
			if (validErr == nil) != tt.valid {
				t.Fatalf("ValidateTraitKey(%q) = %v, want valid=%v", tt.key, validErr, tt.valid)
			}

			script, parseErr := zapscript.NewParser("#" + tt.key).ParseScript()
			if !tt.valid {
				if !errors.Is(parseErr, zapscript.ErrInvalidTraitKey) {
					t.Fatalf("ParseScript(%q) error = %v, want %v",
						"#"+tt.key, parseErr, zapscript.ErrInvalidTraitKey)
				}
				return
			}
			if parseErr != nil {
				t.Fatalf("ParseScript(%q) unexpected error: %v", "#"+tt.key, parseErr)
			}
			if _, ok := script.Traits[zapscript.NormalizeTraitKey(tt.key)]; !ok {
				t.Errorf("ParseScript(%q) traits = %v, want key %q",
					"#"+tt.key, script.Traits, zapscript.NormalizeTraitKey(tt.key))
			}
		})
	}
}

func TestNormalizeTraitKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
		want string
	}{
		{name: "already lower", key: "tap", want: "tap"},
		{name: "upper", key: "TAP", want: "tap"},
		{name: "mixed", key: "ScanMode", want: "scanmode"},
		{name: "underscore preserved", key: "Scan_Mode", want: "scan_mode"},
		{name: "empty", key: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := zapscript.NormalizeTraitKey(tt.key); got != tt.want {
				t.Errorf("NormalizeTraitKey(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

// The parser lowercases keys as it reads them, so a normalized key is what a
// caller must look up in Script.Traits.
func TestNormalizeTraitKeyMatchesParsedKey(t *testing.T) {
	t.Parallel()

	script, err := zapscript.NewParser("#ScanMode=tap").ParseScript()
	if err != nil {
		t.Fatalf("ParseScript() unexpected error: %v", err)
	}
	if _, ok := script.Traits[zapscript.NormalizeTraitKey("ScanMode")]; !ok {
		t.Errorf("traits = %v, want key %q", script.Traits, zapscript.NormalizeTraitKey("ScanMode"))
	}
}

// Both trait spellings must land on the same map key, so a caller looking a
// trait up by a normalized key finds it however the script was written.
func TestParseTraitsJSONKeysNormalized(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantTraits map[string]any
		name       string
		input      string
	}{
		{
			name:       "json key lowercased",
			input:      `**traits:{"Tap":true}`,
			wantTraits: map[string]any{"tap": true},
		},
		{
			name:       "json key with underscore preserved",
			input:      `**traits:{"Scan_Mode":"tap"}`,
			wantTraits: map[string]any{"scan_mode": "tap"},
		},
		{
			name:       "invalid json key dropped",
			input:      `**traits:{"bad-key":1,"good":2}`,
			wantTraits: map[string]any{"good": float64(2)},
		},
		{
			name:       "inline and json agree",
			input:      `#Hold||**traits:{"Tap":false}`,
			wantTraits: map[string]any{"hold": true, "tap": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			script, err := zapscript.NewParser(tt.input).ParseScript()
			if err != nil {
				t.Fatalf("ParseScript(%q) unexpected error: %v", tt.input, err)
			}
			if len(script.Traits) != len(tt.wantTraits) {
				t.Fatalf("ParseScript(%q) traits = %v, want %v", tt.input, script.Traits, tt.wantTraits)
			}
			for key, want := range tt.wantTraits {
				got, ok := script.Traits[key]
				if !ok {
					t.Errorf("ParseScript(%q) traits = %v, want key %q", tt.input, script.Traits, key)
					continue
				}
				if got != want {
					t.Errorf("ParseScript(%q) traits[%q] = %v (%T), want %v (%T)",
						tt.input, key, got, got, want, want)
				}
			}
		})
	}
}

// Two JSON keys differing only in case normalize to one trait. Merging them
// through a map let Go's randomized iteration order pick the winner, so the
// same script parsed to different results run to run.
func TestParseTraitsJSONKeyCollisionIsDeterministic(t *testing.T) {
	t.Parallel()

	inputs := []string{
		`**traits:{"Tap":true,"tap":false,"TAP":"last"}||**echo:hi`,
		`**traits:{"Hold":true,"hold":false}||**echo:hi`,
		`**traits:{"Keep":1,"kept":2}`,
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			first, err := zapscript.NewParser(input).ParseScript()
			require.NoError(t, err)

			for range 200 {
				got, parseErr := zapscript.NewParser(input).ParseScript()
				require.NoError(t, parseErr)
				require.Equal(t, first.Traits, got.Traits, "same input parsed to different traits")
			}
		})
	}
}

// JSON arguments are normalized through a map before the traits merge sees
// them, so there is no member order to resolve a collision by. A contradictory
// object drops the trait rather than picking a winner.
func TestParseTraitsJSONKeyCollisionDropsTrait(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantTraits map[string]any
		name       string
		input      string
	}{
		{
			name:       "collision drops only the contested trait",
			input:      `**traits:{"Tap":true,"tap":false,"hold":true}`,
			wantTraits: map[string]any{"hold": true},
		},
		{
			name:       "collision on every key leaves no traits",
			input:      `**traits:{"Tap":true,"tap":false}||**echo:hi`,
			wantTraits: map[string]any{},
		},
		{
			name:       "three way collision drops the trait",
			input:      `**traits:{"TAP":1,"Tap":2,"tap":3,"hold":true}`,
			wantTraits: map[string]any{"hold": true},
		},
		{
			name:       "unique keys are unaffected",
			input:      `**traits:{"Tap":true,"Hold":false}`,
			wantTraits: map[string]any{"tap": true, "hold": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			script, err := zapscript.NewParser(tt.input).ParseScript()
			require.NoError(t, err)
			assert.Equal(t, tt.wantTraits, script.Traits)
		})
	}
}

func TestParseTraitsJSONDecoding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantTraits map[string]any
		name       string
		input      string
		wantCmds   int
	}{
		{
			name:       "null is consumed and adds nothing",
			input:      `**traits:null||**echo:hi`,
			wantCmds:   1,
			wantTraits: map[string]any{},
		},
		{
			name:       "empty object adds nothing",
			input:      `**traits:{}||**echo:hi`,
			wantCmds:   1,
			wantTraits: map[string]any{},
		},
		{
			name:     "values keep their inferred types",
			input:    `**traits:{"name":"mario","level":5,"flag":true,"items":["a"]}`,
			wantCmds: 0,
			wantTraits: map[string]any{
				"name": "mario", "level": float64(5), "flag": true, "items": []any{"a"},
			},
		},
		{
			name:       "non object json stays a command",
			input:      `**traits:[1,2]||**echo:hi`,
			wantCmds:   2,
			wantTraits: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			script, err := zapscript.NewParser(tt.input).ParseScript()
			require.NoError(t, err)
			assert.Len(t, script.Cmds, tt.wantCmds)
			if tt.wantTraits == nil {
				assert.Empty(t, script.Traits)
				return
			}
			assert.Equal(t, tt.wantTraits, script.Traits)
		})
	}
}
