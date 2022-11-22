# IaViewer

`iaviewer` is a utility to inspect the contents of a persisted iavl tree, given (a copy of) the leveldb store.
This can be quite useful for debugging, especially when you find odd errors, or non-deterministic behavior.
Below is a brief introduction to the tool.

## Installation

Once this is merged into the offical repo, master, you should be able to do:

```shell
go get github.com/cosmos/iavl
cd ${GOPATH}/src/github.com/cosmos/iavl
make install
```

## Using the tool

First make sure it is properly installed and you have `${GOPATH}/bin` in your `PATH`.
Typing `iaviewer` should run and print out a usage message.

### Sample databases

Once you understand the tool, you will most likely want to run it on captures from your
own abci app (built on cosmos-sdk or weave), but as a tutorial, you can try to use some
captures from an actual bug I found in my code... Same data, different hash.

```shell
mkdir ./testdata
cd ./testdata
curl -L https://github.com/iov-one/iavl/files/2860877/bns-a.db.zip > bns-a.db.zip
unzip bns-a.db.zip
curl -L https://github.com/iov-one/iavl/files/2860878/bns-b.db.zip > bns-b.db.zip
unzip bns-b.db.zip
```

Now, if you run `ls -l`, you should see two directories... `bns-a.db` and `bns-b.db`

### Inspecting available versions

```shell
iaviewer versions ./bns-a.db ""
```

This should print out a list of 20 versions of the code. Note the the iavl tree will persist multiple
historical versions, which is a great aid in forensic queries (thanks Tendermint team!). For the rest
of the cases, we will consider only the last two versions, 190257 (last one where they match) and 190258
(where they are different).

### Checking keys and app hash

First run these two and take a quick a look at the output:

```shell
iaviewer data ./bns-a.db ""
iaviewer data ./bns-a.db "" 190257
```

Notice you see the different heights and there is a change in size and app hash.
That's what happens when we process a transaction. Let's go further and use
the handy tool `diff` to compare two states.

```shell
iaviewer data ./bns-a.db "" 190257 > a-last.data
iaviewer data ./bns-b.db "" 190257 > b-last.data

diff a-last.data b-last.data
```

Same, same :)
But if we take the current version...

```shell
iaviewer data ./bns-a.db "" 190258 > a-cur.data
iaviewer data ./bns-b.db "" 190258 > b-cur.data

diff a-cur.data b-cur.data
```

Hmmm... everything is the same, except the hash. Odd...
So odd that I [wrote an article about it](https://medium.com/@ethan.frey/tracking-down-a-tendermint-consensus-failure-77f6ff414406)

And finally, if we want to inspect which keys were modified in the last block:

```shell
diff a-cur.data a-last.data
```

You should see 6 writes.. the `_i.usernft_*` are the secondary indexes on the username nft.
`sigs.*` is setting the nonce (if this were an update, you would see a previous value).
And `usrnft:*` is creating the actual username nft.

### Checking the tree shape

So, remember above, when we found that the current state of a and b have the same data
but different hashes. This must be due to the shape of the iavl tree.
To confirm that, and possibly get more insights, there is another command.

```shell
iaviewer shape ./bns-a.db "" 190258 > a-cur.shape
iaviewer shape ./bns-b.db "" 190258 > b-cur.shape

diff a-cur.shape b-cur.shape
```

Yup, that is quite some difference. You can also look at the tree as a whole.
So, stretch your terminal nice and wide, and... 

```shell
less a-cur.shape
```

It has `-5 ` for an inner node of depth 5, and `*6 ` for a leaf node (data) of depth 6.
Indentation also suggests the shape of the tree. 

Note, if anyone wants to improve the visualization, that would be awesome.
I have no idea how to do this well, but at least text output makes some
sense and is diff-able.