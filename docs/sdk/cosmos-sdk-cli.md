# cosmos-sdk-cli

Create a new application specific blockchain project based on the Cosmos SDK with a single command.

::: warning
🚧 cosmos-sdk-cli is an experimental project
:::

## Installation

```shell
$ go get github.com/cosmos/cosmos-sdk
$ cd $GOPATH/src/github.com/cosmos/cosmos-sdk
$ make install_cosmos-sdk-cli
```

This will install a binary `cosmos-sdk-cli`

## Creating a new project

```shell
$ cosmos-sdk-cli init <your_proyect_name>
```

This will initialize a project, the dependencies, directory structures with the specified project name.

### Example:
```shell
$ cosmos-sdk-cli init testzone -p github.com/your_user_name/testzone
```
`-p [remote-project-path]`. If this is not provided, it creates testzone under `$GOPATH/src/`

```shell
$ cd $GOPATH/src/github.com/your_user_name/testzone
$ make
```

This will create two binaries (`testzonecli` and `testzoned`) under `bin` folder. `testzoned` is the full node of the application which you can run, and `testzonecli` is your light client.
