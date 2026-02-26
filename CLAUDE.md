# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Test
go test ./...
go test -coverprofile coverage.out -covermode atomic ./...
go test ./parser/ -run TestParseStatements   # single test

# Lint (golangci-lint v2, config in .golangci.yml)
golangci-lint run ./...

# Build
go build .
```

## Architecture

T-SQL Language Server Protocol (LSP) implementation, forked from [sqls](https://github.com/sqls-server/sqls) and adapted for T-SQL/MSSQL.

**Request flow**: `main.go` → jsonrpc2 stdio server → `internal/handler/Server` (25 LSP methods) → delegates to parser/completer/formatter/database subsystems.

### Key packages

- **`internal/handler/`** — LSP handlers. Each method follows: unmarshal params → delegate → marshal result. `handler.go` has the `Server` struct and method dispatch.
- **`parser/`** — Recursive descent T-SQL parser. `batch.go` splits on `GO` separator; all operations are batch-aware via `BatchAtLine()`.
- **`token/`** — Lexer with dialect-aware tokenization.
- **`ast/`** — ~10 node types: Item, MultiKeyword, Aliased, Identifier, Parenthesis, FunctionLiteral, Query, Statement, BeginEnd, TryCatch, IfStatement.
- **`dialect/`** — `Dialect` interface; `MSSQLDialect` supports `[bracket]` identifiers, `@` variables, `#` temp tables.
- **`internal/completer/`** — Completion logic (keywords, columns, tables, schemas, joins).
- **`internal/formatter/`** — SQL formatting, batch-aware.
- **`internal/database/`** — DB interface, async worker, metadata caching.
- **`internal/config/`** — YAML config loading.

### Patterns

- **Batch-based processing**: T-SQL `GO` separator means all parsing/completion/formatting operates per-batch.
- **Dialect system**: Abstract `Dialect` interface; keywords matched via `dialect.MatchKeyword()`.
- **Table-driven tests** with subtests and positional assertions using `token.Pos{Line, Col}`.
- **Config hierarchy**: `-config` CLI flag → LSP `workspace/configuration` → `$XDG_CONFIG_HOME/sqls/config.yml`.

### Development phases (commit history)

1. Fork sqls into tsql-ls
2. GO batch separator preprocessing
3. MSSQLDialect and bracket quoting
4. Extended T-SQL keyword/function coverage
5. T-SQL block parsing (BEGIN/END, TRY/CATCH, IF/ELSE)
