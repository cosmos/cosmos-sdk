# Install

On a good day, basecoin can be installed like a normal Go program:

```
go get -u github.com/tendermint/basecoin/cmd/basecoin
```

In some cases, if that fails, or if another branch is required,
we use `glide` for dependency management.
Thus, assuming you've already run `go get` or otherwise cloned the repo,
the correct way to install is:

```
cd $GOPATH/src/github.com/tendermint/basecoin
git pull origin master
make get_vendor_deps
make install
```

This will create the `basecoin` binary in `$GOPATH/bin`.
Note the `make get_vendor_deps` command will install `glide` and the correct version of all dependencies.

If you need another branch, make sure to run `git checkout <branch>` before the `make` commands.

