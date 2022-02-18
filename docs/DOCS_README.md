# Updating the docs

If you want to open a PR in Cosmos SDK to update the documentation, please follow the guidelines in [`CONTRIBUTING.md`](https://github.com/cosmos/cosmos-sdk/tree/master/CONTRIBUTING.md#updating-documentation).

## Internationalization

* Translations for documentation live in a `docs/<locale>/` folder, where `<locale>` is the language code for a specific language. For example, `zh` for Chinese, `ko` for Korean, `ru` for Russian, etc.
* Each `docs/<locale>/` folder must follow the same folder structure within `docs/`, but only content in the following folders needs to be translated and included in the respective `docs/<locale>/` folder:
    * `docs/basics/`
    * `docs/building-modules/`
    * `docs/core/`
    * `docs/ibc/`
    * `docs/intro/`
    * `docs/migrations/`
    * `docs/run-node/`
* Each `docs/<locale>/` folder must also have a `README.md` that includes a translated version of both the layout and content within the root-level [`README.md`](https://github.com/cosmos/cosmos-sdk/tree/master/docs/README.md). The layout defined in the `README.md` is used to build the homepage.
* Always translate content living on `master` unless you are revising documentation for a specific release. Translated documentation like the root-level documentation is semantically versioned.
* For additional configuration options, please see [VuePress Internationalization](https://vuepress.vuejs.org/guide/i18n.html).

## Docs Build Workflow

The documentation for Cosmos SDK is hosted at https://docs.cosmos.network/ and built from the files in the `/docs` directory.

### How It Works

There is a CircleCI job listening for changes in the `/docs` directory for the `master` branch and each supported version tag (`v0.39` and `v0.42`). Any updates to files in the `/docs` directory will automatically trigger a website deployment. Under the hood, the private website repository has a `make build-docs` target consumed by a CircleCI job within that repository.

## README

The [README.md](./README.md) is both the README for the repository and the configuration for the layout of the landing page.

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

* works both on GitHub and for the VuePress build
* confusing / annoying to have things like: `../../../../myfile.md`
* requires more updates when files are re-shuffled

### Absolute

Where is the other file, given the root of the repo?

* works on GitHub, doesn't work for the VuePress build
* this is much nicer: `/docs/hereitis/myfile.md`
* if you move that file around, the links inside it are preserved (but not to it, of course)

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
