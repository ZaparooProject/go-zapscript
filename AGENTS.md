# CLAUDE.md - go-zapscript

A README for AI coding agents working on go-zapscript.

## Project Overview

go-zapscript is a Go parser and evaluator library for ZapScript, a custom domain-specific language designed for the Zaparoo Project. It provides functionality for parsing and executing commands related to game launching, playlist management, media playback, input simulation, and system control through physical tokens in gaming emulation environments.

**Tech Stack**: Go 1.24+, expr-lang/expr (expression evaluation), testify (testing), rapid (property-based testing)

**Testing Standards**: Comprehensive test coverage with unit tests, property-based tests, and fuzz tests

## Development Guidelines

### Do

- **Write tests for all new code** - comprehensive coverage required
- **Use `task lint-fix`** to resolve all linting and formatting issues
- **Keep diffs small and focused** - one concern per change
- **Use parallel tests** with `t.Parallel()` for all test functions
- **Reference existing patterns** before writing new code - consistency matters
- **Handle all errors explicitly** - use golangci-lint's error handling checks
- **Keep functions small** and focused on single responsibility

### Don't

- ❌ Skip writing tests for new features or bug fixes
- ❌ Make large, unfocused diffs - keep changes small and targeted
- ❌ Add new dependencies without discussion
- ❌ Write comments that restate what the code does - comments should explain *why*, not *what*
- ❌ Use standard `log` or `fmt.Println` in library code
- ❌ Skip the `t.Parallel()` call in test functions

### Code Quality

- **Use Go 1.24+** with Go modules enabled
- **Handle all errors explicitly** - use golangci-lint's error handling checks
- **Use explicit returns** in functions longer than 5 lines (avoid naked returns)
- **Keep functions small** and focused on single responsibility
- **Line length**: 120 characters max

### Testing

**The goal is useful tests, not coverage metrics.** High coverage means nothing if tests don't catch bugs.

#### What to Test
- **Parser behavior** - valid inputs produce correct output
- **Edge cases** - empty inputs, boundary conditions, escape sequences
- **Error paths** - invalid inputs return appropriate errors
- **Expression evaluation** - dynamic expressions evaluate correctly

#### How to Test
- **Use property-based tests** for parser robustness (see `parser_property_test.go`)
- **Use fuzz tests** to catch panics and crashes (see `parser_fuzz_test.go`)
- **Use table-driven tests** for multiple scenarios
- **Use `t.Parallel()`** in all test functions
- **Use `cmp.Diff`** from google/go-cmp for detailed assertion output

#### Test Quality Checklist
- Would this test catch a real bug?
- Does it test behavior the user/caller cares about?
- Is this testing MY code or a library's code?

## Commands

### File-scoped checks (preferred for speed)

```bash
# Test a specific file or function
go test -run TestSpecificFunction ./...
go test -race ./...

# Lint and format - ALWAYS prefer task commands
task lint-fix                          # PREFERRED: Full lint with auto-fixes
task lint                              # Run linting without fixes

# Run single test by name
go test -run TestParseScript ./...

# Run tests with verbose output
go test -v ./...

# Run fuzz tests
go test -fuzz=FuzzParseScript -fuzztime=30s ./...
```

### Project-wide commands

```bash
# Full test suite with race detection
task test

# Full lint with auto-fixes
task lint-fix

# Run all linting checks
task lint
```

## When Stuck

**Don't guess - ask for help or gather more information first.**

- **Ask clarifying questions** - Get requirements clear before coding
- **Propose a plan first** - Outline approach, then implement
- **Reference existing patterns** - Check similar code in the codebase for consistency
- **Look at git history** - `git log -p filename` shows how code evolved
- **Keep scope focused** - Small, well-defined changes are easier to review and debug

**Remember**: It's better to ask than to make incorrect assumptions. The project values correctness over speed.

## Project Structure

```
go-zapscript/
├── parser.go           # Main parsing engine with ScriptReader
├── reader.go           # ScriptReader input handling
├── symbols.go          # Parser symbols and character constants
├── traits.go           # Traits syntax parsing (#key=value)
├── types.go            # Type definitions, constants, and argument structures
├── models.go           # JSON-serializable data structures for scripts
├── advargs.go          # Type-safe advanced argument wrapper
├── exprenv.go          # Expression evaluation environment types
├── parser_test.go      # Core parsing tests
├── parser_coverage_test.go    # Extended coverage tests
├── parser_media_title_test.go # Media title syntax tests
├── parser_traits_test.go      # Traits syntax tests
├── parser_property_test.go    # Property-based tests (rapid)
├── parser_fuzz_test.go        # Fuzz tests
├── parser_internal_test.go    # Internal implementation tests
├── exprenv_test.go            # Expression environment tests
├── Taskfile.yml        # Build and development tasks
└── .golangci.yml       # Linting configuration
```

## ZapScript Syntax

### Command Structure

```
**command:arg1,arg2?advkey=advvalue&advkey2=advvalue2
```

- `**` or `*` - Command prefix (required)
- `command` - Command name
- `:` - Argument separator (before first arg)
- `,` - Multiple argument separator
- `?` - Advanced argument prefix
- `&` - Advanced argument separator
- `=` - Advanced argument assignment
- `|` or `||` - Command chain separator

### Expression Placeholders

```
[[expression]]
```

Expressions are evaluated at runtime using expr-lang. The environment includes device info, active media, last scanned token, and launching context.

### Media Title Syntax

```
@system/title
```

Direct media launch by system and title.

### Escape Sequences

- `^n` - Newline
- `^t` - Tab
- `^r` - Carriage return
- `^^` - Literal caret

### Tag Filters

- `+tag` - AND (must have tag)
- `-tag` - NOT (must not have tag)
- `~tag` - OR (any of these tags)

### Traits Syntax

Traits provide metadata key-value pairs using `#` prefix:

```
#key=value #key2=value2
```

- `#key=value` - Key-value pair
- `#flag` - Boolean shorthand (sets value to `true`)
- `#key="quoted value"` - Quoted string (preserves spaces)
- `#key=[a,b,c]` - Array values

**Type Inference** (unquoted values):
- `true`/`false` → boolean
- `123` → integer
- `3.14` → float
- Everything else → string
- Quoted values are always strings

**Examples**:
```
#favorite                     → {"favorite": true}
#count=5                      → {"count": 5}
#name="My Game"               → {"name": "My Game"}
#tags=[action,rpg,indie]      → {"tags": ["action", "rpg", "indie"]}
#enabled=true #priority=1     → {"enabled": true, "priority": 1}
```

**Fallback Behavior**: Invalid trait syntax (e.g., `#my-trait` with hyphen) falls back to auto-launch content.

## Key API Entry Points

```go
// Create a parser instance
parser := zapscript.NewParser(scriptText)

// Parse commands from script
script, err := parser.ParseScript()

// Parse expression placeholders
result, err := parser.ParseExpressions("text with [[expr]]")

// Evaluate expressions with environment
result, err := parser.EvalExpressions(envStruct)
```

## Supported Commands

40+ commands including:

- **Launch**: `launch`, `launch.system`, `launch.random`, `launch.search`, `launch.title`
- **Playlist**: `playlist.play`, `playlist.stop`, `playlist.next`, `playlist.previous`, `playlist.goto`, `playlist.pause`, `playlist.load`, `playlist.open`
- **System**: `execute`, `delay`, `evaluate`, `stop`, `echo`
- **MiSTer**: `mister.ini`, `mister.core`, `mister.script`, `mister.mgl`
- **HTTP**: `http.get`, `http.post`
- **Input**: `input.keyboard`, `input.gamepad`, `input.coinp1`, `input.coinp2`
- **UI**: `ui.notice`, `ui.picker`
- **Metadata**: `traits` (parsed from `#key=value` syntax)

## Good Examples to Follow

**Copy these patterns for new code:**

- **Parser Methods**: `parser.go` - ScriptReader parsing methods
- **Traits Parsing**: `traits.go` - Type inference, arrays, quoted values
- **Type Definitions**: `types.go` - Constant and type patterns
- **Table-Driven Tests**: `parser_test.go` - Test organization pattern
- **Traits Tests**: `parser_traits_test.go` - Comprehensive trait syntax tests
- **Property Tests**: `parser_property_test.go` - Property-based testing with rapid
- **Fuzz Tests**: `parser_fuzz_test.go` - Fuzz testing pattern

## Code Style & Standards

Following golangci-lint configuration in `.golangci.yml`:

- **Line length**: 120 characters max
- **Error handling**: All errors must be checked (errcheck, wrapcheck)
- **Imports**: Grouped and sorted with gci formatter
- **Formatting**: Use gofumpt (stricter than gofmt)
- **JSON tags**: snake_case (enforced by tagliatelle)
- **License headers**: Apache 2.0 required on all files (enforced by goheader)
- **Nil checks**: Comprehensive (nilnil, nilerr rules)

## Git & Commit Guidelines

### Commit message format

Use **Conventional Commits** format:

```
<type>[optional scope]: <description>
```

**Types**:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation only
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Other changes

**Examples**:

```bash
git commit -m "feat: add support for new command type"
git commit -m "fix: resolve escape sequence parsing issue"
git commit -m "test: add property tests for expression evaluation"
```

### Before committing

**ALWAYS run these commands in order:**

```bash
# 1. Fix all linting and formatting issues (REQUIRED)
task lint-fix

# 2. Run all tests with race detection (REQUIRED)
task test
```

### Commit checklist

- [ ] Tests pass (`task test`)
- [ ] Linting passes (`task lint-fix`)
- [ ] License headers on new files
- [ ] Commit message follows Conventional Commits format
- [ ] Diff is small and focused on one concern
- [ ] No commented-out code or debug prints

## Safety & Permissions

### Allowed without asking:

- Read any files in the repository
- Run tests: `go test ./...` or `task test`
- Run `task lint-fix` to fix linting and formatting issues
- Run `task lint` to check linting
- View git history: `git log`, `git diff`

### Ask before:

- Installing new Go dependencies
- Running `git push` or `git commit`
- Deleting files or directories
- Changing the public API surface
- Adding new command types to the parser

## Remember

1. **Write tests** - comprehensive test coverage is required for all new code
2. **Small diffs** - focused changes are easier to review
3. **Use existing patterns** - consistency matters
4. **Ask when uncertain** - better than wrong assumptions
5. **Use `t.Parallel()`** - all tests should run in parallel
