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
	"fmt"
	"strings"
	"testing"

	zapscript "github.com/ZaparooProject/go-zapscript"
)

// benchArgLengths span the range a caller can realistically hand the parser.
// Parse cost must stay proportional to length: a doubling that costs four
// times as much is the quadratic accumulator regressing.
var benchArgLengths = []int{1024, 4096, 16384, 65536}

func BenchmarkParseScript_LongArg(b *testing.B) {
	for _, size := range benchArgLengths {
		script := "**launch:" + strings.Repeat("A", size)
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := zapscript.NewParser(script).ParseScript(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkParseScript_LongQuotedArg(b *testing.B) {
	for _, size := range benchArgLengths {
		script := `**launch:"` + strings.Repeat("A", size) + `"`
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := zapscript.NewParser(script).ParseScript(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkParseScript_LongAdvArg(b *testing.B) {
	for _, size := range benchArgLengths {
		script := "**launch:game?name=" + strings.Repeat("A", size)
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := zapscript.NewParser(script).ParseScript(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkParseScript_LongCmdName covers the command-name accumulator, which
// is reached before argument parsing starts.
func BenchmarkParseScript_LongCmdName(b *testing.B) {
	for _, size := range benchArgLengths {
		script := "**" + strings.Repeat("a", size)
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				// An unknown command name still parses; only the name
				// accumulator is being measured here.
				_, _ = zapscript.NewParser(script).ParseScript()
			}
		})
	}
}

// BenchmarkParseScript_TypicalToken is the shape that actually runs on a tap.
// It is dominated by per-parse fixed costs rather than argument length.
func BenchmarkParseScript_TypicalToken(b *testing.B) {
	scripts := map[string]string{
		"launch_path":   "**launch:/media/fat/games/SNES/Super Metroid (USA).sfc",
		"media_title":   "@SNES/Super Metroid",
		"bare_path":     "/media/fat/games/SNES/Super Metroid (USA).sfc",
		"command_chain": "**launch.system:snes||**delay:500||**input.keyboard:{f12}",
		"adv_args":      "**launch.search:mario?system=snes&launcher=retroarch",
	}

	for name, script := range scripts {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := zapscript.NewParser(script).ParseScript(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkEvalExpressions_NoExpression is the common case on the launch
// path: every argument of every command is evaluated, and almost none of them
// contain an expression.
func BenchmarkEvalExpressions_NoExpression(b *testing.B) {
	arg := "/media/fat/games/SNES/Super Metroid (USA).sfc"
	b.ReportAllocs()
	for b.Loop() {
		if _, err := zapscript.NewParser(arg).EvalExpressions(zapscript.ArgExprEnv{}); err != nil {
			b.Fatal(err)
		}
	}
}
