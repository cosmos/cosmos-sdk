# Documentation Maintenance Overview

The documentation found in this directory is hosted at:

- https://cosmos.network/docs/

and built using [VuePress](https://vuepress.vuejs.org/) from the Cosmos website repo:

- https://github.com/cosmos/cosmos.network

Under the hood, Jenkins listens for changes (on develop or master) in ./docs then rebuilds
either the staging or production site depending on which branch the changes were made.

To update the Table of Contents (layout of the documentation sidebar), edit the
`config.js` in this directory, while the `README.md` is the landing page for the
website documentation.
