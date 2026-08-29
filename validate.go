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
	"fmt"
	"strings"
)

// ValidateTraitKey checks a trait key that did not come from the parser, such
// as one supplied by a user or an API caller. A valid key starts with a letter
// and continues with letters, digits and underscores.
//
// It applies the same rules the parser uses while reading trait keys, so
// callers holding a key as a plain string reach the same verdict the parser
// would. Errors wrap ErrInvalidTraitKey.
func ValidateTraitKey(key string) error {
	if key == "" {
		return fmt.Errorf("%w: key is empty", ErrInvalidTraitKey)
	}

	// Positions are counted in runes so a multi-byte character is reported at
	// the place a reader would count it, not at its byte offset.
	for i, ch := range []rune(key) {
		if i == 0 {
			if !isAdvArgNameStart(ch) {
				return fmt.Errorf("%w: %q must start with a letter", ErrInvalidTraitKey, key)
			}
			continue
		}
		if !isAdvArgName(ch) {
			return fmt.Errorf(
				"%w: %q has invalid character %q at position %d",
				ErrInvalidTraitKey, key, ch, i+1,
			)
		}
	}

	return nil
}

// NormalizeTraitKey returns key in the form the parser stores it as, so a key
// held as a plain string can be compared against a parsed Script's traits.
func NormalizeTraitKey(key string) string {
	return strings.ToLower(key)
}
