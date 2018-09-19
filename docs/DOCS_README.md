# Docs Build Workflow

The documentation for the Cosmos SDK is hosted at:

- https://cosmos.network/docs/ and
- https://cosmos-staging.interblock.io/docs/

built from the files in this (`/docs`) directory for
[master](https://github.com/cosmos/cosmos-sdk/tree/master/docs)
and [develop](https://github.com/cosmos/cosmos-sdk/tree/develop/docs),
respectively.

## How It Works

There is a Jenkins job listening for changes in the `/docs` directory, on both
the  `master` and `develop` branches. Any updates to files in this directory
on those branches will automatically trigger a website deployment. Under the hood,
a private website repository has make targets consumed by a standard Jenkins task.

## README

The [README.md](./README.md) is also the landing page for the documentation
on the website.

## Config.js

The [config.js](./config.js) generates the sidebar and Table of Contents
on the website docs. Note the use of relative links and the omission of
file extensions. Additional features are available to improve the look
of the sidebar.

## Links

**NOTE:** Strongly consider the existing links - both within this directory
and to the website docs - when moving or deleting files.

Relative links should be used nearly everywhere, having discovered and weighed the following:

### Relative

Where is the other file, relative to the current one?

- works both on GitHub and for the VuePress build
- confusing / annoying to have things like: `../../../../myfile.md`
- requires more updates when files are re-shuffled

### Absolute

Where is the other file, given the root of the repo?

- works on GitHub, doesn't work for the VuePress build
- this is much nicer: `/docs/hereitis/myfile.md`
- if you move that file around, the links inside it are preserved (but not to it, of course)

### Full

The full GitHub URL to a file or directory. Used occasionally when it makes sense
to send users to the GitHub.

## Building Locally

Not currently possible but coming soon! Doing so requires
assets held in the (private) website repo, installing
[VuePress](https://vuepress.vuejs.org/), and modifying the `config.js`.

## Consistency

Because the build processes are identical (as is the information contained herein), this file should be kept in sync as
much as possible with its [counterpart in the Tendermint Core repo](https://github.com/tendermint/tendermint/blob/develop/docs/DOCS_README.md).
