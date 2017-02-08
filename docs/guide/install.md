# Install

We use glide for dependency management.  The prefered way of compiling from source is the following:

```
go get -u github.com/tendermint/basecoin
cd $GOPATH/src/github.com/tendermint/basecoin
git checkout develop # (until we release v0.9)
make get_vendor_deps
make install
```

This will create the `basecoin` binary in `$GOPATH/bin`.

