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
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// AdvArgs is a wrapper around raw advanced arguments that enforces type-safe access.
// Direct map access is not allowed; use the getter/setter methods for pre-parse operations.
type AdvArgs struct {
	raw map[string]string
}

func NewAdvArgs(m map[string]string) AdvArgs {
	return AdvArgs{raw: m}
}

func (a AdvArgs) Get(key Key) string {
	return a.raw[string(key)]
}

// With returns a new AdvArgs with the key set to value. Does not mutate the receiver.
func (a AdvArgs) With(key Key, value string) AdvArgs {
	newMap := make(map[string]string, len(a.raw)+1)
	for k, v := range a.raw {
		newMap[k] = v
	}
	newMap[string(key)] = value
	return AdvArgs{raw: newMap}
}

func (a AdvArgs) GetWhen() (string, bool) {
	v, ok := a.raw[string(KeyWhen)]
	return v, ok
}

func (a AdvArgs) IsEmpty() bool {
	return len(a.raw) == 0
}

func (a AdvArgs) Range(fn func(key Key, value string) bool) {
	for k, v := range a.raw {
		if !fn(Key(k), v) {
			return
		}
	}
}

func (a AdvArgs) Raw() map[string]string {
	return a.raw
}

func (a AdvArgs) MarshalJSON() ([]byte, error) {
	if a.raw == nil {
		return []byte("null"), nil
	}
	b, err := json.Marshal(a.raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AdvArgs: %w", err)
	}
	return b, nil
}

func (a *AdvArgs) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &a.raw); err != nil {
		return fmt.Errorf("failed to unmarshal AdvArgs: %w", err)
	}
	return nil
}

type Command struct {
	AdvArgs AdvArgs
	Name    string
	Args    []string
}

// argNeedsQuoting returns true if the arg contains characters that require
// double-quoting to be safely represented in ZapScript.
func argNeedsQuoting(s string) bool {
	for _, ch := range s {
		switch ch {
		case SymArgSep, SymArgStart, SymAdvArgStart, SymAdvArgSep,
			SymAdvArgEq, SymArgDoubleQuote, SymArgSingleQuote, SymCmdSep,
			SymEscapeSeq, SymCmdStart, SymExpressionStart, SymTraitsStart,
			SymJSONStart, '\n', '\r', '\t':
			return true
		}
	}
	return false
}

// escapeArg re-escapes control characters using ZapScript escape sequences
// and wraps the arg in double quotes.
func escapeArg(s string) string {
	var b strings.Builder
	_, _ = b.WriteRune('"')
	for _, ch := range s {
		switch ch {
		case '"':
			_, _ = b.WriteRune(SymEscapeSeq)
			_, _ = b.WriteRune('"')
		case '\n':
			_, _ = b.WriteRune(SymEscapeSeq)
			_, _ = b.WriteRune('n')
		case '\r':
			_, _ = b.WriteRune(SymEscapeSeq)
			_, _ = b.WriteRune('r')
		case '\t':
			_, _ = b.WriteRune(SymEscapeSeq)
			_, _ = b.WriteRune('t')
		case SymEscapeSeq:
			_, _ = b.WriteRune(SymEscapeSeq)
			_, _ = b.WriteRune(SymEscapeSeq)
		case SymExpressionStart:
			_, _ = b.WriteRune(SymEscapeSeq)
			_, _ = b.WriteRune(SymExpressionStart)
		default:
			_, _ = b.WriteRune(ch)
		}
	}
	_, _ = b.WriteRune('"')
	return b.String()
}

// String returns the canonical ZapScript representation of the command.
// The output is valid ZapScript that can be re-parsed to produce an
// equivalent Command.
func (c Command) String() string {
	var b strings.Builder
	_, _ = b.WriteString("**")
	_, _ = b.WriteString(c.Name)

	if len(c.Args) > 0 {
		_, _ = b.WriteRune(SymArgStart)

		switch {
		case isInputMacroCmd(normalizeCmdName(c.Name)):
			// Input macro commands concatenate args directly
			for _, arg := range c.Args {
				if len(arg) > 1 && rune(arg[0]) == SymInputMacroExtStart &&
					rune(arg[len(arg)-1]) == SymInputMacroExtEnd {
					_, _ = b.WriteString(arg)
				} else {
					for _, ch := range arg {
						switch ch {
						case SymInputMacroEscapeSeq:
							_, _ = b.WriteRune(SymInputMacroEscapeSeq)
							_, _ = b.WriteRune(ch)
						case SymAdvArgStart, SymExpressionStart, SymInputMacroExtStart, SymCmdSep:
							_, _ = b.WriteRune(SymInputMacroEscapeSeq)
							_, _ = b.WriteRune(ch)
						default:
							_, _ = b.WriteRune(ch)
						}
					}
				}
			}
		case isInputRawCmd(normalizeCmdName(c.Name)):
			// Raw text: reverse the parseInputRawArg mappings so the output re-parses
			// to the same args. {enter} → newline, {tab} → tab; all other args are
			// single chars written as-is.
			for _, arg := range c.Args {
				switch arg {
				case "{enter}":
					_, _ = b.WriteRune('\n')
				case "{tab}":
					_, _ = b.WriteRune('\t')
				default:
					_, _ = b.WriteString(arg)
				}
			}
		default:
			for i, arg := range c.Args {
				if i > 0 {
					_, _ = b.WriteRune(SymArgSep)
				}
				if arg == "" || argNeedsQuoting(arg) {
					_, _ = b.WriteString(escapeArg(arg))
				} else {
					_, _ = b.WriteString(arg)
				}
			}
		}
	}

	if !c.AdvArgs.IsEmpty() {
		_, _ = b.WriteRune(SymAdvArgStart)

		// Collect and sort keys for deterministic output
		var keys []string
		c.AdvArgs.Range(func(key Key, _ string) bool {
			keys = append(keys, string(key))
			return true
		})
		sort.Strings(keys)

		for i, key := range keys {
			if i > 0 {
				_, _ = b.WriteRune(SymAdvArgSep)
			}
			_, _ = b.WriteString(key)
			_, _ = b.WriteRune(SymAdvArgEq)
			value := c.AdvArgs.Get(Key(key))
			if argNeedsQuoting(value) {
				_, _ = b.WriteString(escapeArg(value))
			} else {
				_, _ = b.WriteString(value)
			}
		}
	}

	return b.String()
}

type Script struct {
	Traits map[string]any `json:"traits,omitempty"`
	Cmds   []Command      `json:"cmds"`
}

type PostArgPartType int

const (
	ArgPartTypeUnknown PostArgPartType = iota
	ArgPartTypeString
	ArgPartTypeExpression
)

type PostArgPart struct {
	Value string
	Type  PostArgPartType
}

type mediaTitleParseResult struct {
	advArgs    map[string]string
	rawContent string
	valid      bool
}

var (
	errUnreadNotPossible = errors.New("no rune available to unread")
	errRuneError         = errors.New("rune error")
)

// ScriptReader walks the input one rune at a time, forward only.
//
// The input is held as the caller's string and decoded in place. Reading
// through a buffer would copy the whole script and allocate a buffer for
// every parser, which is dead weight for input that is already in memory and
// is measurable when a caller parses the same short script several times.
type ScriptReader struct {
	src string
	// off is the byte offset of the next rune to read.
	off int
	// lastWidth is the byte width of the rune returned by the most recent
	// read, or 0 when there is nothing to unread.
	lastWidth int
	// pos counts runes consumed. Argument parsing compares it against a
	// saved position to tell whether a quote or brace opened an argument, so
	// it counts runes rather than bytes.
	pos int64
}

func NewParser(value string) *ScriptReader {
	return &ScriptReader{src: value}
}

func (sr *ScriptReader) read() (rune, error) {
	if sr.off >= len(sr.src) {
		sr.lastWidth = 0
		return eof, nil
	}
	ch, width := utf8.DecodeRuneInString(sr.src[sr.off:])
	sr.off += width
	sr.lastWidth = width
	sr.pos++
	return ch, nil
}

// unread steps back over the rune returned by the most recent read. Only that
// rune can be returned: a second unread, or one after a peek or at end of
// input, is a bug in the caller rather than a condition to recover from.
func (sr *ScriptReader) unread() error {
	if sr.lastWidth == 0 {
		return errUnreadNotPossible
	}
	sr.off -= sr.lastWidth
	sr.lastWidth = 0
	sr.pos--
	return nil
}

// remaining returns the text the reader has not consumed yet. It is a slice
// of the original input, not a copy.
func (sr *ScriptReader) remaining() string {
	return sr.src[sr.off:]
}

// consumeAll advances the reader to the end of the input, for the fast paths
// that return the remaining text verbatim instead of walking it.
func (sr *ScriptReader) consumeAll() {
	sr.pos += int64(utf8.RuneCountInString(sr.remaining()))
	sr.off = len(sr.src)
	sr.lastWidth = 0
}

func (sr *ScriptReader) peek() (rune, error) {
	// A peek invalidates the pending unread, matching the buffered reader
	// this replaced, so a caller cannot start relying on the two composing.
	sr.lastWidth = 0
	if sr.off >= len(sr.src) {
		return eof, nil
	}
	r, width := utf8.DecodeRuneInString(sr.src[sr.off:])
	// DecodeRuneInString reports invalid encoding as U+FFFD with a width of
	// one. A U+FFFD that was actually written in the input decodes to the
	// same rune but is three bytes wide, and is ordinary content.
	if r == utf8.RuneError && width <= 1 {
		return r, errRuneError
	}
	return r, nil
}

func (sr *ScriptReader) skip() error {
	_, err := sr.read()
	if err != nil {
		return err
	}
	return nil
}

func (sr *ScriptReader) checkEndOfCmd(ch rune) (bool, error) {
	if ch != SymCmdSep {
		return false, nil
	}

	next, err := sr.peek()
	if err != nil {
		return false, err
	}

	switch next {
	case eof:
		return true, nil
	case SymCmdSep:
		err := sr.skip()
		if err != nil {
			return false, err
		}
		return true, nil
	default:
		return false, nil
	}
}

func (sr *ScriptReader) parseEscapeSeq() (string, error) {
	ch, err := sr.read()
	if err != nil {
		return "", err
	}
	switch ch {
	case eof:
		return "", nil
	case 'n':
		return "\n", nil
	case 'r':
		return "\r", nil
	case 't':
		return "\t", nil
	case SymEscapeSeq:
		return string(SymEscapeSeq), nil
	case SymArgDoubleQuote:
		return string(SymArgDoubleQuote), nil
	case SymArgSingleQuote:
		return string(SymArgSingleQuote), nil
	default:
		return string(ch), nil
	}
}

func (sr *ScriptReader) parseQuotedArg(start rune) (string, error) {
	var arg strings.Builder

	for {
		ch, err := sr.read()
		if err != nil {
			return arg.String(), err
		} else if ch == eof {
			return arg.String(), ErrUnmatchedQuote
		}

		if ch == SymEscapeSeq {
			next, err := sr.parseEscapeSeq()
			if err != nil {
				return arg.String(), err
			}
			_, _ = arg.WriteString(next)
			continue
		} else if ch == SymExpressionStart {
			exprValue, err := sr.parseExpression()
			if err != nil {
				return arg.String(), err
			}
			_, _ = arg.WriteString(exprValue)
			continue
		}

		if ch == start {
			break
		}

		_, _ = arg.WriteRune(ch)
	}

	return arg.String(), nil
}
