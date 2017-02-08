# Install

We use glide for dependency management.  The prefered way of compiling from source is the following:

```
go get -d github.com/tendermint/basecoin/cmd/basecoin
cd $GOPATH/src/github.com/tendermint/basecoin
make get_vendor_deps
make install
```

This will create the `basecoin` binary in `$GOPATH/bin`.

