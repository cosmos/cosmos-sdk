# Validator Set Changes

The validator set may be updated by state transitions that run at the beginning and
end of every block. This can happen one of three ways:

- voting power of a validator changes due to bonding and unbonding
- voting power of validator is "slashed" due to conflicting signed messages
- validator is automatically unbonded due to inactivity

## Voting Power Changes

At the end of every block, we run the following:

(TODO remove inflation from here)

```golang
tick(ctx Context):
    hrsPerYr = 8766   // as defined by a julian year of 365.25 days
    
    time = ctx.Time()
    if time > gs.InflationLastTime + ProvisionTimeout 
        gs.InflationLastTime = time
        gs.Inflation = nextInflation(hrsPerYr).Round(1000000000)
        
        provisions = gs.Inflation * (gs.TotalSupply / hrsPerYr)
        
        gs.BondedPool += provisions
        gs.TotalSupply += provisions
        
        saveGlobalState(store, gs)
    
    if time > unbondDelegationQueue.head().InitTime + UnbondingPeriod 
        for each element elem in the unbondDelegationQueue where time > elem.InitTime + UnbondingPeriod do
    	    transfer(unbondingQueueAddress, elem.Payout, elem.Tokens)
    	    unbondDelegationQueue.remove(elem)
    
    if time > reDelegationQueue.head().InitTime + UnbondingPeriod 
        for each element elem in the unbondDelegationQueue where time > elem.InitTime + UnbondingPeriod do
            candidate = getCandidate(store, elem.PubKey)
            returnedCoins = removeShares(candidate, elem.Shares)
            candidate.RedelegatingShares -= elem.Shares 
            delegateWithCandidate(TxDelegate(elem.NewCandidate, returnedCoins), candidate)
            reDelegationQueue.remove(elem)
            
    return UpdateValidatorSet()

nextInflation(hrsPerYr rational.Rat):
    if gs.TotalSupply > 0 
        bondedRatio = gs.BondedPool / gs.TotalSupply
    else 
        bondedRation = 0
   
    inflationRateChangePerYear = (1 - bondedRatio / params.GoalBonded) * params.InflationRateChange
    inflationRateChange = inflationRateChangePerYear / hrsPerYr

    inflation = gs.Inflation + inflationRateChange
    if inflation > params.InflationMax then inflation = params.InflationMax
	
    if inflation < params.InflationMin then inflation = params.InflationMin
	
    return inflation 

UpdateValidatorSet():
    candidates = loadCandidates(store)

    v1 = candidates.Validators()
    v2 = updateVotingPower(candidates).Validators()

    change = v1.validatorsUpdated(v2) // determine all updated validators between two validator sets
    return change

updateVotingPower(candidates Candidates):
    foreach candidate in candidates do
	    candidate.VotingPower = (candidate.IssuedDelegatorShares - candidate.RedelegatingShares) * delegatorShareExRate(candidate)	
	    
    candidates.Sort()
	
    foreach candidate in candidates do
	    if candidate is not in the first params.MaxVals  
	        candidate.VotingPower = rational.Zero
	        if candidate.Status == Bonded then bondedToUnbondedPool(candidate Candidate)
		
	    else if candidate.Status == UnBonded then unbondedToBondedPool(candidate)
                      
	saveCandidate(store, c)
	
    return candidates

unbondedToBondedPool(candidate Candidate):
    removedTokens = exchangeRate(gs.UnbondedShares, gs.UnbondedPool) * candidate.GlobalStakeShares 
    gs.UnbondedShares -= candidate.GlobalStakeShares
    gs.UnbondedPool -= removedTokens
	
    gs.BondedPool += removedTokens
    issuedShares = removedTokens / exchangeRate(gs.BondedShares, gs.BondedPool)
    gs.BondedShares += issuedShares
    
    candidate.GlobalStakeShares = issuedShares
    candidate.Status = Bonded

    return transfer(address of the unbonded pool, address of the bonded pool, removedTokens)
```


## Slashing

Messges which may compromise the safety of the underlying consensus protocol ("equivocations")
result in some amount of the offending validator's shares being removed ("slashed").

Currently, such messages include only the following:

- prevotes by the same validator for more than one BlockID at the same
  Height and Round 
- precommits by the same validator for more than one BlockID at the same
  Height and Round 

We call any such pair of conflicting votes `Evidence`. Full nodes in the network prioritize the 
detection and gossipping of `Evidence` so that it may be rapidly included in blocks and the offending
validators punished.

For some `evidence` to be valid, it must satisfy: 

`evidence.Timestamp >= block.Timestamp - MAX_EVIDENCE_AGE`

where `evidence.Timestamp` is the timestamp in the block at height
`evidence.Height` and `block.Timestamp` is the current block timestamp.

If valid evidence is included in a block, the offending validator loses
a constant `SLASH_PROPORTION` of their current stake at the beginning of the block:

```
oldShares = validator.shares
validator.shares = oldShares * (1 - SLASH_PROPORTION)
```

This ensures that offending validators are punished the same amount whether they
act as a single validator with X stake or as N validators with collectively X
stake.


## Automatic Unbonding

Every block includes a set of precommits by the validators for the previous block, 
known as the LastCommit. A LastCommit is valid so long as it contains precommits from +2/3 of voting power.

Proposers are incentivized to include precommits from all
validators in the LastCommit by receiving additional fees
proportional to the difference between the voting power included in the
LastCommit and +2/3 (see [TODO](https://github.com/cosmos/cosmos-sdk/issues/967)).

Validators are penalized for failing to be included in the LastCommit for some
number of blocks by being automatically unbonded.

The following information is stored with each validator candidate, and is only non-zero if the candidate becomes an active validator:

```go
type ValidatorSigningInfo struct {
	StartHeight				int64
	SignedBlocksBitArray	BitArray
}
```

Where:
* `StartHeight` is set to the height that the candidate became an active validator (with non-zero voting power).
* `SignedBlocksBitArray` is a bit-array of size `SIGNED_BLOCKS_WINDOW` that records, for each of the last `SIGNED_BLOCKS_WINDOW` blocks,
whether or not this validator was included in the LastCommit. It uses a `0` if the validator was included, and a `1` if it was not.
Note it is initialized with all 0s. 

At the beginning of each block, we update the signing info for each validator and check if they should be automatically unbonded:

```
h = block.Height
index = h % SIGNED_BLOCKS_WINDOW

for val in block.Validators:
	signInfo = val.SignInfo
	if val in block.LastCommit:
		signInfo.SignedBlocksBitArray.Set(index, 0)
	else 
		signInfo.SignedBlocksBitArray.Set(index, 1)

	// validator must be active for at least SIGNED_BLOCKS_WINDOW
	// before they can be automatically unbonded for failing to be 
	// included in 50% of the recent LastCommits
	minHeight = signInfo.StartHeight + SIGNED_BLOCKS_WINDOW
	minSigned = SIGNED_BLOCKS_WINDOW / 2
	blocksSigned = signInfo.SignedBlocksBitArray.Sum() 
	if h > minHeight AND blocksSigned < minSigned:
		unbond the validator
```
