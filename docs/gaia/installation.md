## Install Gaia

This guide will explain how to install the `gaiad` and `gaiacli` entrypoints onto your system. With these installed on a server, you can participate in the latest testnet as either a [Full Node](./join-testnet.md#run-a-full-node) or a [Validator](./validators/validator-setup.md).

### Install Go

Install `go` by following the [official docs](https://golang.org/doc/install). Remember to set your `$GOPATH`, `$GOBIN`, and `$PATH` environment variables, for example:

```bash
mkdir -p $HOME/go/bin
echo "export GOPATH=$HOME/go" >> ~/.bash_profile
echo "export GOBIN=$GOPATH/bin" >> ~/.bash_profile
echo "export PATH=$PATH:$GOBIN" >> ~/.bash_profile
```

::: tip
**Go 1.11.5+** is required for the Cosmos SDK.
:::

### Install the binaries

Next, let's install the latest version of Gaia. Here we'll use the `master` branch, which contains the latest stable release.
If necessary, make sure you `git checkout` the correct
[released version](https://github.com/cosmos/cosmos-sdk/releases).

```bash
mkdir -p $GOPATH/src/github.com/cosmos
cd $GOPATH/src/github.com/cosmos
git clone https://github.com/cosmos/cosmos-sdk
cd cosmos-sdk && git checkout master
make tools install
```

> *NOTE*: If you have issues at this step, please check that you have the latest stable version of GO installed.

That will install the `gaiad` and `gaiacli` binaries. Verify that everything is OK:

```bash
$ gaiad version --long
$ gaiacli version --long
```

`gaiacli` for instance should output something similar to:

```
cosmos-sdk: 0.31.2-10-g1fba7308
git commit: 1fba7308fa226e971964cd6baad9527d4b51d9fc
vendor hash: 1aec7edfad9888a967b3e9063e42f66b28f447e6
build tags: netgo ledger
go version go1.11.5 linux/amd64
```

##### Build Tags

Build tags indicate special features that have been enabled in the binary.

| Build Tag | Description                                     |
| --------- | ----------------------------------------------- |
| netgo     | Name resolution will use pure Go code           |
| ledger    | Ledger devices are supported (hardware wallets) |

### Install binary distribution via snap (Linux only)

Gaia can be installed on various GNU/Linux distributions from the [Snapcraft.io](https://snapcraft.io/gaia) store:

```bash
$ sudo snap install gaia
```

Development builds are available through the `edge` channel:

```bash
$ sudo snap install --edge gaia
```

::: tip
At the time of writing, only the following [architectures are supported](https://build.snapcraft.io/user/cosmos/cosmos-sdk): `amd64` `i386` `arm64` `armhf` `ppc64el` `s390x`.
:::

`snap` installs Gaia binaries as `gaia.gaiad` and `gaia.gaiacli`. It is recommended to create commands aliases for the user's convenience once the package is installed:

```
$ sudo snap alias gaia.gaiad gaiad
$ sudo snap alias gaia.gaiacli gaiacli
```

::: warning
Note that the binaries provided by the snap package save their data into **$HOME/snap/gaia/** instead of **$HOME**.
:::

Please refer to [Snap documentation](https://docs.snapcraft.io/installing-snapd/6735) for specific information on how to install `snap` on your distribution.


### Next

Now you can [join the public testnet](./join-testnet.md) or [create you own  testnet](./deploy-testnet.md)
