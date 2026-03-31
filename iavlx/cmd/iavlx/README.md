# iavlx CLI

Offline inspection and management tool for iavlx data directories.

## Commands

- `iavlx view [dir]` — Interactive TUI for browsing tree data (changesets, checkpoints, WAL entries, nodes, orphans, commit info)
- `iavlx import --from [dir] --to [dir] --format v1-leveldb` — One-time offline migration from iavl/v1 LevelDB to iavlx format
- `iavlx rollback [dir] --version [version]` — Roll back a multi-tree to a specific version (node must be stopped)

## Code quality warning

Unlike the `iavlx/internal` package (which has been carefully reviewed and documented), the CLI
code in this directory was mostly vibe-coded as quick exploration tooling. It works, but:

- Minimal doc comments and no tests
- Some silent error swallowing (especially in the TUI views)
- Magic numbers for table column widths and layout
- Duplicated patterns across view files (could be consolidated)
- The interactive TUI (bubbletea-based) is the roughest part — the `import` and `rollback`
  commands are thin wrappers around well-tested `internal` functions and are more reliable

If you're modifying this code, treat it as a prototype that works rather than production-grade
tooling. The `internal` package it delegates to IS production-grade.