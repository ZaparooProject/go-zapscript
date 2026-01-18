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
)

// FuzzParseScript tests that ParseScript never panics on arbitrary input.
// The parser should either return a valid Script or an error, never crash.
func FuzzParseScript(f *testing.F) {
	// Seed corpus with valid and edge-case inputs
	seeds := []string{
		// Valid commands
		`**launch:game.rom`,
		`**delay:1000`,
		`**launch.title:snes/Super Mario World`,
		`**cmd:arg1,arg2,arg3?key=value&other=thing`,
		// Chained commands
		`**launch:game||**delay:500||**notify:done`,
		// Generic launch (no ** prefix)
		`/path/to/game.rom`,
		`Genesis/Sonic.md?launcher=custom`,
		// Media title syntax
		`@snes/Super Mario World`,
		`@genesis/Sonic (USA) (Rev 1)?tags=region:us`,
		// Expressions
		`**launch:[[game_path]]`,
		`**notify:Hello [[username]]!`,
		// Quotes
		`**cmd:"quoted arg",unquoted`,
		`**cmd:'single quotes'`,
		// Escapes
		`**cmd:arg^,with^,commas`,
		`**path:C^:^/Games^/ROM.bin`,
		// JSON-like
		`**api:{"key": "value"}`,
		// Edge cases
		``,
		`**`,
		`**:`,
		`**cmd:`,
		`||`,
		`||||`,
		`**cmd?`,
		`**cmd?=`,
		`**cmd?key=`,
		`**cmd?=value`,
		// Malformed
		`[[`,
		`]]`,
		`[[unclosed`,
		`"unclosed quote`,
		`'unclosed single`,
		// Special characters
		`**cmd:émoji🎮`,
		`**cmd:日本語`, //nolint:gosmopolitan // Japanese test case
		`**cmd:	tabs	and  spaces`,
		// Long input
		`**cmd:` + string(make([]byte, 1000)),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(_ *testing.T, input string) {
		p := NewParser(input)
		// Should not panic - either returns result or error
		_, _ = p.ParseScript()
	})
}

// FuzzParseExpressions tests that ParseExpressions never panics.
// Expression parsing handles [[variable]] syntax.
func FuzzParseExpressions(f *testing.F) {
	seeds := []string{
		// Valid expressions
		`[[var]]`,
		`[[a.b.c]]`,
		`[[under_score]]`,
		`prefix [[var]] suffix`,
		`[[one]] and [[two]]`,
		// Nested/complex
		`[[var1]][[var2]]`,
		`text[[var]]more[[other]]end`,
		// Edge cases
		``,
		`no expressions here`,
		`[[]]`,
		`[[ ]]`,
		`[[`,
		`]]`,
		`[`,
		`]`,
		`[[unclosed`,
		`unclosed]]`,
		`[[][]]`,
		`[[[[nested]]]]`,
		// Special chars in expressions
		`[[var-name]]`,
		`[[var.name]]`,
		`[[var_name]]`,
		`[[123]]`,
		`[[über]]`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(_ *testing.T, input string) {
		p := NewParser(input)
		// Should not panic
		_, _ = p.ParseExpressions()
	})
}

// FuzzEvalExpressions tests expression evaluation with arbitrary environments.
// First parses expressions, then evaluates them with a variable environment.
func FuzzEvalExpressions(f *testing.F) {
	// Seed with (input, varName, varValue) tuples
	seeds := []struct {
		input    string
		varName  string
		varValue string
	}{
		{`[[game]]`, `game`, `mario.rom`},
		{`[[path]]/[[file]]`, `path`, `/roms`},
		{`[[path]]/[[file]]`, `file`, `game.bin`},
		{`Hello [[name]]!`, `name`, `World`},
		{`[[a]][[b]][[c]]`, `a`, `1`},
		{`no vars`, `unused`, `value`},
		{`[[missing]]`, `other`, `value`},
		// Edge cases
		{``, ``, ``},
		{`[[]]`, ``, `value`},
		{`[[var]]`, `var`, ``},
		{`[[var]]`, ``, ``},
		// Special values
		{`[[x]]`, `x`, `value with spaces`},
		{`[[x]]`, `x`, `value,with,commas`},
		{`[[x]]`, `x`, `value||pipes`},
		{`[[x]]`, `x`, `[[nested]]`},
		{`[[x]]`, `x`, `"quotes"`},
	}

	for _, s := range seeds {
		f.Add(s.input, s.varName, s.varValue)
	}

	f.Fuzz(func(_ *testing.T, input, varName, varValue string) {
		// First parse expressions to get internal token format
		p := NewParser(input)
		parsed, err := p.ParseExpressions()
		if err != nil {
			return // Parse failed, nothing to evaluate
		}

		env := make(map[string]string)
		if varName != "" {
			env[varName] = varValue
		}

		// Now evaluate the parsed string with the environment
		evalParser := NewParser(parsed)
		// Should not panic
		_, _ = evalParser.EvalExpressions(env)
	})
}
