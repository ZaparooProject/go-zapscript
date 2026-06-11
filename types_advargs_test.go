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

import (
	"reflect"
	"testing"
)

func TestLaunchArgsSetNameAdvargFields(t *testing.T) {
	t.Parallel()

	requireAdvargField(t, reflect.TypeOf(LaunchArgs{}), "set_name")
	requireAdvargField(t, reflect.TypeOf(LaunchArgs{}), "set_name_same_dir")
	requireAdvargField(t, reflect.TypeOf(LaunchRandomArgs{}), "set_name")
	requireAdvargField(t, reflect.TypeOf(LaunchRandomArgs{}), "set_name_same_dir")
	requireAdvargField(t, reflect.TypeOf(LaunchSearchArgs{}), "set_name")
	requireAdvargField(t, reflect.TypeOf(LaunchSearchArgs{}), "set_name_same_dir")
	requireAdvargField(t, reflect.TypeOf(LaunchTitleArgs{}), "set_name")
	requireAdvargField(t, reflect.TypeOf(LaunchTitleArgs{}), "set_name_same_dir")
	requireAdvargField(t, reflect.TypeOf(LaunchLastArgs{}), "set_name")
	requireAdvargField(t, reflect.TypeOf(LaunchLastArgs{}), "set_name_same_dir")
}

func TestSlotAdvargFields(t *testing.T) {
	t.Parallel()

	requireAdvargField(t, reflect.TypeOf(LaunchArgs{}), "slot")
	requireAdvargField(t, reflect.TypeOf(LaunchRandomArgs{}), "slot")
	requireAdvargField(t, reflect.TypeOf(LaunchSearchArgs{}), "slot")
	requireAdvargField(t, reflect.TypeOf(LaunchTitleArgs{}), "slot")
	requireAdvargField(t, reflect.TypeOf(LaunchLastArgs{}), "slot")
	requireAdvargField(t, reflect.TypeOf(PlaylistArgs{}), "slot")
}

func requireAdvargField(t *testing.T, typ reflect.Type, tag string) {
	t.Helper()

	for i := range typ.NumField() {
		field := typ.Field(i)
		if field.Tag.Get("advarg") == tag {
			return
		}
	}
	t.Fatalf("%s missing advarg field %q", typ.Name(), tag)
}
