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
	"regexp"
	"strings"
)

var (
	reColonSpacing = regexp.MustCompile(`\s*:\s*`)
	reSpecialChars = regexp.MustCompile(`[^a-z0-9:,+\-]`)
)

// NormalizeTag normalizes a tag string for consistent matching.
// Applied to BOTH type and value parts separately.
// Rules: trim whitespace, normalize colon spacing, lowercase, spaces→dashes,
// periods→dashes, and remove special chars (except colon, dash, and comma).
func NormalizeTag(s string) string {
	// 1. Trim whitespace
	s = strings.TrimSpace(s)

	// 2. Normalize colon spacing - remove spaces around colons first
	s = reColonSpacing.ReplaceAllString(s, ":")

	// 3. Convert to lowercase
	s = strings.ToLower(s)

	// 4. Replace remaining spaces with dashes
	s = strings.ReplaceAll(s, " ", "-")

	// 5. Convert periods to dashes (for version numbers like "1.2.3" → "1-2-3")
	s = strings.ReplaceAll(s, ".", "-")

	// 6. Remove other special chars (except colon, dash, and comma)
	// Keep: a-z, 0-9, dash, colon, comma
	s = reSpecialChars.ReplaceAllString(s, "")

	return s
}

// ParseTagFilters parses a comma-separated tag filter string into TagFilter structs.
// Supports operator prefixes:
//   - "+" or no prefix: AND (default) - must have tag
//   - "-": NOT - must not have tag
//   - "~": OR - at least one OR tag must match
//
// Format: "type:value" or "+type:value" (AND), "-type:value" (NOT), "~type:value" (OR)
// Example: "region:usa,-unfinished:demo,~lang:en,~lang:es"
// Returns normalized, deduplicated filters.
func ParseTagFilters(raw string) ([]TagFilter, error) {
	if raw == "" {
		return []TagFilter{}, nil
	}

	parts := strings.Split(raw, ",")

	// Use map for deduplication while maintaining order
	type filterKey struct {
		typ      string
		value    string
		operator TagOperator
	}
	seenFilters := make(map[filterKey]bool)
	result := make([]TagFilter, 0, len(parts))

	for _, tagStr := range parts {
		trimmedTag := strings.TrimSpace(tagStr)
		if trimmedTag == "" {
			continue
		}

		// Parse operator prefix
		operator := TagOperatorAND // default
		switch trimmedTag[0] {
		case '+':
			operator = TagOperatorAND
			trimmedTag = trimmedTag[1:]
		case '-':
			operator = TagOperatorNOT
			trimmedTag = trimmedTag[1:]
		case '~':
			operator = TagOperatorOR
			trimmedTag = trimmedTag[1:]
		}

		// Validate type:value format
		colonIdx := strings.Index(trimmedTag, ":")
		if colonIdx == -1 {
			return nil, fmt.Errorf("invalid tag format for %q: must be in 'type:value' format", tagStr)
		}

		tagType := strings.TrimSpace(trimmedTag[:colonIdx])
		tagValue := strings.TrimSpace(trimmedTag[colonIdx+1:])

		// Apply normalization
		normalizedType := NormalizeTag(tagType)
		normalizedValue := NormalizeTag(tagValue)

		// Validate after normalization
		if normalizedType == "" || normalizedValue == "" {
			return nil, fmt.Errorf("invalid tag %q: type and value cannot be empty after normalization", tagStr)
		}

		filter := TagFilter{
			Type:     normalizedType,
			Value:    normalizedValue,
			Operator: operator,
		}

		// Deduplicate by normalized key (including operator), preserving order
		key := filterKey{
			typ:      filter.Type,
			value:    filter.Value,
			operator: operator,
		}
		if !seenFilters[key] {
			seenFilters[key] = true
			result = append(result, filter)
		}
	}

	return result, nil
}
