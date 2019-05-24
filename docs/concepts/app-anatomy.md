# Anatomy of an SDK Application

## Pre-requisite reading

- [High-level overview of an SDK application architecture](../intro/sdk-app-architecture.md)
- [Cosmos SDK design overview](../intro/sdk-design.md)

## Synopsis

This document describes the core parts of a Cosmos SDK application. The placeholder name for this application will be `app`.

- [Node Client](#node-client)
- [Core Application File](#core-application-file)
- [Modules](#modules)
- [Intefaces](#interfaces)
- [Dependencies and Makefile](#dependencies-and-makefile)

The core parts listed above will generally translate to the following file tree in the application directory:

```
./application
├── cmd/
│   ├── appd
│   └── appcli
├── app.go
├── x/
│   ├── auth
│   └── bank
├── Gopkg.toml
└── Makefile
``` 

## Node Client (Daemon)

The Daemon, or Full-Node Client, is the core process of an SDK-based blockchain. Participants in the network run this process to initialize their state-machine, connect with other full-nodes and update their state-machine as new blocks come in. 

```
                ^  +-------------------------------+  ^
                |  |                               |  |   Built with Cosmos SDK
                |  |  State-machine = Application  |  |
                |  |                               |  v
                |  +-------------------------------+
                |  |                               |  ^
Blockchain node |  |           Consensus           |  |
                |  |                               |  |
                |  +-------------------------------+  |   Tendermint Core
                |  |                               |  |
                |  |           Networking          |  |
                |  |                               |  |
                v  +-------------------------------+  v
```
The blockchain full-node presents itself as a binary, generally suffixed by `-d` (e.g. `appd` for `app` or `gaiad` for the `gaia`) for "daemon". This binary is built by running a simple `main.go` function placed in `cmd/appd/`. This operation usually happens through the [Makefil](#dependencies-and-makefile).

To learn more about the `main.go` function, [click here](./node#`main.go`).

Once the main binary is built, the node can be started by running the `start` command. The core logic behind the `start` command is implemented in the SDK itself in the [`/server/start.go`](https://github.com/cosmos/cosmos-sdk/blob/master/server/start.go) file. The main [`start` command function](https://github.com/cosmos/cosmos-sdk/blob/master/server/start.go#L31) takes a [`context`](https://godoc.org/github.com/cosmos/cosmos-sdk/client/context) and [`appCreator`](#constructor-function-(`appCreator`)) as arguments. The `appCreator` is a constructor function for the SDK application, and is used in the starting process of the full-node. 

The `start` command function primarily does three things:

1- Create an instance of the state-machine defined in [`app.go`](#core-application-file) using the `appCreator`. 
2- Initialize the state-machine with the latest known state, extracted from the `db` stored in the `~/.appd/data` folder. 
3- Create and start a new Tendermint instance. Among other things, the node will perform a handshake with its peer. It will get the latest `appBlockHeight` from them, and replay blocks to get there if `appBlockHeight`is greater than the current height. If the node is starte

To learn more about the `start` command, [click here](./node#`start`-command).

## Core Application File

In general, the core of the state-machine is defined in a file called `app.go`. It mainly contains the **type definition of the application** and functions to **create and initialize it**. 

### Type Definition of the Application

### Constructor Function (`appCreator`)

This function constructs a new application of the type defined above. It is passed to the `start` function that runs the daemon, and called in the `InitChain`

### InitChainer

## Modules (`./x/`)

## Interfaces

## Dependencies and Makefile 