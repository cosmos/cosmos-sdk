# internal-7000 testnet

This is a testnet only for internal testing. 

## Install golang

One way to install golang on an Ubuntu server is to use [gvm](https://github.com/moovweb/gvm).

You can use whatever way works best for you, just make sure that it is correctly installed.

## Install Gaia

```bash
go get github.com/cosmos/cosmos-sdk

cd $GOPATH/src/github.com/cosmos/cosmos-sdk

git checkout <next RC>

make get_tools

make get_vendor_deps

make install

gaiad version -> 0.20.0-dev-0d94c5a2

gaiacli version -> 0.20.0-dev-0d94c5a2
```

## Generating a genesis transaction

```bash
gaiad init gentx --name=<some_name>

cat $HOME/.gaiad/config/gentx/... -> there is a single json file, add it as your_name.json into the `internal-7000` folder
```

## Starting the testnet

```bash
rm -r $HOME/.gaiad/config/gentx

cd $GOPATH/src/github.com/cosmos/cosmos-sdk

git checkout develop

git pull origin develop

cp -r $GOPATH/src/github.com/cosmos/cosmos-sdk/cmd/gaia/testnets/internal-7000 $HOME/.gaiad/config/gentx

gaiad init --gen-txs --chain-id=internal-7000

gaiad start
```

## Testing the testnet

Go nuts, try everything. 