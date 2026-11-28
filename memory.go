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
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

// MemoryEntry is one key and value held in a script's memory.
type MemoryEntry struct {
	Value any
	Key   string
}

// Memory is the changeable state a card carries in the last item of its
// script, written as $key=value pairs. Entries keep the order they were
// written in, so rewriting a card leaves its keys where the user put them.
//
// Values are int64, float64, bool, string, or a []any of those.
type Memory []MemoryEntry

// Get returns the value stored under key.
func (m Memory) Get(key string) (any, bool) {
	key = NormalizeTraitKey(key)
	for _, e := range m {
		if e.Key == key {
			return e.Value, true
		}
	}
	return nil, false
}

// Set stores value under key, replacing an existing value in place or adding
// the key at the end. The key must be valid trait key, and the value is
// normalized to a type the memory syntax can hold.
func (m *Memory) Set(key string, value any) error {
	if err := ValidateTraitKey(key); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidMemory, err)
	}
	norm, err := NormalizeMemoryValue(value)
	if err != nil {
		return err
	}
	key = NormalizeTraitKey(key)
	for i := range *m {
		if (*m)[i].Key == key {
			(*m)[i].Value = norm
			return nil
		}
	}
	*m = append(*m, MemoryEntry{Key: key, Value: norm})
	return nil
}

// Delete removes key and reports whether it was present.
func (m *Memory) Delete(key string) bool {
	key = NormalizeTraitKey(key)
	for i := range *m {
		if (*m)[i].Key == key {
			*m = append((*m)[:i], (*m)[i+1:]...)
			return true
		}
	}
	return false
}

// Clone returns a copy that shares nothing with m.
func (m Memory) Clone() Memory {
	if m == nil {
		return nil
	}
	out := make(Memory, len(m))
	for i, e := range m {
		out[i] = MemoryEntry{Key: e.Key, Value: cloneMemoryValue(e.Value)}
	}
	return out
}

// Map returns the memory as a map, for use as an expression environment.
func (m Memory) Map() map[string]any {
	out := make(map[string]any, len(m))
	for _, e := range m {
		out[e.Key] = cloneMemoryValue(e.Value)
	}
	return out
}

// MarshalJSON writes the memory as a JSON object in entry order.
func (m Memory) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	_ = buf.WriteByte('{')
	for i, e := range m {
		if i > 0 {
			_ = buf.WriteByte(',')
		}
		k, err := json.Marshal(e.Key)
		if err != nil {
			return nil, fmt.Errorf("marshal memory key: %w", err)
		}
		v, err := json.Marshal(e.Value)
		if err != nil {
			return nil, fmt.Errorf("marshal memory value: %w", err)
		}
		_, _ = buf.Write(k)
		_ = buf.WriteByte(':')
		_, _ = buf.Write(v)
	}
	_ = buf.WriteByte('}')
	return buf.Bytes(), nil
}

func cloneMemoryValue(v any) any {
	if list, ok := v.([]any); ok {
		out := make([]any, len(list))
		copy(out, list)
		return out
	}
	return v
}

// NormalizeMemoryValue converts value to the type memory stores it as: any
// integer becomes int64, any float becomes float64, and a list becomes []any.
// It rejects nil, NaN, infinities, strings that are not valid UTF-8 or hold a NUL, nested
// lists and other types, none of which the memory syntax can write.
func NormalizeMemoryValue(value any) (any, error) {
	return normalizeValue(value, true)
}

func normalizeValue(value any, allowList bool) (any, error) {
	switch v := value.(type) {
	case nil:
		return nil, fmt.Errorf("%w: nil", ErrInvalidMemoryValue)
	case string:
		if !utf8.ValidString(v) {
			return nil, fmt.Errorf("%w: string is not valid UTF-8", ErrInvalidMemoryValue)
		}
		// The reader treats NUL as the end of the input.
		if strings.ContainsRune(v, 0) {
			return nil, fmt.Errorf("%w: string contains a NUL character", ErrInvalidMemoryValue)
		}
		return v, nil
	case bool, int64:
		return v, nil
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint:
		if uint64(v) > math.MaxInt64 {
			return nil, fmt.Errorf("%w: %d is out of range", ErrInvalidMemoryValue, v)
		}
		return int64(v), nil
	case uint64:
		if v > math.MaxInt64 {
			return nil, fmt.Errorf("%w: %d is out of range", ErrInvalidMemoryValue, v)
		}
		return int64(v), nil
	case float32:
		return normalizeMemoryFloat(float64(v))
	case float64:
		return normalizeMemoryFloat(v)
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("%w: %T", ErrInvalidMemoryValue, value)
	}
	if !allowList {
		return nil, fmt.Errorf("%w: nested list", ErrInvalidMemoryValue)
	}
	out := make([]any, rv.Len())
	for i := range out {
		elem, err := normalizeValue(rv.Index(i).Interface(), false)
		if err != nil {
			return nil, err
		}
		out[i] = elem
	}
	return out, nil
}

func normalizeMemoryFloat(f float64) (any, error) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil, fmt.Errorf("%w: %v", ErrInvalidMemoryValue, f)
	}
	return f, nil
}

// FormatMemory writes memory in the syntax the parser reads, without the
// leading separator: $key=value pairs joined by spaces, in entry order.
// Parsing the result gives back the same memory.
func FormatMemory(m Memory) (string, error) {
	var b strings.Builder
	for i, e := range m {
		if err := ValidateTraitKey(e.Key); err != nil {
			return "", fmt.Errorf("%w: %w", ErrInvalidMemory, err)
		}
		if i > 0 {
			_ = b.WriteByte(' ')
		}
		_ = b.WriteByte(SymMemoryStart)
		_, _ = b.WriteString(NormalizeTraitKey(e.Key))
		_ = b.WriteByte(SymAdvArgEq)
		norm, err := NormalizeMemoryValue(e.Value)
		if err != nil {
			return "", fmt.Errorf("memory key %q: %w", e.Key, err)
		}
		if list, ok := norm.([]any); ok {
			_ = b.WriteByte(SymArrayStart)
			for j, elem := range list {
				if j > 0 {
					_ = b.WriteByte(SymArraySep)
				}
				formatMemoryScalar(&b, elem)
			}
			_ = b.WriteByte(SymArrayEnd)
			continue
		}
		formatMemoryScalar(&b, norm)
	}
	return b.String(), nil
}

// formatMemoryScalar writes a normalized scalar. Strings are always quoted so
// one that looks like a number or a bool keeps its type.
func formatMemoryScalar(b *strings.Builder, v any) {
	switch x := v.(type) {
	case bool:
		_, _ = b.WriteString(strconv.FormatBool(x))
	case int64:
		_, _ = b.WriteString(strconv.FormatInt(x, 10))
	case float64:
		s := strconv.FormatFloat(x, 'g', -1, 64)
		// A float with no point or exponent would read back as an integer.
		if !strings.ContainsAny(s, ".eE") {
			s += ".0"
		}
		_, _ = b.WriteString(s)
	case string:
		_ = b.WriteByte(SymArgDoubleQuote)
		for _, ch := range x {
			switch ch {
			case SymEscapeSeq, SymArgDoubleQuote, SymCmdSep:
				_ = b.WriteByte(SymEscapeSeq)
				_, _ = b.WriteRune(ch)
			case '\n':
				_, _ = b.WriteString("^n")
			case '\r':
				_, _ = b.WriteString("^r")
			case '\t':
				_, _ = b.WriteString("^t")
			default:
				_, _ = b.WriteRune(ch)
			}
		}
		_ = b.WriteByte(SymArgDoubleQuote)
	}
}

// ReplaceMemory returns text with its memory item replaced by m. Everything
// before the memory item is kept byte for byte. Text without a memory item
// gets one appended after a separator, and an empty m removes the item.
//
// Text holding a NUL character is refused.
// The result is parsed again before it is returned, and an error is returned
// if its commands, traits or memory differ from what was asked for.
func ReplaceMemory(text string, m Memory) (string, error) {
	// The reader takes a NUL for the end of the input, so where an appended
	// item would land in such a text can't be relied on.
	if strings.ContainsRune(text, 0) {
		return "", fmt.Errorf("%w: script contains a NUL character", ErrInvalidMemory)
	}

	sr := NewParser(text)
	orig, err := sr.ParseScript()
	if err != nil {
		return "", fmt.Errorf("parse script: %w", err)
	}

	if orig.Memory == nil && len(m) == 0 {
		return text, nil
	}

	formatted, err := FormatMemory(m)
	if err != nil {
		return "", err
	}

	var head string
	if orig.Memory != nil {
		head = text[:sr.memoryStart]
	} else {
		head = text
	}

	var out string
	if len(m) == 0 {
		out = trimTrailingSeparator(head)
	} else {
		// Whitespace before a separator can be part of the item it follows,
		// so the text is kept as it is and only a separator is added.
		trimmed := strings.TrimRight(head, " \t\r\n")
		switch {
		case orig.Memory != nil, trimmed == "", endsWithSeparator(trimmed):
			out = head + formatted
		default:
			out = head + "||" + formatted
		}
	}

	if err := checkReplacedMemory(orig, out, m); err != nil {
		return "", err
	}
	return out, nil
}

// trimTrailingSeparator removes the separator, and any whitespace after it,
// that led into a removed memory item.
func trimTrailingSeparator(head string) string {
	s := strings.TrimRight(head, " \t\r\n")
	if endsWithSeparator(s) {
		return strings.TrimSuffix(s, "||")
	}
	return head
}

func endsWithSeparator(s string) bool {
	return strings.HasSuffix(s, "||") && !strings.HasSuffix(s, "^||")
}

func checkReplacedMemory(orig Script, out string, m Memory) error {
	if out == "" {
		if len(orig.Cmds) == 0 && len(orig.Traits) == 0 {
			return nil
		}
		return fmt.Errorf("%w: rewritten script is empty", ErrInvalidMemory)
	}
	got, err := NewParser(out).ParseScript()
	if err != nil {
		return fmt.Errorf("%w: rewritten script does not parse: %w", ErrInvalidMemory, err)
	}
	want, err := normalizedMemory(m)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(got.Memory, want) {
		return fmt.Errorf("%w: rewritten memory does not read back", ErrInvalidMemory)
	}
	if !reflect.DeepEqual(got.Cmds, orig.Cmds) || !reflect.DeepEqual(got.Traits, orig.Traits) {
		return fmt.Errorf("%w: rewriting memory changed the script", ErrInvalidMemory)
	}
	return nil
}

func normalizedMemory(m Memory) (Memory, error) {
	if len(m) == 0 {
		return nil, nil
	}
	out := make(Memory, 0, len(m))
	for _, e := range m {
		if err := out.Set(e.Key, e.Value); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// parseMemorySyntax reads a memory item: $key=value pairs separated by
// whitespace, up to the end of the input or a command separator. The leading
// $ has been read and a letter follows it.
func (sr *ScriptReader) parseMemorySyntax() (Memory, error) {
	var mem Memory
	for {
		key, err := sr.readMemoryKey()
		if err != nil {
			return nil, err
		}

		next, err := sr.peek()
		if err != nil {
			return nil, err
		}
		if next != SymAdvArgEq {
			return nil, fmt.Errorf("%w: key %q needs =value", ErrInvalidMemory, key)
		}
		if skipErr := sr.skip(); skipErr != nil {
			return nil, skipErr
		}

		value, _, err := sr.parseTraitValue(SymMemoryStart)
		if err != nil {
			return nil, err
		}
		if setErr := mem.Set(key, value); setErr != nil {
			return nil, setErr
		}

		done, err := sr.afterMemoryValue(key)
		if err != nil {
			return nil, err
		}
		if done {
			return mem, nil
		}
	}
}

func (sr *ScriptReader) readMemoryKey() (string, error) {
	var key strings.Builder
	for {
		next, err := sr.peek()
		if err != nil {
			return "", err
		}
		if !isAdvArgName(next) {
			break
		}
		ch, err := sr.read()
		if err != nil {
			return "", err
		}
		_, _ = key.WriteRune(ch)
	}
	return NormalizeTraitKey(key.String()), nil
}

// afterMemoryValue moves past the whitespace after a value. It reports done at
// the end of the input or a command separator, and returns with the reader on
// the next key when another $ follows.
func (sr *ScriptReader) afterMemoryValue(key string) (bool, error) {
	for {
		ch, err := sr.read()
		if err != nil {
			return false, err
		}
		switch {
		case ch == eof:
			return true, nil
		case isWhitespace(ch):
			continue
		case ch == SymMemoryStart:
			next, err := sr.peek()
			if err != nil {
				return false, err
			}
			if !isAdvArgNameStart(next) {
				return false, fmt.Errorf("%w: a key must start with a letter", ErrInvalidMemory)
			}
			return false, nil
		case ch == SymCmdSep:
			eoc, err := sr.checkEndOfCmd(ch)
			if err != nil {
				return false, err
			}
			if eoc {
				return true, nil
			}
			return false, fmt.Errorf("%w: unexpected %q after %q", ErrInvalidMemory, ch, key)
		default:
			return false, fmt.Errorf("%w: unexpected %q after %q", ErrInvalidMemory, ch, key)
		}
	}
}

// ParseMemoryValue reads one value by the rules a memory item uses: a quoted
// string, a [a,b] list, true or false, an integer, a float, or else the text
// as a string. Unlike a value inside a memory item, unquoted text may hold
// spaces, because nothing follows it. The result is normalized as
// NormalizeMemoryValue does.
func ParseMemoryValue(s string) (any, error) {
	s = strings.TrimSpace(s)
	if s == "" || (s[0] != SymArgDoubleQuote && s[0] != SymArgSingleQuote && s[0] != SymArrayStart) {
		return NormalizeMemoryValue(inferType(s, false))
	}

	sr := NewParser(s)
	value, _, err := sr.parseTraitValue(SymMemoryStart)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidMemoryValue, err)
	}
	if rest := strings.TrimSpace(s[sr.off:]); rest != "" {
		return nil, fmt.Errorf("%w: unexpected %q after value", ErrInvalidMemoryValue, rest)
	}
	return NormalizeMemoryValue(value)
}
