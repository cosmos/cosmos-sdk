# Documentation Maintenance Overview

The documentation found in this directory is hosted at:

- https://cosmos.network/docs/

and built using [VuePress](https://vuepress.vuejs.org/) from the Cosmos website repo:

- https://github.com/cosmos/cosmos.network

which has a [configuration file](https://github.com/cosmos/cosmos.network/blob/develop/docs/.vuepress/config.js) for displaying
the Table of Contents that lists all the documentation. 

Under the hood, Jenkins listens for changes in ./docs then pushes a `docs-staging` branch to the cosmos.network repo with the latest documentation. That branch must be manually PR'd to `develop` then `master` for staging then production. This process should happen in synchrony with a release.

The `README.md` in this directory is the landing page for
website documentation.
