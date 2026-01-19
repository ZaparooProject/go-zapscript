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

	"github.com/google/go-cmp/cmp"
)

func TestNormalizeTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "lowercase",
			input: "REGION",
			want:  "region",
		},
		{
			name:  "trim whitespace",
			input: "  region  ",
			want:  "region",
		},
		{
			name:  "spaces to dashes",
			input: "my tag",
			want:  "my-tag",
		},
		{
			name:  "periods to dashes",
			input: "version.1.2.3",
			want:  "version-1-2-3",
		},
		{
			name:  "colon spacing normalized",
			input: "type : value",
			want:  "type:value",
		},
		{
			name:  "special chars removed",
			input: "tag!@#$%name",
			want:  "tagname",
		},
		{
			name:  "keeps allowed chars",
			input: "my-tag:value,other",
			want:  "my-tag:value,other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NormalizeTag(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeTag(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseTagFilters_Operators(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []TagFilter
		wantErr bool
	}{
		{
			name:  "AND operator default",
			input: "region:usa",
			want: []TagFilter{
				{Type: "region", Value: "usa", Operator: TagOperatorAND},
			},
		},
		{
			name:  "AND operator explicit",
			input: "+region:usa",
			want: []TagFilter{
				{Type: "region", Value: "usa", Operator: TagOperatorAND},
			},
		},
		{
			name:  "NOT operator",
			input: "-region:usa",
			want: []TagFilter{
				{Type: "region", Value: "usa", Operator: TagOperatorNOT},
			},
		},
		{
			name:  "OR operator",
			input: "~region:usa",
			want: []TagFilter{
				{Type: "region", Value: "usa", Operator: TagOperatorOR},
			},
		},
		{
			name:  "mixed operators",
			input: "region:usa,-lang:ja,~genre:rpg",
			want: []TagFilter{
				{Type: "region", Value: "usa", Operator: TagOperatorAND},
				{Type: "lang", Value: "ja", Operator: TagOperatorNOT},
				{Type: "genre", Value: "rpg", Operator: TagOperatorOR},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseTagFilters(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseTagFilters(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseTagFilters(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func TestParseTagFilters_Normalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []TagFilter
	}{
		{
			name:  "lowercases type and value",
			input: "REGION:USA",
			want: []TagFilter{
				{Type: "region", Value: "usa", Operator: TagOperatorAND},
			},
		},
		{
			name:  "trims whitespace",
			input: "  region : usa  ",
			want: []TagFilter{
				{Type: "region", Value: "usa", Operator: TagOperatorAND},
			},
		},
		{
			name:  "spaces to dashes in value",
			input: "genre:role playing game",
			want: []TagFilter{
				{Type: "genre", Value: "role-playing-game", Operator: TagOperatorAND},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseTagFilters(tt.input)
			if err != nil {
				t.Fatalf("ParseTagFilters(%q) unexpected error: %v", tt.input, err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseTagFilters(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func TestParseTagFilters_Deduplication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []TagFilter
	}{
		{
			name:  "exact duplicates removed",
			input: "region:usa,region:usa",
			want: []TagFilter{
				{Type: "region", Value: "usa", Operator: TagOperatorAND},
			},
		},
		{
			name:  "case-different duplicates removed",
			input: "region:usa,REGION:USA",
			want: []TagFilter{
				{Type: "region", Value: "usa", Operator: TagOperatorAND},
			},
		},
		{
			name:  "different operators kept",
			input: "region:usa,-region:usa",
			want: []TagFilter{
				{Type: "region", Value: "usa", Operator: TagOperatorAND},
				{Type: "region", Value: "usa", Operator: TagOperatorNOT},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseTagFilters(tt.input)
			if err != nil {
				t.Fatalf("ParseTagFilters(%q) unexpected error: %v", tt.input, err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseTagFilters(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func TestParseTagFilters_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []TagFilter
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			want:    []TagFilter{},
			wantErr: false,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			want:    []TagFilter{},
			wantErr: false,
		},
		{
			name:  "empty between commas",
			input: "region:usa,,lang:en",
			want: []TagFilter{
				{Type: "region", Value: "usa", Operator: TagOperatorAND},
				{Type: "lang", Value: "en", Operator: TagOperatorAND},
			},
			wantErr: false,
		},
		{
			name:    "missing colon",
			input:   "invalidtag",
			wantErr: true,
		},
		{
			name:    "empty type after normalization",
			input:   "!!!:value",
			wantErr: true,
		},
		{
			name:    "empty value after normalization",
			input:   "type:!!!",
			wantErr: true,
		},
		{
			name:  "value with colon",
			input: "url:http://example.com",
			want: []TagFilter{
				{Type: "url", Value: "http:example-com", Operator: TagOperatorAND},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseTagFilters(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseTagFilters(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("ParseTagFilters(%q) mismatch (-want +got):\n%s", tt.input, diff)
				}
			}
		})
	}
}
