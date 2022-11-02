# Updating the docs

If you want to open a PR in Cosmos SDK to update the documentation, please follow the guidelines in [`CONTRIBUTING.md`](https://github.com/cosmos/cosmos-sdk/tree/main/CONTRIBUTING.md#updating-documentation) and the [Documentation Writing Guidelines](./DOC_WRITING_GUIDELINES.md).

## Stack

The documentation for Cosmos SDK is hosted at https://docs.cosmos.network and built from the files in the `/docs` directory.
It is built using the following stack:

* [Docusaurus 2](https://docusaurus.io)
* Vuepress (pre v0.47)
* [Algolia DocSearch](https://docsearch.algolia.com/)

  ```js
          algolia: {
        appId: "QLS2QSP47E",
        apiKey: "067b84458bfa80c295e1d4f12c461911",
        indexName: "cosmos_network",
        contextualSearch: false,
      },
  ```

## Docs Build Workflow

// TODO

### How It Works

There is a GitHub Action listening for changes in the `/docs` directory for the `main` branch and each supported version branch (e.g. `release/v0.46.x`). Any updates to files in the `/docs` directory will automatically trigger a website deployment. Under the hood, the private website repository has a `make build-docs` target consumed by a Github Action within that repository.

## How to Build the Docs Locally

Go to the `docs` directory and run the following commands:

```shell
cd docs
npm install
```

For starting only the local documentation, run:

```shell
npm start
```

It runs `pre.sh` scripts to get all the docs that are not already in the `docs/docs` folder.
It also runs `post.sh` scripts to clean up the docs and remove unnecessary files when quitting.

To build documentation as a static website run:

```shell
npm run build
```

## What to for new major SDK versions

When a new major version of the SDK is released, the following steps should be taken:

