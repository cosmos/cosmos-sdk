# Basic run through of using basecli....

To keep things clear, let's have two shells...

`$` is for basecoin (server), `%` is for basecli (client)

## Set up a clean basecoin, but don't start the chain

```
$ export BCHOME=~/.demoserve
$ basecoin init
```

## Set up your basecli with a new key

```
% export BCHOME=~/.democli
% basecli keys new demo
% basecli keys get demo -o json
```

And set up a few more keys for fun...

```
% basecli keys new buddy
% basecli keys list
% ME=`basecli keys get demo -o json | jq .address | tr -d '"'`
% YOU=`basecli keys get buddy -o json | jq .address | tr -d '"'`
```

## Update genesis so you are rich, and start

```
$ vi $BCHOME/genesis.json
-> cut/paste your pubkey from the results above

$ basecoin start
```

## Connect your basecli the first time

```
% basecli init --chainid test_chain_id --node localhost:45567
```

## Check your balances...

```
% basecli proof state get --app=account --key=$ME
% basecli proof state get --app=account --key=$YOU
```

## Send the money

```
% basecli tx send --name demo --amount 1000mycoin --sequence 1 --to $YOU
-> copy hash to HASH
% basecli proof tx get --key $HASH

% basecli proof tx get --key $HASH --app base
% basecli proof state get --key $YOU --app account
```

## Any questions???

```
% basecli seeds show --height 1767
```
