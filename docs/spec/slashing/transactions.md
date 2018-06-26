
### TxProveLive

If a validator was automatically unbonded due to liveness issues and wishes to
assert it is still online, it can send `TxProveLive`:

```golang
type TxProveLive struct {
    PubKey crypto.PubKey
}
```

All delegators in the temporary unbonding pool which have not
transacted to move will be bonded back to the now-live validator and begin to
once again collect provisions and rewards. 

```
TODO: pseudo-code
```
