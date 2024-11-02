# Website

This website is built using [Docusaurus 2](https://docusaurus.io/), a modern static website generator.

:::note
The documentation is hosted on [docs.cosmos.network](https://docs.cosmos.network).
:::

If you want to open a PR in Cosmos SDK to update the documentation, please follow the guidelines in [`CONTRIBUTING.md`](https://github.com/cosmos/cosmos-sdk/tree/main/CONTRIBUTING.md#updating-documentation) and the [Documentation Writing Guidelines](./DOC_WRITING_GUIDELINES.md).

### Installation

```
$ yarn
```

### Local Development

```
$ yarn start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

### Build

```
$ yarn build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

### Deployment

Using SSH:

```
$ USE_SSH=true yarn deploy
```

Not using SSH:

```
$ GIT_USER=<Your GitHub username> yarn deploy
```

If you are using GitHub pages for hosting, this command is a convenient way to build the website and push to the `gh-pages` branch.
