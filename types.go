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

// Package zapscript provides a parser for ZapScript, a custom scripting language
// for launching games and executing commands through physical tokens.
package zapscript

// TagOperator defines the logical operator for tag filtering.
type TagOperator string

const (
	TagOperatorAND TagOperator = "AND"
	TagOperatorNOT TagOperator = "NOT"
	TagOperatorOR  TagOperator = "OR"
)

// TagFilter represents a filter for matching media by tags.
type TagFilter struct {
	Type     string
	Value    string
	Operator TagOperator
}

// Key is a typed key for advanced argument map lookups.
type Key string

// Argument key constants for advarg struct tags and map lookups.
const (
	KeyWhen      Key = "when"
	KeyLauncher  Key = "launcher"
	KeySystem    Key = "system"
	KeyAction    Key = "action"
	KeyTags      Key = "tags"
	KeyMode      Key = "mode"
	KeyName      Key = "name"
	KeyPreNotice Key = "pre_notice"
	KeyHidden    Key = "hidden"
)

// Action values for the action advanced argument.
const (
	// ActionRun is the default action - launch/play the media.
	ActionRun = "run"
	// ActionDetails shows the media details/info page instead of launching.
	ActionDetails = "details"
)

// Mode values for the mode advanced argument.
const (
	// ModeShuffle randomizes playlist order.
	ModeShuffle = "shuffle"
)

// GlobalArgs contains advanced arguments available to all commands.
type GlobalArgs struct {
	// When controls conditional execution. If non-empty and falsy, command is skipped.
	When string `advarg:"when"`
}

// LaunchArgs contains advanced arguments for the launch command.
type LaunchArgs struct {
	GlobalArgs
	// Launcher overrides the default launcher by ID.
	Launcher string `advarg:"launcher" validate:"omitempty,launcher"` //nolint:revive // custom validator
	// System specifies the target system for path resolution.
	System string `advarg:"system" validate:"omitempty,system"` //nolint:revive // custom validator
	// Action specifies the launch action (run, details).
	Action string `advarg:"action" validate:"omitempty,oneof=run details"`
	// Name is the filename for remote file installation.
	Name string `advarg:"name"`
	// PreNotice is shown before remote file download.
	PreNotice string `advarg:"pre_notice"`
}

// LaunchRandomArgs contains advanced arguments for the launch.random command.
type LaunchRandomArgs struct {
	GlobalArgs
	// Launcher overrides the default launcher by ID.
	Launcher string `advarg:"launcher" validate:"omitempty,launcher"` //nolint:revive // custom validator
	// Action specifies the launch action (run, details).
	Action string `advarg:"action" validate:"omitempty,oneof=run details"`
	// Tags filters results by tag criteria.
	Tags []TagFilter `advarg:"tags"`
}

// LaunchSearchArgs contains advanced arguments for the launch.search command.
type LaunchSearchArgs struct {
	GlobalArgs
	// Launcher overrides the default launcher by ID.
	Launcher string `advarg:"launcher" validate:"omitempty,launcher"` //nolint:revive // custom validator
	// Action specifies the launch action (run, details).
	Action string `advarg:"action" validate:"omitempty,oneof=run details"`
	// Tags filters results by tag criteria.
	Tags []TagFilter `advarg:"tags"`
}

// LaunchTitleArgs contains advanced arguments for the launch.title command.
type LaunchTitleArgs struct {
	GlobalArgs
	// Launcher overrides the default launcher by ID.
	Launcher string `advarg:"launcher" validate:"omitempty,launcher"` //nolint:revive // custom validator
	// Action specifies the launch action (run, details).
	Action string `advarg:"action" validate:"omitempty,oneof=run details"`
	// Tags filters results by tag criteria.
	Tags []TagFilter `advarg:"tags"`
}

// PlaylistArgs contains advanced arguments for playlist commands.
type PlaylistArgs struct {
	GlobalArgs
	// Mode controls playlist behavior (e.g., "shuffle").
	Mode string `advarg:"mode" validate:"omitempty,oneof=shuffle"`
}

// MisterScriptArgs contains advanced arguments for MiSTer script commands.
type MisterScriptArgs struct {
	GlobalArgs
	// Hidden controls whether the script window is hidden.
	Hidden string `advarg:"hidden"`
}
