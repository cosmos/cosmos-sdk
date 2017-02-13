# Go Basics

This document is designed for developers new to the go language, especially
experienced developers who are learning go for the purpose of using Tendermint.

Go is a rather simple language, which aims to produce fast, maintainable
programs, while minimizing development effort.  In order to speed up
development, the go community has adopted quite a number of conventions, which
are used in almost every open source project. The same way one rails dev can
learn a new project quickly as they all have the same enforced layout,
programming following these conventions allows for interoperability with much
of the go tooling, and a much more fluid development experience.

First of all, you should read through [Effective
Go](https://golang.org/doc/effective_go.html) to get a feel for the language
and the constructs. And maybe pick up a book, read a tutorial, or do what you
feel best to feel comfortable with the syntax.

Second, you need to set up your go environment.  In go, all code hangs out
GOPATH.  You don't have a separate root directory for each project. Pick a nice
locations (like `$HOME/go`) and `export GOPATH` in your startup scripts
(`.bashrc` or the like). Note that go compiles all programs to `$GOPATH/bin`,
similarly PATH will need to be updated in the startup scripts. If your are
editing `.bashrc` (typically found in HOME)  you would add the following lines:

```
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```

Now, when you run `go get github.com/tendermint/basecoin`, this will create the
directory `$GOPATH/src/github.com/tendermint/basecoin`, checkout the master
branch with git, and try to compile if there are any scripts.  All your repos
will fit under GOPATH with a similar logic.  Just pick good names for your
github repos. If you put your code outside of GOPATH/src or have a path other
than the url of the repo, you can expect errors.  There are ways to do this,
but quite complex and not worth the bother.

Third, every repo in `$GOPATH/src` is checkout out of a version control system
(commonly git), and you can go into those directories and manipulate them like
any git repo (`git checkout develop`, `git pull`, `git remote set-url origin
$MY_FORK`).  `go get -u $REPO` is a nice convenience to do a `git pull` on the
master branch and recompile if needed.  If you work on develop, get used to
using the git commands directly in these repos.
[Here](https://tendermint.com/docs/guides/contributing) are some more tips on
using git with open source go projects with absolute dependencies such as
Tendermint. 

Fourth, installing a go program is rather easy if you know what to do.  First
to note is all programs compiles with `go install` and end up in `$GOPATH/bin`.
`go get` will checkout the repo, then try to `go install` it. Many repos are
mainly a library that also export (one or more) commands, in these cases there
is a subdir called `cmd`, with a different subdir for each command, using the
command name as the directory name.  To compile these commands, you can go
something like `go install github.com/tendermint/basecoin/cmd/basecoin` or to
compile all the commands `go install github.com/tendermint/basecoin/cmd/...`
(... is a go tooling shortcut for all subdirs, like `*`).

Fifth, there isn't good dependency management built into go. By default, when
compiling a go program which imports another repo, go will compile using the
latest master branch, or whichever version you have checked out and located.
This can cause serious issues, and there is tooling to do dependency
management.  As of go 1.6, the `vendor` directory is standard and a copy of a
repo will be used rather than the repo under GOPATH.  In order to create and
maintain the code in the vendor directory, various tools have been created,
with [glide](https://github.com/Masterminds/glide) being popular and in use in
all the Tendermint repos. In this case, `go install` is not enough.  If you are
working on code from the Tendermint, you will usually want to do:

```
go get github.com/tendermint/$REPO
cd $GOPATH/src/github.com/tendermint/$REPO
make get_vendor_deps
make install
make test
```

`make get_vendor_deps` should update the vendor directory using glide, `make
install` will compile all commands.  `make test` is good to run the test suite
and make sure things are working with your environment... failing tests are
much easier to debug than a malfunctioning program.

Okay, that's it, with this info you should be able to follow along and
trouble-shoot any issues you have with the rest of the guide.

