# Basecoin Example
===============

Here we explain how to get started with a basic Basecoin blockchain, how
to send transactions between accounts using the ``basecli`` tool, and
what is happening under the hood.

## Setup and Install
-------

You will need to have go installed on your computer. Please refer to the [cosmos testnet tutorial](https://cosmos.network/validators/tutorial), which will always have the most updated instructions on how to get setup with go and the cosmos repository. 

Once you have go installed, run the command:

```
go get github.com/cosmos/cosmos-sdk
```

There will be an error stating `can't load package: package github.com/cosmos/cosmos-sdk: no Go files`, however you can ignore this error, it doesn't affect us. Now change directories to:

`cd $GOPATH/src/github.com/cosmos/cosmos-sdk`

And run :

```
make get_tools // run $ make update_tools if already installed
make get_vendor_deps
make install_examples
```
In this case we run `make install_examples`, which creates binaries for `basecli` and `basecoind`. 

##Using basecli and basecoind

Check the versions by running:

```
basecli version
basecoind version
```

They should read something like `0.17.1-5d18d5f`, but the versions will be constantly updating so don't worry if your version is higher that 0.17.1.

Note that you can always check help in the terminal by running `basecli -h` or `basecoind -h`. It is good to check these out if you are stuck, because updates to the code base might slighty change the commands, and you might find the correct command in there. 

Let's start by initializing the basecoind daemon. Run the command

`basecoind init`

And you should see something like this:

```
{
  "chain_id": "test-chain-z77iHG",
  "node_id": "e14c5056212b5736e201dd1d64c89246f3288129",
  "app_message": {
    "secret": "pluck life bracket worry guilt wink upgrade olive tilt output reform census member trouble around abandon"
  }
}
```

This creates the ~/.basecoind folder, which has config.toml, genesis.json, node_key.json, priv_validator.json. Take some time to review what is contained in these files if you want to understand what is going on at a deeper level. 


# Generating keys

The next thing we'll need to do is add the key from priv_validator.json to the gaiacli key manager. For this we need a 16 word seed and a password. You can also get the 16 word seed from the output seen above, under `"secret"`. Then run the command:

`basecli keys add alice --recover`

Which will give you three prompts:

Enter a passphrase for your key:
Repeat the passphrase:
Enter your recovery seed phrase:

NAME:	ADDRESS:					                PUBKEY:
alice	90B0B9BE0914ECEE0B6DB74E67B07A00056B9BBD	1624DE62201D47E63694448665F5D0217EA8458177728C91C373047A42BD3C0FB78BD0BFA7
bob	    29D721F054537C91F618A0FDBF770DA51EF8C48D	1624DE6220F54B2A2CA9EB4EE30DE23A73D15902E087C09CC5616456DDDD3814769E2E0A16
charlie	2E8E13EEB8E3F0411ACCBC9BE0384732C24FBD5E	1624DE6220F8C9FB8B07855FD94126F88A155BD6EB973509AE5595EFDE1AF05B4964836A53

Creating you first locally saved key creates the ~/.basecli folder which holds the keys you are storing. Now that you have the key for alice, you can start up the blockchain by running 

`basecoind start`

You should see blocks start getting created at a fast rate, with a lot of output in the terminal.

Next we need to make some more keys so we can send them some tokens. Open a new terminal, and run the following commands, to make two new accounts, and give each account a password you can remember:

`basecli keys add bob`
`basecli keys add charlie`

You can see your keys with the command:

`basecli keys list`

You should now see alice, bob and charlie's account all show up. 

# Send transactions

Lets send bob and charlie some tokens. First, lets query alice's account so we can see what kind of tokens she has:

`basecli account 90B0B9BE0914ECEE0B6DB74E67B07A00056B9BBD`

Where `90B0B9BE0914ECEE0B6DB74E67B07A00056B9BBD` is alices address we got from running `basecli keys list`. You should see a large amount of "mycoin" there. If you search the bob or charlies address, the command will fail, because they haven't been added into the blockchain database yet since they have no coins. We need to send them some!

The following command will send coins from alice, to bob:

`basecli send --name=alice --amount=10000mycoin --to=29D721F054537C91F618A0FDBF770DA51EF8C48D --sequence=0 --chain-id=test-chain-AE4XQo`

Where
- name is the name you gave your key
- `mycoin` is the name of the token for this basecoin demo, initialized in the genesis.json file
- sequence is a tally of how many transactions have been made by this account. Sicne this is the first tx on this account, it is 0
- chain-id is the unique ID that helps tendermint identify which network to connect to. You can find it in the terminal output from the gaiad daemon in the header block , or in the genesis.json file  at `~/.basecoind/config/gensis.json`

Now if we check bobs account, it should have ``10000`` 'mycoin' :

`basecli account 29D721F054537C91F618A0FDBF770DA51EF8C48D`

Now lets send some from bob to charlie 

`basecli send --name=bob --amount=5000mycoin --to=2E8E13EEB8E3F0411ACCBC9BE0384732C24FBD5E --sequence=0 --chain-id=test-chain-AE4XQo`

Note how we use the ``--name`` flag to select a different account to send from.

Lets now try to send from bob back to alice:

`basecli send --name=bob --amount=3000mycoin --to=90B0B9BE0914ECEE0B6DB74E67B07A00056B9BBD --sequence=1 --chain-id=test-chain-AE4XQo`

Notice that the sequence is now 1, since we have already recorded bobs 1st transaction as sequnce 0. Also note the ``hash`` value in the response - this is the hash of the transaction. We can query for the transaction by this hash:

`basecli tx <HASH>`

It will return the details of the transaction hash, such as how many coins were send and to which address. 

That is the basic implementation of basecoin!


## Clean up the basecoind and basecli data

**WARNING:** Running these commands will wipe out any existing
information in both the ``~/.basecli`` and ``~/.basecoind`` directories,
including private keys.

To remove all the files created and refresh your environment (e.g., if
starting this tutorial again or trying something new), the following
commands are run:

```
basecoind unsafe_reset_all
rm -rf ~/.basecoind
rm -rf ~/.basecli
```


