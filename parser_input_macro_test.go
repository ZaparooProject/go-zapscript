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
	"strings"
	"testing"

	zapscript "github.com/ZaparooProject/go-zapscript"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseInputMacroArgs is a test helper that parses **input.keyboard:<macro> and
// returns the Args slice so tests can assert on expanded tokens directly.
func parseInputMacroArgs(t *testing.T, macro string) []string {
	t.Helper()
	p := zapscript.NewParser("**input.keyboard:" + macro)
	script, err := p.ParseScript()
	require.NoError(t, err)
	require.Len(t, script.Cmds, 1)
	return script.Cmds[0].Args
}

// parseInputTextArgs is a test helper for **input.text:<raw>.
func parseInputTextArgs(t *testing.T, raw string) []string {
	t.Helper()
	p := zapscript.NewParser("**input.text:" + raw)
	script, err := p.ParseScript()
	require.NoError(t, err)
	require.Len(t, script.Cmds, 1)
	return script.Cmds[0].Args
}

// ─── Backward-compatibility ──────────────────────────────────────────────────

func TestInputMacro_BackwardCompat_BareChars(t *testing.T) {
	t.Parallel()
	// "hello" outside braces: each char is its own arg (unchanged behaviour)
	got := parseInputMacroArgs(t, "hello")
	assert.Equal(t, []string{"h", "e", "l", "l", "o"}, got)
}

func TestInputMacro_BackwardCompat_Space(t *testing.T) {
	t.Parallel()
	// Spaces are literal (issue #939 context: " *MENU" keeps the leading space)
	got := parseInputMacroArgs(t, " *MENU")
	assert.Equal(t, []string{" ", "*", "M", "E", "N", "U"}, got)
}

func TestInputMacro_BackwardCompat_BracedSpecial(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{enter}")
	assert.Equal(t, []string{"{enter}"}, got)
}

func TestInputMacro_BackwardCompat_Combo(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{ctrl+c}")
	assert.Equal(t, []string{"{ctrl+c}"}, got)
}

func TestInputMacro_BackwardCompat_EscapeSeq(t *testing.T) {
	t.Parallel()
	// \\ before a char escapes it
	got := parseInputMacroArgs(t, `\{`)
	assert.Equal(t, []string{"{"}, got)
}

func TestInputMacro_BackwardCompat_AdvArgs(t *testing.T) {
	t.Parallel()
	p := zapscript.NewParser("**input.keyboard:ab?speed=50")
	script, err := p.ParseScript()
	require.NoError(t, err)
	require.Len(t, script.Cmds, 1)
	cmd := script.Cmds[0]
	assert.Equal(t, []string{"a", "b"}, cmd.Args)
	assert.Equal(t, "50", cmd.AdvArgs.Get("speed"))
}

// ─── Single-key brace form ────────────────────────────────────────────────────

func TestInputMacro_BracedSingleChar(t *testing.T) {
	t.Parallel()
	// {a} expands to "a" — single-char token without braces
	got := parseInputMacroArgs(t, "{a}")
	assert.Equal(t, []string{"a"}, got)
}

func TestInputMacro_BracedSpecialNoRepeat(t *testing.T) {
	t.Parallel()
	// {f1} passes through with braces (multi-char name)
	got := parseInputMacroArgs(t, "{f1}")
	assert.Equal(t, []string{"{f1}"}, got)
}

// ─── Repeat: key *N ──────────────────────────────────────────────────────────

func TestInputMacro_Repeat_SingleChar(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{a*5}")
	assert.Equal(t, []string{"a", "a", "a", "a", "a"}, got)
}

func TestInputMacro_Repeat_Special(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{enter*3}")
	assert.Equal(t, []string{"{enter}", "{enter}", "{enter}"}, got)
}

func TestInputMacro_Repeat_Combo(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{ctrl+c*3}")
	assert.Equal(t, []string{"{ctrl+c}", "{ctrl+c}", "{ctrl+c}"}, got)
}

func TestInputMacro_Repeat_One(t *testing.T) {
	t.Parallel()
	// *1 is valid and equals no repeat
	got := parseInputMacroArgs(t, "{a*1}")
	assert.Equal(t, []string{"a"}, got)
}

func TestInputMacro_Repeat_AsteriskViaQuote(t *testing.T) {
	t.Parallel()
	// Correct way to type 5 asterisks
	got := parseInputMacroArgs(t, `{"*"*5}`)
	assert.Equal(t, []string{"*", "*", "*", "*", "*"}, got)
}

func TestInputMacro_Repeat_InMixedSequence(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "a{enter*2}b")
	assert.Equal(t, []string{"a", "{enter}", "{enter}", "b"}, got)
}

// ─── Quoted literal ───────────────────────────────────────────────────────────

func TestInputMacro_QuotedLiteral_Basic(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, `{"hello"}`)
	assert.Equal(t, []string{"h", "e", "l", "l", "o"}, got)
}

func TestInputMacro_QuotedLiteral_WithRepeat(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, `{"hi"*3}`)
	assert.Equal(t, []string{"h", "i", "h", "i", "h", "i"}, got)
}

func TestInputMacro_QuotedLiteral_EscapedQuote(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, `{"say \"hi\""}`)
	assert.Equal(t, []string{"s", "a", "y", " ", `"`, "h", "i", `"`}, got)
}

func TestInputMacro_QuotedLiteral_AsteriskIsLiteral(t *testing.T) {
	t.Parallel()
	// Inside quotes, * is a literal character to type
	got := parseInputMacroArgs(t, `{"a*b"}`)
	assert.Equal(t, []string{"a", "*", "b"}, got)
}

func TestInputMacro_QuotedLiteral_PlusIsLiteral(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, `{"a+b"}`)
	assert.Equal(t, []string{"a", "+", "b"}, got)
}

func TestInputMacro_QuotedLiteral_Empty(t *testing.T) {
	t.Parallel()
	// {""}  — empty quoted literal produces no tokens
	got := parseInputMacroArgs(t, `{""}`)
	assert.Empty(t, got)
}

// ─── text: verb ───────────────────────────────────────────────────────────────

func TestInputMacro_TextVerb_Basic(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{text:hi}")
	assert.Equal(t, []string{"h", "i"}, got)
}

func TestInputMacro_TextVerb_WithRepeat(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{text:ab*3}")
	assert.Equal(t, []string{"a", "b", "a", "b", "a", "b"}, got)
}

func TestInputMacro_TextVerb_AsteriskInContent(t *testing.T) {
	t.Parallel()
	// {text:a*b} — *b is not a valid repeat (b is not a number) so * is literal
	got := parseInputMacroArgs(t, "{text:a*b}")
	assert.Equal(t, []string{"a", "*", "b"}, got)
}

func TestInputMacro_QuotedAndTextVerbAreEquivalent(t *testing.T) {
	t.Parallel()
	quote := parseInputMacroArgs(t, `{"hello"*2}`)
	verb := parseInputMacroArgs(t, "{text:hello*2}")
	assert.Equal(t, quote, verb)
}

// ─── delay: pass-through ─────────────────────────────────────────────────────

func TestInputMacro_Delay_PassThrough(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{delay:500}")
	assert.Equal(t, []string{"{delay:500}"}, got)
}

func TestInputMacro_Delay_HumanDuration(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{delay:1s}")
	assert.Equal(t, []string{"{delay:1s}"}, got)
}

func TestInputMacro_Delay_InSequence(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "a{delay:100}b")
	assert.Equal(t, []string{"a", "{delay:100}", "b"}, got)
}

// ─── Hold verbs and sigils ────────────────────────────────────────────────────

func TestInputMacro_HoldVerb_Press(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{press:a}")
	assert.Equal(t, []string{"{press:a}"}, got)
}

func TestInputMacro_HoldVerb_Release(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{release:a}")
	assert.Equal(t, []string{"{release:a}"}, got)
}

func TestInputMacro_HoldVerb_Hold(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{hold:a:500}")
	assert.Equal(t, []string{"{hold:a:500}"}, got)
}

func TestInputMacro_Sigil_Down(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{_a}")
	assert.Equal(t, []string{"{_a}"}, got)
}

func TestInputMacro_Sigil_Up(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{^a}")
	assert.Equal(t, []string{"{^a}"}, got)
}

func TestInputMacro_Sigil_Hold(t *testing.T) {
	t.Parallel()
	got := parseInputMacroArgs(t, "{~a:500}")
	assert.Equal(t, []string{"{~a:500}"}, got)
}

func TestInputMacro_HoldWhile_Sequence(t *testing.T) {
	t.Parallel()
	// {_right}bbb{^right} — hold right, tap b three times, release right
	got := parseInputMacroArgs(t, "{_right}bbb{^right}")
	assert.Equal(t, []string{"{_right}", "b", "b", "b", "{^right}"}, got)
}

// ─── Safety caps ─────────────────────────────────────────────────────────────

func TestInputMacro_RepeatCap(t *testing.T) {
	t.Parallel()
	// *1001 exceeds InputMacroMaxRepeat (1000)
	p := zapscript.NewParser("**input.keyboard:{a*1001}")
	_, err := p.ParseScript()
	require.Error(t, err)
	assert.ErrorIs(t, err, zapscript.ErrInputMacroRepeatTooLarge, "want ErrInputMacroRepeatTooLarge, got: %v", err)
}

func TestInputMacro_RepeatAtMax(t *testing.T) {
	t.Parallel()
	// *1000 is exactly at the cap — should succeed
	p := zapscript.NewParser("**input.keyboard:{a*1000}")
	script, err := p.ParseScript()
	require.NoError(t, err)
	assert.Len(t, script.Cmds[0].Args, 1000)
}

func TestInputMacro_TotalKeysCap(t *testing.T) {
	t.Parallel()
	// 6 reps × 1000 = 6000 > InputMacroMaxKeys (5000)
	p := zapscript.NewParser("**input.keyboard:{a*1000}{b*1000}{c*1000}{d*1000}{e*1000}{f*1000}")
	_, err := p.ParseScript()
	require.Error(t, err)
	assert.ErrorIs(t, err, zapscript.ErrInputMacroTooLong, "want ErrInputMacroTooLong, got: %v", err)
}

func TestInputMacro_TotalKeysAtMax(t *testing.T) {
	t.Parallel()
	// 5 × 1000 = 5000 exactly at the cap — should succeed
	p := zapscript.NewParser("**input.keyboard:{a*1000}{b*1000}{c*1000}{d*1000}{e*1000}")
	script, err := p.ParseScript()
	require.NoError(t, err)
	assert.Len(t, script.Cmds[0].Args, 5000)
}

func TestInputMacro_LiteralCharsCap(t *testing.T) {
	t.Parallel()
	// Plain chars outside braces also count toward the total
	// 5001 'a' chars should exceed the cap
	p := zapscript.NewParser("**input.keyboard:" + strings.Repeat("a", 5001))
	_, err := p.ParseScript()
	require.Error(t, err)
	assert.ErrorIs(t, err, zapscript.ErrInputMacroTooLong)
}

// ─── input.text raw mode ─────────────────────────────────────────────────────

func TestInputText_Basic(t *testing.T) {
	t.Parallel()
	got := parseInputTextArgs(t, "hello")
	assert.Equal(t, []string{"h", "e", "l", "l", "o"}, got)
}

func TestInputText_BracesAreLiteral(t *testing.T) {
	t.Parallel()
	// {} syntax is NOT interpreted in input.text
	got := parseInputTextArgs(t, "{enter}")
	assert.Equal(t, []string{"{", "e", "n", "t", "e", "r", "}"}, got)
}

func TestInputText_AsteriskIsLiteral(t *testing.T) {
	t.Parallel()
	got := parseInputTextArgs(t, "a*5")
	assert.Equal(t, []string{"a", "*", "5"}, got)
}

func TestInputText_QuestionMarkIsLiteral(t *testing.T) {
	t.Parallel()
	// ? is NOT parsed as adv-arg start in raw mode
	got := parseInputTextArgs(t, "what?")
	assert.Equal(t, []string{"w", "h", "a", "t", "?"}, got)
}

func TestInputText_URLIsLiteral(t *testing.T) {
	t.Parallel()
	// A URL with ? should type literally, not trigger adv-arg parsing
	got := parseInputTextArgs(t, "https://x.com/search?q=foo")
	require.Len(t, got, 26)
	// spot-check the ? character
	assert.Equal(t, "?", got[20])
}

func TestInputText_NewlineMapsToEnter(t *testing.T) {
	t.Parallel()
	got := parseInputTextArgs(t, "a\nb")
	assert.Equal(t, []string{"a", "{enter}", "b"}, got)
}

func TestInputText_TabMapsToTab(t *testing.T) {
	t.Parallel()
	got := parseInputTextArgs(t, "a\tb")
	assert.Equal(t, []string{"a", "{tab}", "b"}, got)
}

func TestInputText_NoAdvArgs(t *testing.T) {
	t.Parallel()
	// input.text has no adv-args; a trailing ? is typed literally
	p := zapscript.NewParser("**input.text:hello?speed=50")
	script, err := p.ParseScript()
	require.NoError(t, err)
	cmd := script.Cmds[0]
	// All chars including ? s p e e d = 5 0 are args
	assert.Contains(t, cmd.Args, "?")
	assert.Empty(t, cmd.AdvArgs.Get("speed"))
}

func TestInputText_CapEnforced(t *testing.T) {
	t.Parallel()
	p := zapscript.NewParser("**input.text:" + strings.Repeat("a", 5001))
	_, err := p.ParseScript()
	require.Error(t, err)
	assert.ErrorIs(t, err, zapscript.ErrInputMacroTooLong)
}

func TestInputText_EmptyArg(t *testing.T) {
	t.Parallel()
	p := zapscript.NewParser("**input.text:")
	script, err := p.ParseScript()
	require.NoError(t, err)
	assert.Empty(t, script.Cmds[0].Args)
}

// ─── Error cases ─────────────────────────────────────────────────────────────

func TestInputMacro_EmptyBraces(t *testing.T) {
	t.Parallel()
	p := zapscript.NewParser("**input.keyboard:{}")
	_, err := p.ParseScript()
	require.Error(t, err)
	assert.ErrorIs(t, err, zapscript.ErrUnmatchedInputMacroExt)
}

func TestInputMacro_UnclosedBrace(t *testing.T) {
	t.Parallel()
	p := zapscript.NewParser("**input.keyboard:{enter")
	_, err := p.ParseScript()
	require.Error(t, err)
	assert.ErrorIs(t, err, zapscript.ErrUnmatchedInputMacroExt)
}

func TestInputMacro_UnclosedQuotedLiteral(t *testing.T) {
	t.Parallel()
	// {"unclosed — EOF reached inside brace content before closing }
	// parseInputMacroExtContent returns ErrUnmatchedInputMacroExt on EOF
	p := zapscript.NewParser(`**input.keyboard:{"unclosed`)
	_, err := p.ParseScript()
	require.Error(t, err)
	assert.ErrorIs(t, err, zapscript.ErrUnmatchedInputMacroExt)
}

func TestInputMacro_EmptyKeyAfterRepeat(t *testing.T) {
	t.Parallel()
	// {*5} — content "*5" → name="" after removing *5 → ErrInputMacroEmptyKey
	p := zapscript.NewParser("**input.keyboard:{*5}")
	_, err := p.ParseScript()
	require.Error(t, err)
	assert.ErrorIs(t, err, zapscript.ErrInputMacroEmptyKey)
}
