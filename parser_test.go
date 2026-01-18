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
	"errors"
	"testing"

	"github.com/ZaparooProject/go-zapscript"
	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		input   string
		want    zapscript.Script
	}{
		{
			name:  "single command with no args",
			input: `**hello`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "hello"},
				},
			},
		},
		{
			name:  "multiple commands with no args",
			input: `**hello||**goodbye||**world`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "hello"},
					{Name: "goodbye"},
					{Name: "world"},
				},
			},
		},
		{
			name:  "single command with args",
			input: `**greet:hi,there`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "greet", Args: []string{"hi", "there"}},
				},
			},
		},
		{
			name:  "two commands separated",
			input: `**first:1,2||**second:3,4`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "first", Args: []string{"1", "2"}},
					{Name: "second", Args: []string{"3", "4"}},
				},
			},
		},
		{
			name:  "whitespace is trimmed in args 1",
			input: `  **trim:  a , b `,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "trim", Args: []string{"a", "b"}},
				},
			},
		},
		{
			name:    "missing command name",
			input:   `**:x,y`,
			wantErr: zapscript.ErrEmptyCmdName,
		},
		{
			name:  "invalid character in command name",
			input: `**he@llo`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{`**he@llo`}},
				},
			},
		},
		{
			name:    "unexpected EOF after asterisk",
			input:   `*`,
			wantErr: zapscript.ErrUnexpectedEOF,
		},
		{
			name:  "command with trailing ||",
			input: `**cmd:1,2||`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "cmd", Args: []string{"1", "2"}},
				},
			},
		},
		{
			name:  "command with one advanced arg",
			input: `**example?debug=true`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "example", AdvArgs: zapscript.NewAdvArgs(map[string]string{"debug": "true"})},
				},
			},
		},
		{
			name:  "command with args and one advanced arg",
			input: `**download:file1.txt?verify=sha256`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{
						Name: "download", Args: []string{"file1.txt"},
						AdvArgs: zapscript.NewAdvArgs(map[string]string{"verify": "sha256"}),
					},
				},
			},
		},
		{
			name:  "command with multiple advanced args",
			input: `**launch:game.exe?platform=win&fullscreen=yes&lang=en`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{"game.exe"}, AdvArgs: zapscript.NewAdvArgs(map[string]string{
						"platform":   "win",
						"fullscreen": "yes",
						"lang":       "en",
					})},
				},
			},
		},
		{
			name:  "generic launch 1",
			input: `DOS/some/game/to/play.iso`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{`DOS/some/game/to/play.iso`}},
				},
			},
		},
		{
			name:  "generic launch 2",
			input: `/media/fat/games/DOS/some/game/to/play.iso`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{`/media/fat/games/DOS/some/game/to/play.iso`}},
				},
			},
		},
		{
			name:  "generic launch 3",
			input: `C:\game\to\to\play.iso`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "launch", Args: []string{`C:\game\to\to\play.iso`}},
				},
			},
		},
		{
			name:  "single quoted arg",
			input: `**say:"hello, world"`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "say", Args: []string{"hello, world"}},
				},
			},
		},
		{
			name:    "unmatched quote in arg",
			input:   `**fail:"unterminated`,
			wantErr: zapscript.ErrUnmatchedQuote,
		},
		{
			name:  "simple json argument",
			input: `**config:{"key": "value"}`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "config", Args: []string{`{"key":"value"}`}},
				},
			},
		},
		{
			name:    "invalid json argument - missing closing brace",
			input:   `**config:{"key": "value"`,
			wantErr: zapscript.ErrInvalidJSON,
		},
		{
			name:    "empty script",
			input:   "",
			want:    zapscript.Script{},
			wantErr: zapscript.ErrEmptyZapScript,
		},
		{
			name:    "empty command name",
			input:   "**",
			want:    zapscript.Script{},
			wantErr: zapscript.ErrEmptyCmdName,
		},
		{
			name:  "input.keyboard basic characters",
			input: `**input.keyboard:abcXYZ123`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.keyboard", Args: []string{
						"a", "b", "c", "X", "Y", "Z", "1", "2", "3",
					}},
				},
			},
		},
		{
			name:  "input.gamepad basic directions",
			input: `**input.gamepad:^^VV<><>`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "input.gamepad", Args: []string{
						"^", "^", "V", "V", "<", ">", "<", ">",
					}},
				},
			},
		},
		{
			name:  "simple expression in arg",
			input: `**greet:Hello [[name]]`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "greet", Args: []string{"Hello " + zapscript.TokExpStart + "name" + zapscript.TokExprEnd}},
				},
			},
		},
		{
			name:    "unmatched expression - missing closing bracket",
			input:   `**test:[[variable`,
			wantErr: zapscript.ErrUnmatchedExpression,
		},
		{
			name:  "escaped newline in arg",
			input: `**echo:hello^nworld`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "echo", Args: []string{"hello\nworld"}},
				},
			},
		},
		{
			name:  "escaped tab in arg",
			input: `**tabby:one^ttwo`,
			want: zapscript.Script{
				Cmds: []zapscript.Command{
					{Name: "tabby", Args: []string{"one\ttwo"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseScript()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseScript() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("ParseScript() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseExpressions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		input   string
		want    string
	}{
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "plain text only",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "single expression",
			input: "hello [[name]]",
			want:  "hello " + zapscript.TokExpStart + "name" + zapscript.TokExprEnd,
		},
		{
			name:  "multiple expressions",
			input: "[[first]] and [[second]]",
			want: zapscript.TokExpStart + "first" + zapscript.TokExprEnd + " and " +
				zapscript.TokExpStart + "second" + zapscript.TokExprEnd,
		},
		{
			name:    "unmatched opening brackets",
			input:   "test[[unclosed",
			want:    "test",
			wantErr: zapscript.ErrUnmatchedExpression,
		},
		{
			name:  "closing brackets without opening",
			input: "test]]closed",
			want:  "test]]closed",
		},
		{
			name:  "escaped brackets",
			input: "text with ^[[escaped]] brackets",
			want:  "text with [[escaped]] brackets",
		},
		{
			name:  "escaped newline",
			input: "line one^nline two",
			want:  "line one\nline two",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := zapscript.NewParser(tt.input)
			got, err := p.ParseExpressions()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseExpressions() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("ParseExpressions() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPostProcess(t *testing.T) {
	t.Parallel()
	tests := []struct {
		wantErr error
		name    string
		input   string
		want    string
	}{
		{
			name:  "empty arg",
			input: "",
			want:  "",
		},
		{
			name:  "value only",
			input: "test",
			want:  "test",
		},
		{
			name:  "expression only",
			input: zapscript.TokExpStart + " 2 + 2 + 2 " + " " + zapscript.TokExprEnd,
			want:  "6",
		},
		{
			name:  "test expression 1",
			input: "something " + zapscript.TokExpStart + "platform" + zapscript.TokExprEnd,
			want:  `something mister`,
		},
		{
			name:  "test expression 2",
			input: "something " + zapscript.TokExpStart + "2+2" + zapscript.TokExprEnd,
			want:  `something 4`,
		},
		{
			name:  "test expression bool 1",
			input: "something " + zapscript.TokExpStart + "true" + zapscript.TokExprEnd,
			want:  `something true`,
		},
		{
			name:    "bad return type",
			input:   zapscript.TokExpStart + "device" + zapscript.TokExprEnd,
			wantErr: zapscript.ErrBadExpressionReturn,
		},
		{
			name:  "test expression int",
			input: zapscript.TokExpStart + "5+5" + zapscript.TokExprEnd,
			want:  `10`,
		},
		{
			name:  "test expression float 1",
			input: zapscript.TokExpStart + "2.5" + zapscript.TokExprEnd,
			want:  `2.5`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			env := zapscript.ArgExprEnv{
				Platform:     "mister",
				Version:      "1.2.3",
				MediaPlaying: true,
				ScanMode:     "tap",
				Device: zapscript.ExprEnvDevice{
					Hostname: "test-device",
					OS:       "linux",
					Arch:     "arm",
				},
			}

			p := zapscript.NewParser(tt.input)
			got, err := p.EvalExpressions(env)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("EvalExpressions() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(zapscript.AdvArgs{})); diff != "" {
				t.Errorf("EvalExpressions() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAdvArgs_WithMustAssignReturn(t *testing.T) {
	t.Parallel()

	const testKey zapscript.Key = "test_key"

	t.Run("With without assign on nil map loses changes", func(t *testing.T) {
		t.Parallel()

		advArgs := zapscript.NewAdvArgs(nil)
		advArgs.With(testKey, "new_value")

		got := advArgs.Get(testKey)
		if got != "" {
			t.Errorf("Expected empty string (change lost without assign), got %q", got)
		}
	})

	t.Run("With does not mutate original", func(t *testing.T) {
		t.Parallel()

		original := zapscript.NewAdvArgs(map[string]string{
			"existing": "value",
		})
		modified := original.With(testKey, "new_value")

		if original.Get(testKey) != "" {
			t.Error("Original should not have new key")
		}

		if modified.Get(testKey) != "new_value" {
			t.Errorf("Modified should have new key, got %q", modified.Get(testKey))
		}
		if modified.Get("existing") != "value" {
			t.Errorf("Modified should preserve existing keys, got %q", modified.Get("existing"))
		}
	})

	t.Run("With assign always preserves changes", func(t *testing.T) {
		t.Parallel()

		advArgs := zapscript.NewAdvArgs(nil)
		advArgs = advArgs.With(testKey, "new_value")

		got := advArgs.Get(testKey)
		if got != "new_value" {
			t.Errorf("Expected %q, got %q", "new_value", got)
		}
	})
}

func TestAdvArgs_GetWhen(t *testing.T) {
	t.Parallel()

	t.Run("when key exists", func(t *testing.T) {
		t.Parallel()

		advArgs := zapscript.NewAdvArgs(map[string]string{
			"when": "true",
		})

		val, ok := advArgs.GetWhen()
		if !ok {
			t.Error("Expected GetWhen to return ok=true when key exists")
		}
		if val != "true" {
			t.Errorf("Expected %q, got %q", "true", val)
		}
	})

	t.Run("when key missing", func(t *testing.T) {
		t.Parallel()

		advArgs := zapscript.NewAdvArgs(map[string]string{
			"other": "value",
		})

		val, ok := advArgs.GetWhen()
		if ok {
			t.Error("Expected GetWhen to return ok=false when key missing")
		}
		if val != "" {
			t.Errorf("Expected empty string, got %q", val)
		}
	})

	t.Run("nil map", func(t *testing.T) {
		t.Parallel()

		advArgs := zapscript.NewAdvArgs(nil)

		val, ok := advArgs.GetWhen()
		if ok {
			t.Error("Expected GetWhen to return ok=false for nil map")
		}
		if val != "" {
			t.Errorf("Expected empty string, got %q", val)
		}
	})
}

func TestAdvArgs_IsEmpty(t *testing.T) {
	t.Parallel()

	t.Run("nil map is empty", func(t *testing.T) {
		t.Parallel()

		advArgs := zapscript.NewAdvArgs(nil)
		if !advArgs.IsEmpty() {
			t.Error("Expected nil map to be empty")
		}
	})

	t.Run("empty map is empty", func(t *testing.T) {
		t.Parallel()

		advArgs := zapscript.NewAdvArgs(map[string]string{})
		if !advArgs.IsEmpty() {
			t.Error("Expected empty map to be empty")
		}
	})

	t.Run("non-empty map is not empty", func(t *testing.T) {
		t.Parallel()

		advArgs := zapscript.NewAdvArgs(map[string]string{"key": "value"})
		if advArgs.IsEmpty() {
			t.Error("Expected non-empty map to not be empty")
		}
	})
}

func TestAdvArgs_Range(t *testing.T) {
	t.Parallel()

	t.Run("iterates all entries", func(t *testing.T) {
		t.Parallel()

		advArgs := zapscript.NewAdvArgs(map[string]string{
			"a": "1",
			"b": "2",
			"c": "3",
		})

		collected := make(map[string]string)
		advArgs.Range(func(k zapscript.Key, v string) bool {
			collected[string(k)] = v
			return true
		})

		if len(collected) != 3 {
			t.Errorf("Expected 3 entries, got %d", len(collected))
		}
		if collected["a"] != "1" || collected["b"] != "2" || collected["c"] != "3" {
			t.Errorf("Unexpected collected values: %v", collected)
		}
	})

	t.Run("stops on false return", func(t *testing.T) {
		t.Parallel()

		advArgs := zapscript.NewAdvArgs(map[string]string{
			"a": "1",
			"b": "2",
			"c": "3",
		})

		count := 0
		advArgs.Range(func(_ zapscript.Key, _ string) bool {
			count++
			return false
		})

		if count != 1 {
			t.Errorf("Expected 1 iteration, got %d", count)
		}
	})

	t.Run("nil map", func(t *testing.T) {
		t.Parallel()

		advArgs := zapscript.NewAdvArgs(nil)

		count := 0
		advArgs.Range(func(_ zapscript.Key, _ string) bool {
			count++
			return true
		})

		if count != 0 {
			t.Errorf("Expected 0 iterations for nil map, got %d", count)
		}
	})
}

func TestAdvArgs_Raw(t *testing.T) {
	t.Parallel()

	t.Run("returns underlying map", func(t *testing.T) {
		t.Parallel()

		m := map[string]string{"key": "value"}
		advArgs := zapscript.NewAdvArgs(m)

		raw := advArgs.Raw()
		if raw["key"] != "value" {
			t.Error("Expected raw map to contain key=value")
		}
	})

	t.Run("nil map returns nil", func(t *testing.T) {
		t.Parallel()

		advArgs := zapscript.NewAdvArgs(nil)
		if advArgs.Raw() != nil {
			t.Error("Expected Raw() to return nil for nil map")
		}
	})
}

func TestAdvArgs_MarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("marshals non-nil map", func(t *testing.T) {
		t.Parallel()

		advArgs := zapscript.NewAdvArgs(map[string]string{
			"key": "value",
		})

		data, err := advArgs.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		expected := `{"key":"value"}`
		if string(data) != expected {
			t.Errorf("Expected %s, got %s", expected, string(data))
		}
	})

	t.Run("marshals nil map as null", func(t *testing.T) {
		t.Parallel()

		advArgs := zapscript.NewAdvArgs(nil)

		data, err := advArgs.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		if string(data) != "null" {
			t.Errorf("Expected null, got %s", string(data))
		}
	})
}

func TestAdvArgs_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("unmarshals valid JSON", func(t *testing.T) {
		t.Parallel()

		var advArgs zapscript.AdvArgs
		err := advArgs.UnmarshalJSON([]byte(`{"key":"value"}`))
		if err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		if advArgs.Get("key") != "value" {
			t.Error("Expected key=value after unmarshal")
		}
	})

	t.Run("unmarshals null as nil map", func(t *testing.T) {
		t.Parallel()

		var advArgs zapscript.AdvArgs
		err := advArgs.UnmarshalJSON([]byte(`null`))
		if err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		if !advArgs.IsEmpty() {
			t.Error("Expected empty AdvArgs after unmarshalling null")
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		t.Parallel()

		var advArgs zapscript.AdvArgs
		err := advArgs.UnmarshalJSON([]byte(`{invalid}`))
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}
