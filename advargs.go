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

import "strings"

// IsActionDetails returns true if the action is "details" (case-insensitive).
func IsActionDetails(action string) bool {
	return strings.EqualFold(action, ActionDetails)
}

// IsActionRun returns true if the action is "run" or empty (case-insensitive).
func IsActionRun(action string) bool {
	return action == "" || strings.EqualFold(action, ActionRun)
}

// IsModeShuffle returns true if the mode is "shuffle" (case-insensitive).
func IsModeShuffle(mode string) bool {
	return strings.EqualFold(mode, ModeShuffle)
}
