#!/usr/bin/env bash

find docs/modules ! -name '_category_.json' -type f -exec rm -rf {} +
rm -rf docs/tooling/01-cosmovisor.md
rm -rf docs/tooling/02-confix.md
rm -rf docs/tooling/03-hubl.md
rm -rf docs/packages/01-depinject.md
rm -rf docs/packages/02-collections.md
rm -rf docs/packages/03-orm.md
rm -rf docs/core/17-autocli.md
rm -rf docs/run-node/04-rosetta.md
rm -rf docs/architecture
rm -rf docs/spec
rm -rf docs/rfc
rm -rf docs/migrations/02-upgrading.md
rm -rf versioned_docs versioned_sidebars versions.json