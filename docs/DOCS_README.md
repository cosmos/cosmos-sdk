# Updating the docs

If you want to open a PR on the Cosmos SDK to update the documentation, please follow the guidelines in the [`CONTRIBUTING.md`](https://github.com/cosmos/cosmos-sdk/tree/master/CONTRIBUTING.md#updating-documentation)

## Translating

- Docs translations live in a `docs/country-code/` folder, where `country-code` stands for the country code of the language used (`cn` for Chinese, `kr` for Korea, `fr` for France, ...).
- Always translate content living on `master`.
- Only content under `/docs/intro/`, `/docs/basics/`, `/docs/core/`, `/docs/building-modules/` and `docs/run-node/` needs to be translated, as well as `docs/README.md`. It is also nice (but not mandatory) to translate `/docs/spec/`.
- Specify the release/tag of the translation in the README of your translation folder. Update the release/tag each time you update the translation.

## Docs Build Workflow

The documentation for the Cosmos SDK is hosted at https://cosmos.network/docs/

built from the files in this (`/docs`) directory for
[master](https://github.com/cosmos/cosmos-sdk/tree/master/docs).

### How It Works

There is a CircleCI job listening for changes in the `/docs` directory, on
the `master` branch. Any updates to files in this directory
on that branch will automatically trigger a website deployment. Under the hood,
the private website repository has a `make build-docs` target consumed by a CircleCI job in that repo.

## README

The [README.md](./README.md) is also the landing page for the documentation
on the website. During the Jenkins build, the current commit is added to the bottom
of the README.

## Config.js

The [config.js](./.vuepress/config.js) generates the sidebar and Table of Contents
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

Make sure you are in the `docs` directory and run the following commands:

```sh
rm -rf node_modules
```

This command will remove old version of the visual theme and required packages. This step is optional.

```sh
npm install
```

Install the theme and all dependencies.

```sh
npm run serve
```

Run `pre` and `post` hooks and start a hot-reloading web-server. See output of this command for the URL (it is often https://localhost:8080).

To build documentation as a static website run `npm run build`. You will find the website in `.vuepress/dist` directory.

## Build RPC Docs

First, run `make tools` from the root of repo, to install the swagger-ui tool.

Then, edit the `swagger.yaml` manually; it is found [here](https://github.com/cosmos/cosmos-sdk/blob/master/client/lcd/swagger-ui/swagger.yaml)

Finally, run `make update_gaia_lite_docs` from the root of the repo.

## Search

We are using [Algolia](https://www.algolia.com) to power full-text search. This uses a public API search-only key in the `config.js` as well as a [cosmos_network.json](https://github.com/algolia/docsearch-configs/blob/master/configs/cosmos_network.json) configuration file that we can update with PRs.

## Consistency

Because the build processes are identical (as is the information contained herein), this file should be kept in sync as
much as possible with its [counterpart in the Tendermint Core repo](https://github.com/tendermint/tendermint/blob/v0.34.0/docs/DOCS_README.md).

### Update and Build the RPC docs

1. Execute the following command at the root directory to install the swagger-ui generate tool.
   ```bash
   make tools
   ```
2. Edit API docs
   1. Directly Edit API docs manually: `client/lcd/swagger-ui/swagger.yaml`.
   2. Edit API docs within the [Swagger Editor](https://editor.swagger.io/). Please refer to this [document](https://swagger.io/docs/specification/2-0/basic-structure/) for the correct structure in `.yaml`.
3. Download `swagger.yaml` and replace the old `swagger.yaml` under fold `client/lcd/swagger-ui`.
4. Compile gaiacli
   ```bash
   make install
   ```
