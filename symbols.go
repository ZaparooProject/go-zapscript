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

import "errors"

var (
	ErrUnexpectedEOF          = errors.New("unexpected end of file")
	ErrInvalidCmdName         = errors.New("invalid characters in command name")
	ErrInvalidAdvArgName      = errors.New("invalid characters in advanced arg name")
	ErrEmptyCmdName           = errors.New("command name is empty")
	ErrEmptyZapScript         = errors.New("script is empty")
	ErrUnmatchedQuote         = errors.New("unmatched quote")
	ErrInvalidJSON            = errors.New("invalid JSON argument")
	ErrUnmatchedInputMacroExt = errors.New("unmatched input macro extension")
	ErrUnmatchedExpression    = errors.New("unmatched expression")
	ErrBadExpressionReturn    = errors.New("expression return type not supported")
)

const (
	SymCmdStart            = '*'
	SymCmdSep              = '|'
	SymEscapeSeq           = '^'
	SymArgStart            = ':'
	SymArgSep              = ','
	SymArgDoubleQuote      = '"'
	SymArgSingleQuote      = '\''
	SymAdvArgStart         = '?'
	SymAdvArgSep           = '&'
	SymAdvArgEq            = '='
	SymJSONStart           = '{'
	SymJSONEnd             = '}'
	SymJSONEscapeSeq       = '\\'
	SymJSONString          = '"'
	SymInputMacroEscapeSeq = '\\'
	SymInputMacroExtStart  = '{'
	SymInputMacroExtEnd    = '}'
	SymExpressionStart     = '['
	SymExpressionEnd       = ']'
	SymMediaTitleStart     = '@'
	SymMediaTitleSep       = '/'
	SymTagAnd              = '+'
	SymTagNot              = '-'
	SymTagOr               = '~'
	TokExpStart            = "\uE000"
	TokExprEnd             = "\uE001"
)

var eof = rune(0)

func isCmdName(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '.'
}

func isAdvArgName(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isInputMacroCmd(name string) bool {
	switch name {
	case ZapScriptCmdInputKeyboard, ZapScriptCmdInputGamepad:
		return true
	default:
		return false
	}
}
