# Install

On a good day, basecoin can be installed like a normal Go program:

```
go get -u github.com/tendermint/basecoin/cmd/basecoin
```

In some cases, if that fails, or if another branch is required,
we use `glide` for dependency management.

The correct way of compiling from source, assuming you've already 
run `go get` or otherwise cloned the repo, is:

```
cd $GOPATH/src/github.com/tendermint/basecoin
git checkout develop # (until we release v0.9)
make get_vendor_deps
make install
```

This will create the `basecoin` binary in `$GOPATH/bin`.

