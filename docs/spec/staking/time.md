
### End of block handling

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
