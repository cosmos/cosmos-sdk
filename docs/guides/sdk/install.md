# Install

The fastest and easiest way to install the Cosmos SDK binaries
is to run [this script](https://github.com/cosmos/cosmos-sdk/blob/develop/scripts/install_sdk_ubuntu.sh) on a fresh Ubuntu instance. Similarly, you can run [this script](https://github.com/cosmos/cosmos-sdk/blob/develop/scripts/install_sdk_bsd.sh) on a fresh FreeBSD instance. Read the scripts before running them to ensure no untrusted connection is being made, for example we're making curl requests to download golang. Also read the comments / instructions carefully (i.e., reset your terminal after running the script).

Cosmos SDK can be installed to
`$GOPATH/src/github.com/cosmos/cosmos-sdk` like a normal Go program:

```
go get github.com/cosmos/cosmos-sdk
```

If the dependencies have been updated with breaking changes, or if
another branch is required, `dep` is used for dependency management.
Thus, assuming you've already run `go get` or otherwise cloned the repo,
the correct way to install is:

```
cd $GOPATH/src/github.com/cosmos/cosmos-sdk
make get_tools
make get_vendor_deps
make install
make install_examples
```

This will install `gaiad` and `gaiacli` and four example binaries:
`basecoind`, `basecli`, `democoind`, and `democli`.

Verify that everything is OK by running:

```
gaiad version
```

you should see:

```
0.17.3-a5a78eb
```

then with:

```
gaiacli version
```
you should see the same version (or a later one for both).

## Update

Get latest code (you can also `git fetch` only the version desired),
ensure the dependencies are up to date, then recompile.

```
cd $GOPATH/src/github.com/cosmos/cosmos-sdk
git fetch -a origin
git checkout VERSION
make get_vendor_deps
make install
```
