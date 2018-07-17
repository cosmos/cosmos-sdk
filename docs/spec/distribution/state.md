## State

### Pool
 - key: `01`
 - value: `amino(pool)`

The pool is a space for all dynamic global state of the Cosmos Hub.  It tracks
information about the total amounts of Atoms in all states, moving Atom
inflation information, etc.

```golang
type Pool struct {
    LooseTokens         int64   // tokens not associated with any bonded validator
    BondedTokens        int64   // reserve of bonded tokens
    InflationLastTime   int64   // block which the last inflation was processed // TODO make time
    Inflation           sdk.Rat // current annual inflation rate
    
    DateLastCommissionReset int64  // unix timestamp for last commission accounting reset (daily)
}
_______________________________________

### Validator

Validators are identified according to the `ValOwnerAddr`, 
an SDK account address for the owner of the validator.

Validators also have a `ValTendermintAddr`, the address 
of the public key of the validator.

Validators are indexed in the store using the following maps:
 
 - Validators: `0x02 | ValOwnerAddr -> amino(validator)`
 - ValidatorsByPubKey: `0x03 | ValTendermintAddr -> ValOwnerAddr`
 - ValidatorsByPower: `0x05 | power | blockHeight | blockTx  -> ValOwnerAddr`

* Adjustment factor used to passively calculate each validators entitled fees
  from `GlobalState.FeePool`

Delegation Shares

* AdjustmentFeePool: Adjustment factor used to passively calculate each bonds
  entitled fees from `GlobalState.FeePool`
* AdjustmentRewardPool: Adjustment factor used to passively calculate each
  bonds entitled fees from `Validator.ProposerRewardPool`

### Power Change

 - Key: `0x03 | amino(nonce)`

Every instance that the voting power changes, information about the state of
the validator set during the change must be recorded as a `PowerChange` for
other validators to run through. 


```golang
type PCNonce int64 

type PowerChange struct {
    height      int64        // block height at change
    power       rational.Rat // total power at change
    prevpower   rational.Rat // total power at previous height-1 
    feesIn      coins.Coin   // fees-in at block height
    prevFeePool coins.Coin   // total fees in at previous block height
}
```

### Max Power Change Nonce
 - key: `0x04`
 - value: `amino(PCNonce)`

To track the height of the power change, a nonce The set of all `powerChange`
may be trimmed from its oldest members once all validators have synced past the
height of the oldest `powerChange`.  This trim procedure will occur on an epoch
basis.  

