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

// TestParseCollectionEdit covers the collection links add and remove take,
// including a deck named with spaces, and that each survives String() and
// reparsing unchanged.
func TestParseCollectionEdit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  zapscript.Command
	}{
		{
			name:  "add to a built-in collection",
			input: `**collection.add:favorites://`,
			want:  zapscript.Command{Name: zapscript.ZapScriptCmdCollectionAdd, Args: []string{"favorites://"}},
		},
		{
			name:  "remove from a built-in collection",
			input: `**collection.remove:playlater://`,
			want:  zapscript.Command{Name: zapscript.ZapScriptCmdCollectionRemove, Args: []string{"playlater://"}},
		},
		{
			name:  "add to a deck by name",
			input: `**collection.add:deck://Arcade Night`,
			want:  zapscript.Command{Name: zapscript.ZapScriptCmdCollectionAdd, Args: []string{"deck://Arcade Night"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := zapscript.NewParser(tt.input).ParseScript()
			if err != nil {
				t.Fatalf("ParseScript() error = %v", err)
			}
			want := zapscript.Script{Cmds: []zapscript.Command{tt.want}}
			if diff := cmp.Diff(want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Fatalf("ParseScript() mismatch (-want +got):\n%s", diff)
			}

			again, err := zapscript.NewParser(got.Cmds[0].String()).ParseScript()
			if err != nil {
				t.Fatalf("reparse error = %v", err)
			}
			if diff := cmp.Diff(want, again, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("round trip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
