#!/usr/bin/env bash

find docs/build/modules ! -name '_category_.json' -type f -exec rm -rf {} +
rm -rf docs/build/tooling/01-cosmovisor.md
rm -rf docs/build/tooling/02-confix.md
rm -rf docs/build/tooling/03-hubl.md
rm -rf docs/build/packages/01-depinject.md
rm -rf docs/build/packages/02-collections.md
rm -rf docs/learn/advaced-concepts/17-autocli.md
rm -rf docs/build/architecture
rm -rf docs/build/spec
rm -rf docs/build/rfc
rm -rf  docs/learn/advanced/17-autocli.md
rm -rf docs/build/migrations/02-upgrading.md
rm -rf versioned_docs versioned_sidebars versions.json
