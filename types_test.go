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

package zapscript

import "testing"

func TestLaunchArgsSetNameFields(t *testing.T) {
	t.Parallel()

	args := LaunchArgs{
		LaunchSetNameArgs: LaunchSetNameArgs{
			SetName:        "RA_NES",
			SetNameSameDir: "1",
		},
	}

	if args.SetName != "RA_NES" {
		t.Fatalf("SetName = %q, want %q", args.SetName, "RA_NES")
	}
	if args.SetNameSameDir != "1" {
		t.Fatalf("SetNameSameDir = %q, want %q", args.SetNameSameDir, "1")
	}
}

func TestLaunchSetNameKeys(t *testing.T) {
	t.Parallel()

	if KeySetName != "set_name" {
		t.Fatalf("KeySetName = %q, want %q", KeySetName, "set_name")
	}
	if KeySetNameSameDir != "set_name_same_dir" {
		t.Fatalf("KeySetNameSameDir = %q, want %q", KeySetNameSameDir, "set_name_same_dir")
	}
}
