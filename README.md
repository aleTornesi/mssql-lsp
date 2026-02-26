# tsql-ls: T-SQL Language Server

A Language Server Protocol (LSP) implementation for T-SQL, forked from [sqls](https://github.com/sqls-server/sqls) and specialized for Microsoft SQL Server.

## Features

### LSP Capabilities

- **Completion** — keywords, tables, columns, schemas, joins, system stored procedures, snippets
- **Hover** — column types, table details, object info
- **Go to Definition / Type Definition / Declaration** — navigate to CTEs, temp tables, variables, table aliases
- **References & Highlights** — find all references, document highlights
- **Rename & Prepare Rename** — safe symbol renaming with preview
- **Signature Help** — function parameter hints
- **Formatting** — full document, range, and on-type formatting
- **Code Actions** — 6 actions (quick fixes for unclosed blocks, remove unused CTEs, surround with TRY/CATCH)
- **Diagnostics** — 16 rules (unclosed blocks, duplicate variables, undefined variables, unreferenced CTEs, transaction mismatches, and more)
- **Folding Ranges** — collapse BEGIN/END, TRY/CATCH, IF/ELSE, multi-line comments
- **Document & Workspace Symbols** — outline view and cross-file symbol search
- **Selection Range** — smart expand/shrink selection
- **Semantic Tokens** — full, range, and delta semantic highlighting
- **Inlay Hints** — inline type annotations
- **Linked Editing** — synchronized editing of related identifiers

### T-SQL Specifics

- `[bracket]` quoted identifiers
- `@variables` and `@@globals`
- `#temp` and `##global_temp` tables
- `GO` batch separator — all operations are batch-aware
- BEGIN/END, TRY/CATCH, IF/ELSE block parsing

## Installation

```shell
go install github.com/atornesi/tsql-ls@latest
```

## Configuration

### Configuration Methods (in priority order)

1. `-config` CLI flag
2. `workspace/configuration` from LSP client
3. `$XDG_CONFIG_HOME/sqls/config.yml` (defaults to `~/.config/sqls/config.yml`)

### Configuration File

```yaml
lowercaseKeywords: false
connections:
  - alias: local_mssql
    driver: mssql
    dataSourceName: "sqlserver://user:password@localhost:1433?database=mydb"
  - alias: individual_mssql
    driver: mssql
    proto: tcp
    user: sa
    passwd: YourPassword
    host: 127.0.0.1
    port: 1433
    dbName: mydb
```

The first connection in the list is the default.

| Key            | Description                              |
| -------------- | ---------------------------------------- |
| alias          | Connection alias name. Optional.         |
| driver         | `mssql`. Required.                       |
| dataSourceName | DSN (takes precedence over individual fields). |
| proto          | `tcp`, `udp`, `unix`.                    |
| user           | User name.                               |
| passwd         | Password.                                |
| host           | Host.                                    |
| port           | Port.                                    |
| dbName         | Database name.                           |
| params         | Additional parameters. Optional.         |

## Editor Setup

### Neovim (nvim-lspconfig)

```lua
require('lspconfig').sqls.setup {
  cmd = { 'tsql-ls', '-config', vim.fn.expand('~/.config/sqls/config.yml') },
  filetypes = { 'sql' },
  settings = {
    sqls = {
      connections = {
        {
          driver = 'mssql',
          dataSourceName = 'sqlserver://sa:password@localhost:1433?database=mydb',
        },
      },
    },
  },
}
```

### VS Code

Use a generic LSP client extension (e.g., [vscode-lsp-sample](https://github.com/nicktomlin/vscode-lsp)) pointing to the `tsql-ls` binary.

### Other Editors

Any editor with LSP client support can use tsql-ls. Point the client to the `tsql-ls` binary with `--stdio` transport (the default).
