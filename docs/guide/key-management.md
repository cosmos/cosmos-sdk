# Key Management

Here we explain a bit how to real with your keys, using the `basecli keys` subcommand.

**TODO**

## Creating keys

Create the keys and store a key phrase. No other way to recover it.

```
SEED=$(echo 1234567890 | basecli keys new fred -o json | jq .seed | tr -d \")
echo $SEED
(echo qwertyuiop; echo $SEED stamp) | basecli keys recover oops
(echo qwertyuiop; echo $SEED) | basecli keys recover derf
basecli keys get fred -o json
basecli keys get derf -o json
```

You can type it in to recover... try to do this by hand.
