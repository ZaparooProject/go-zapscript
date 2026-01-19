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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

type ScriptReader struct {
	r   *bufio.Reader
	pos int64
}

func NewParser(value string) *ScriptReader {
	return &ScriptReader{
		r: bufio.NewReader(bytes.NewReader([]byte(value))),
	}
}

func (sr *ScriptReader) read() (rune, error) {
	ch, _, err := sr.r.ReadRune()
	if errors.Is(err, io.EOF) {
		return eof, nil
	} else if err != nil {
		return eof, fmt.Errorf("failed to read rune: %w", err)
	}
	sr.pos++
	return ch, nil
}

func (sr *ScriptReader) unread() error {
	err := sr.r.UnreadRune()
	if err != nil {
		return fmt.Errorf("failed to unread rune: %w", err)
	}
	sr.pos--
	return nil
}

func (sr *ScriptReader) peek() (rune, error) {
	for peekBytes := 4; peekBytes > 0; peekBytes-- {
		b, err := sr.r.Peek(peekBytes)
		if err == nil {
			r, _ := utf8.DecodeRune(b)
			if r == utf8.RuneError {
				return r, errors.New("rune error")
			}
			return r, nil
		}
	}
	return eof, nil
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
	arg := ""

	for {
		ch, err := sr.read()
		if err != nil {
			return arg, err
		} else if ch == eof {
			return arg, ErrUnmatchedQuote
		}

		if ch == SymEscapeSeq {
			next, err := sr.parseEscapeSeq()
			if err != nil {
				return arg, err
			}
			arg += next
			continue
		} else if ch == SymExpressionStart {
			exprValue, err := sr.parseExpression()
			if err != nil {
				return arg, err
			}
			arg += exprValue
			continue
		}

		if ch == start {
			break
		}

		arg += string(ch)
	}

	return arg, nil
}
