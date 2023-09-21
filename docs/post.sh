#!/usr/bin/env bash

<<<<<<< HEAD
find docs/build/modules ! -name '_category_.json' -type f -exec rm -rf {} +
rm -rf docs/build/tooling/01-cosmovisor.md
rm -rf docs/build/tooling/02-confix.md
rm -rf docs/build/tooling/03-hubl.md
rm -rf docs/build/packages/01-depinject.md
rm -rf docs/build/packages/02-collections.md
rm -rf docs/build/packages/03-orm.md
rm -rf docs/develop/advaced-concepts/17-autocli.md
rm -rf docs/user/run-node/04-rosetta.md
rm -rf docs/build/architecture
rm -rf docs/build/spec
rm -rf docs/build/rfc
rm -rf  docs/develop/advanced/17-autocli.md
rm -rf docs/build/migrations/02-upgrading.md
rm -rf versioned_docs versioned_sidebars versions.json
=======
find build/modules ! -name '_category_.json' -type f -exec rm -rf {} +
rm -rf build/tooling/01-cosmovisor.md
rm -rf build/tooling/02-confix.md
rm -rf build/tooling/03-hubl.md
rm -rf build/packages/01-depinject.md
rm -rf build/packages/02-collections.md
rm -rf build/packages/03-orm.md
rm -rf learn/advaced-concepts/17-autocli.md
rm -rf user/run-node/04-rosetta.md
rm -rf build/architecture
rm -rf build/spec
rm -rf build/rfc
rm -rf learn/advanced/17-autocli.md
rm -rf build/migrations/02-upgrading.md
>>>>>>> 2efafee65 (chore: rename develop to learn (#17821))
