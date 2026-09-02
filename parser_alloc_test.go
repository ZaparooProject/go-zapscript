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
	"strings"
	"testing"

	zapscript "github.com/ZaparooProject/go-zapscript"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseScript_LongInputDoesNotAllocatePerCharacter guards the property
// that makes parse time linear.
//
// Accumulating a rune at a time into a string reallocates and copies the
// whole accumulator on every character, which is two allocations per
// character and quadratic total copying. Growing a strings.Builder instead is
// logarithmic in the length. At this input size the two differ by more than
// three orders of magnitude, so the ceiling below is generous and still fails
// loudly if an accumulator regresses.
//
// This test does not call t.Parallel: testing.AllocsPerRun pins GOMAXPROCS
// for the duration and must not run alongside other tests.
func TestParseScript_LongInputDoesNotAllocatePerCharacter(t *testing.T) {
	const (
		size = 16384
		// Per-character accumulation allocated ~32,800 times at this size.
		maxAllocs = 200.0
	)

	arg := strings.Repeat("A", size)

	tests := []struct {
		name   string
		script string
	}{
		{name: "positional arg", script: "**launch:" + arg},
		{name: "quoted arg", script: `**launch:"` + arg + `"`},
		{name: "advanced arg value", script: "**launch:game?name=" + arg},
		{name: "command name", script: "**" + strings.ToLower(arg)},
		{name: "auto launch content", script: arg},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allocs := testing.AllocsPerRun(2, func() {
				//nolint:errcheck // an unknown command name still parses; only allocation is under test
				_, _ = zapscript.NewParser(tt.script).ParseScript()
			})
			assert.Less(t, allocs, maxAllocs,
				"parse allocated %.0f times for %d characters, which is per-character accumulation",
				allocs, size)
		})
	}
}

// A long argument has to survive the parse intact, not merely parse quickly.
func TestParseScript_LongArgumentRoundTrips(t *testing.T) {
	t.Parallel()

	arg := strings.Repeat("A", 16384)

	tests := []struct {
		name   string
		script string
		want   string
	}{
		{name: "positional arg", script: "**launch:" + arg, want: arg},
		{name: "quoted arg", script: `**launch:"` + arg + `"`, want: arg},
		{name: "auto launch content", script: arg, want: arg},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			script, err := zapscript.NewParser(tt.script).ParseScript()
			require.NoError(t, err)
			require.Len(t, script.Cmds, 1)
			require.Len(t, script.Cmds[0].Args, 1)
			assert.Equal(t, tt.want, script.Cmds[0].Args[0])
		})
	}
}

// The advanced-argument value accumulator is separate from the positional
// one, so it needs its own round trip.
func TestParseScript_LongAdvArgRoundTrips(t *testing.T) {
	t.Parallel()

	value := strings.Repeat("A", 16384)

	script, err := zapscript.NewParser("**launch:game?name=" + value).ParseScript()
	require.NoError(t, err)
	require.Len(t, script.Cmds, 1)
	assert.Equal(t, value, script.Cmds[0].AdvArgs.Get("name"))
}
