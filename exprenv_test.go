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
	"encoding/json"
	"testing"

	"github.com/ZaparooProject/go-zapscript"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestArgExprEnv_JSONSerialization verifies that expression env serializes correctly to JSON
// with snake_case field names for external script consumption.
func TestArgExprEnv_JSONSerialization(t *testing.T) {
	t.Parallel()

	env := zapscript.ArgExprEnv{
		Platform: "mister",
		Version:  "2.0.0",
		ScanMode: "hold",
		Device: zapscript.ExprEnvDevice{
			Hostname: "mister",
			OS:       "linux",
			Arch:     "arm",
		},
		LastScanned: zapscript.ExprEnvLastScanned{
			ID:    "abc123",
			Value: "**launch:snes/mario",
			Data:  "extra-data",
		},
		MediaPlaying: true,
		ActiveMedia: zapscript.ExprEnvActiveMedia{
			LauncherID: "retroarch",
			SystemID:   "snes",
			SystemName: "Super Nintendo",
			Path:       "/games/snes/mario.sfc",
			Name:       "Super Mario World",
		},
	}

	jsonBytes, err := json.Marshal(env)
	require.NoError(t, err, "should marshal to JSON")

	jsonStr := string(jsonBytes)

	// Verify snake_case keys are used (not camelCase)
	assert.Contains(t, jsonStr, `"platform"`, "should contain platform field")
	assert.Contains(t, jsonStr, `"version"`, "should contain version field")
	assert.Contains(t, jsonStr, `"scan_mode"`, "should contain scan_mode field")
	assert.Contains(t, jsonStr, `"media_playing"`, "should contain media_playing field")
	assert.Contains(t, jsonStr, `"active_media"`, "should contain active_media field")
	assert.Contains(t, jsonStr, `"last_scanned"`, "should contain last_scanned field")
	assert.Contains(t, jsonStr, `"launcher_id"`, "should contain launcher_id field")
	assert.Contains(t, jsonStr, `"system_id"`, "should contain system_id field")
	assert.Contains(t, jsonStr, `"system_name"`, "should contain system_name field")

	// Verify values are correct
	assert.Contains(t, jsonStr, `"mister"`, "should contain platform value")
	assert.Contains(t, jsonStr, `"2.0.0"`, "should contain version value")
	assert.Contains(t, jsonStr, `"hold"`, "should contain scan_mode value")
	assert.Contains(t, jsonStr, `true`, "should contain media_playing value")
}

// TestArgExprEnv_JSONRoundTrip verifies JSON can be unmarshalled back.
func TestArgExprEnv_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := zapscript.ArgExprEnv{
		Platform:     "test",
		Version:      "1.0.0",
		ScanMode:     "tap",
		MediaPlaying: true,
		Device: zapscript.ExprEnvDevice{
			Hostname: "testhost",
			OS:       "linux",
			Arch:     "amd64",
		},
		LastScanned: zapscript.ExprEnvLastScanned{
			ID:    "id123",
			Value: "value",
			Data:  "data",
		},
		ActiveMedia: zapscript.ExprEnvActiveMedia{
			LauncherID: "launcher",
			SystemID:   "system",
			SystemName: "System Name",
			Path:       "/path/to/game",
			Name:       "Game Name",
		},
	}

	jsonBytes, err := json.Marshal(original)
	require.NoError(t, err, "should marshal to JSON")

	var decoded zapscript.ArgExprEnv
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err, "should unmarshal from JSON")

	assert.Equal(t, original.Platform, decoded.Platform)
	assert.Equal(t, original.Version, decoded.Version)
	assert.Equal(t, original.ScanMode, decoded.ScanMode)
	assert.Equal(t, original.MediaPlaying, decoded.MediaPlaying)
	assert.Equal(t, original.Device.Hostname, decoded.Device.Hostname)
	assert.Equal(t, original.LastScanned.ID, decoded.LastScanned.ID)
	assert.Equal(t, original.ActiveMedia.Path, decoded.ActiveMedia.Path)
}

// TestExprEnvScanned_JSONSerialization verifies scanned context serializes correctly.
func TestExprEnvScanned_JSONSerialization(t *testing.T) {
	t.Parallel()

	scanned := zapscript.ExprEnvScanned{
		ID:    "test-id",
		Value: "test-value",
		Data:  "test-data",
	}

	jsonBytes, err := json.Marshal(scanned)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"id"`)
	assert.Contains(t, jsonStr, `"value"`)
	assert.Contains(t, jsonStr, `"data"`)
}

// TestExprEnvLaunching_JSONSerialization verifies launching context serializes correctly.
func TestExprEnvLaunching_JSONSerialization(t *testing.T) {
	t.Parallel()

	launching := zapscript.ExprEnvLaunching{
		Path:       "/path/to/game.rom",
		SystemID:   "snes",
		LauncherID: "retroarch",
	}

	jsonBytes, err := json.Marshal(launching)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"path"`)
	assert.Contains(t, jsonStr, `"system_id"`)
	assert.Contains(t, jsonStr, `"launcher_id"`)
}

// TestArgExprEnv_EmptyFieldsSerialization verifies that empty struct fields are serialized
// (no omitempty behavior) for consistent JSON structure in external scripts.
func TestArgExprEnv_EmptyFieldsSerialization(t *testing.T) {
	t.Parallel()

	// Create env with empty Scanned and Launching
	env := zapscript.ArgExprEnv{
		Platform: "test",
		Version:  "1.0.0",
		// Scanned and Launching are zero values
	}

	jsonBytes, err := json.Marshal(env)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)

	// Empty struct fields should still be present for consistent JSON schema
	// Scripts can rely on field existence rather than checking for presence
	assert.Contains(t, jsonStr, `"scanned"`, "scanned field should be present even when empty")
	assert.Contains(t, jsonStr, `"launching"`, "launching field should be present even when empty")
	assert.Contains(t, jsonStr, `"platform":"test"`, "platform should have correct value")
	assert.Contains(t, jsonStr, `"version":"1.0.0"`, "version should have correct value")
}
