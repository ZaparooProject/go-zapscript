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

// TestParseAutoAdvanceArgs checks the rotation arguments reach the command as
// plain advanced arguments, for both commands that take them.
func TestParseAutoAdvanceArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		want  map[string]string
		name  string
		input string
		cmd   string
		args  []string
	}{
		{
			name:  "playlist timer",
			input: `**playlist.play:/media/fat/games/Genesis?mode=shuffle&advance=5m&countdown=10s`,
			cmd:   zapscript.ZapScriptCmdPlaylistPlay,
			args:  []string{"/media/fat/games/Genesis"},
			want:  map[string]string{"mode": "shuffle", "advance": "5m", "countdown": "10s"},
		},
		{
			name:  "playlist idle",
			input: `**playlist.load:deck://abc?idle=5m`,
			cmd:   zapscript.ZapScriptCmdPlaylistLoad,
			args:  []string{"deck://abc"},
			want:  map[string]string{"idle": "5m"},
		},
		{
			name:  "random rotation",
			input: `**launch.random:SNES?advance=1h30m&countdown=0`,
			cmd:   zapscript.ZapScriptCmdLaunchRandom,
			args:  []string{"SNES"},
			want:  map[string]string{"advance": "1h30m", "countdown": "0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := zapscript.NewParser(tt.input).ParseScript()
			if err != nil {
				t.Fatalf("ParseScript() error = %v", err)
			}
			if len(got.Cmds) != 1 {
				t.Fatalf("expected 1 command, got %d", len(got.Cmds))
			}
			cmd := got.Cmds[0]
			if cmd.Name != tt.cmd {
				t.Errorf("command = %q, want %q", cmd.Name, tt.cmd)
			}
			if diff := cmp.Diff(tt.args, cmd.Args); diff != "" {
				t.Errorf("args mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, cmd.AdvArgs.Raw()); diff != "" {
				t.Errorf("advanced args mismatch (-want +got):\n%s", diff)
			}

			again, err := zapscript.NewParser(cmd.String()).ParseScript()
			if err != nil {
				t.Fatalf("reparse error = %v", err)
			}
			if diff := cmp.Diff(got, again, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("round-trip mismatch (-original +reparsed):\n%s", diff)
			}
		})
	}
}
