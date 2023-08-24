#!/usr/bin/env bash

find docs/integrate/modules ! -name '_category_.json' -type f -exec rm -rf {} +
rm -rf docs/integrate/tooling/01-cosmovisor.md
rm -rf docs/integrate/tooling/02-confix.md
rm -rf docs/integrate/tooling/03-hubl.md
rm -rf docs/packages/01-depinject.md
rm -rf docs/integrate/packages/02-collections.md
rm -rf docs/integrate/packages/03-orm.md
rm -rf docs/develop/advaced-concepts/17-autocli.md
rm -rf docs/user/run-node/04-rosetta.md
rm -rf docs/architecture
rm -rf docs/spec
rm -rf docs/rfc
rm -rf  docs/develop/advanced-concepts/17-autocli.md
rm -rf docs/migrations/02-upgrading.md
rm -rf versioned_docs versioned_sidebars versions.json