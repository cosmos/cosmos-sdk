# Migration Framework

A generic, AST-based framework for automating Go code migrations across breaking changes in libraries.

## Capabilities

- **Import path rewriting** — update old import paths to new ones, with support for wildcard sub-package matching and exceptions
- **Function argument updates** — handle function signature changes where arguments were added or removed
- **Type/struct renames** — replace references to renamed types across variable declarations, function parameters, return types, struct fields, and type assertions
- **Complex function replacements** — replace single function calls with multi-statement rewrites (e.g., replacing a helper with inline code)
- **Go module management** — update, add, remove, and replace dependencies in `go.mod` files

## Architecture

The framework uses a two-layer design:

1. **Core engine** (`tools/migration/`) — generic AST traversal and transformation logic
2. **Version-specific config** (`tools/migration/v54/`) — declares the concrete migration rules for a specific version upgrade

This separation means the core engine can be reused for future version migrations by simply adding a new config directory.

## Adding a new migration

1. Create a new directory under `tools/migration/` (e.g., `v55/`)
2. Define your migration rules in separate files following the pattern in `v54/`
3. Wire them together in `main.go`
