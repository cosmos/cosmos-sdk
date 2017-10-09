# Basic run through of using basecli....

To keep things clear, let's have two shells...

`$` is for basecoin (server), `%` is for basecli (client)

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
% ME=$(basecli keys get demo | awk '{print $2}')
% YOU=$(basecli keys get buddy | awk '{print $2}')
```

## Set up a clean basecoin, initialized with your account

```
$ export BCHOME=~/.demoserve
$ basecoin init $ME
$ basecoin start
```

## Connect your basecli the first time

```
% basecli init --chain-id test_chain_id --node tcp://localhost:46657
```

## Check your balances...

```
% basecli query account $ME
% basecli query account $YOU
```

## Send the money

```
% basecli tx send --name demo --amount 1000strings --sequence 1 --to $YOU
-> copy hash to HASH
% basecli query tx $HASH
% basecli query account $YOU
```

