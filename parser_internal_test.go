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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCmdName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		char     rune
		expected bool
	}{
		// Valid characters
		{
			name:     "lowercase_a",
			char:     'a',
			expected: true,
		},
		{
			name:     "lowercase_z",
			char:     'z',
			expected: true,
		},
		{
			name:     "uppercase_A",
			char:     'A',
			expected: true,
		},
		{
			name:     "uppercase_Z",
			char:     'Z',
			expected: true,
		},
		{
			name:     "digit_0",
			char:     '0',
			expected: true,
		},
		{
			name:     "digit_9",
			char:     '9',
			expected: true,
		},
		{
			name:     "dot",
			char:     '.',
			expected: true,
		},
		// Invalid characters
		{
			name:     "underscore",
			char:     '_',
			expected: false,
		},
		{
			name:     "dash",
			char:     '-',
			expected: false,
		},
		{
			name:     "space",
			char:     ' ',
			expected: false,
		},
		{
			name:     "exclamation",
			char:     '!',
			expected: false,
		},
		{
			name:     "at_symbol",
			char:     '@',
			expected: false,
		},
		{
			name:     "hash",
			char:     '#',
			expected: false,
		},
		{
			name:     "unicode_letter",
			char:     'ñ',
			expected: false,
		},
		{
			name:     "tab",
			char:     '\t',
			expected: false,
		},
		{
			name:     "newline",
			char:     '\n',
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := isCmdName(tt.char)
			assert.Equal(t, tt.expected, result, "isCmdName result mismatch for character '%c'", tt.char)
		})
	}
}

func TestIsAdvArgName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		char     rune
		expected bool
	}{
		// Valid characters
		{
			name:     "lowercase_a",
			char:     'a',
			expected: true,
		},
		{
			name:     "lowercase_z",
			char:     'z',
			expected: true,
		},
		{
			name:     "uppercase_A",
			char:     'A',
			expected: true,
		},
		{
			name:     "uppercase_Z",
			char:     'Z',
			expected: true,
		},
		{
			name:     "digit_0",
			char:     '0',
			expected: true,
		},
		{
			name:     "digit_9",
			char:     '9',
			expected: true,
		},
		{
			name:     "underscore",
			char:     '_',
			expected: true,
		},
		// Invalid characters
		{
			name:     "dot",
			char:     '.',
			expected: false,
		},
		{
			name:     "dash",
			char:     '-',
			expected: false,
		},
		{
			name:     "space",
			char:     ' ',
			expected: false,
		},
		{
			name:     "exclamation",
			char:     '!',
			expected: false,
		},
		{
			name:     "at_symbol",
			char:     '@',
			expected: false,
		},
		{
			name:     "hash",
			char:     '#',
			expected: false,
		},
		{
			name:     "unicode_letter",
			char:     'ñ',
			expected: false,
		},
		{
			name:     "tab",
			char:     '\t',
			expected: false,
		},
		{
			name:     "newline",
			char:     '\n',
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := isAdvArgName(tt.char)
			assert.Equal(t, tt.expected, result, "isAdvArgName result mismatch for character '%c'", tt.char)
		})
	}
}

func TestIsWhitespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		char     rune
		expected bool
	}{
		// Valid whitespace characters
		{
			name:     "space",
			char:     ' ',
			expected: true,
		},
		{
			name:     "tab",
			char:     '\t',
			expected: true,
		},
		{
			name:     "newline",
			char:     '\n',
			expected: true,
		},
		{
			name:     "carriage_return",
			char:     '\r',
			expected: true,
		},
		// Non-whitespace characters
		{
			name:     "letter_a",
			char:     'a',
			expected: false,
		},
		{
			name:     "letter_Z",
			char:     'Z',
			expected: false,
		},
		{
			name:     "digit_0",
			char:     '0',
			expected: false,
		},
		{
			name:     "underscore",
			char:     '_',
			expected: false,
		},
		{
			name:     "dot",
			char:     '.',
			expected: false,
		},
		{
			name:     "exclamation",
			char:     '!',
			expected: false,
		},
		{
			name:     "unicode_space",
			char:     '\u00A0', // Non-breaking space
			expected: false,    // Function only checks for specific ASCII whitespace
		},
		{
			name:     "form_feed",
			char:     '\f',
			expected: false, // Not included in the function
		},
		{
			name:     "vertical_tab",
			char:     '\v',
			expected: false, // Not included in the function
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := isWhitespace(tt.char)
			assert.Equal(t, tt.expected, result,
				"isWhitespace result mismatch for character '%c' (0x%X)", tt.char, tt.char)
		})
	}
}
