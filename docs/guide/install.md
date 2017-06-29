# Install

If you aren't used to compile go programs and just want the released
version of the code, please head to our [downloads](https://tendermint.com/download)
page to get a pre-compiled binary for your platform.

On a good day, basecoin can be installed like a normal Go program:

```
go get -u github.com/tendermint/basecoin/cmd/...
```

In some cases, if that fails, or if another branch is required,
we use `glide` for dependency management.
Thus, assuming you've already run `go get` or otherwise cloned the repo,
the correct way to install is:

```
cd $GOPATH/src/github.com/tendermint/basecoin
git pull origin master
make all
```

This will create the `basecoin` binary in `$GOPATH/bin`.
`make all` implies `make get_vendor_deps` and uses `glide` to install the
correct version of all dependencies. It also tests the code, including
some cli tests to make sure your binary behaves properly.

If you need another branch, make sure to run `git checkout <branch>`
before `make all`. And if you switch branches a lot, especially
touching other tendermint repos, you may need to `make fresh` sometimes
so glide doesn't get confused with all the branches and versions lying around.

