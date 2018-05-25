# Validator Set Changes

The Tendermint validator set may be updated by state transitions that run at
the beginning and end of every block. The Tendermint validator set may be
changed by validators either being revoked due to inactivity/unexpected
behaviour (covered in slashing) or changed in validator power.

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
            validator = getValidator(store, elem.PubKey)
            returnedCoins = removeShares(validator, elem.Shares)
            validator.RedelegatingShares -= elem.Shares 
            delegateWithValidator(TxDelegate(elem.NewValidator, returnedCoins), validator)
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
    validators = loadValidators(store)

    v1 = validators.Validators()
    v2 = updateVotingPower(validators).Validators()

    change = v1.validatorsUpdated(v2) // determine all updated validators between two validator sets
    return change

updateVotingPower(validators Validators):
    foreach validator in validators do
	    validator.VotingPower = (validator.IssuedDelegatorShares - validator.RedelegatingShares) * delegatorShareExRate(validator)	
	    
    validators.Sort()
	
    foreach validator in validators do
	    if validator is not in the first params.MaxVals  
	        validator.VotingPower = rational.Zero
	        if validator.Status == Bonded then bondedToUnbondedPool(validator Validator)
		
	    else if validator.Status == UnBonded then unbondedToBondedPool(validator)
                      
	saveValidator(store, c)
	
    return validators

unbondedToBondedPool(validator Validator):
    removedTokens = exchangeRate(gs.UnbondedShares, gs.UnbondedPool) * validator.GlobalStakeShares 
    gs.UnbondedShares -= validator.GlobalStakeShares
    gs.UnbondedPool -= removedTokens
	
    gs.BondedPool += removedTokens
    issuedShares = removedTokens / exchangeRate(gs.BondedShares, gs.BondedPool)
    gs.BondedShares += issuedShares
    
    validator.GlobalStakeShares = issuedShares
    validator.Status = Bonded

    return transfer(address of the unbonded pool, address of the bonded pool, removedTokens)
```

