# End-Block 

Two staking activities are intended to be processed in the application end-block.
 - inform Tendermint of validator set changes
 - process and set atom inflation

# Validator Set Changes

The Tendermint validator set may be updated by state transitions that run at
the end of every block. The Tendermint validator set may be changed by
validators either being revoked due to inactivity/unexpected behaviour (covered
in slashing) or changed in validator power. Determining which validator set
changes must be made occurs during staking transactions (and slashing
transactions) - during end-block the already accounted changes are applied and
the changes cleared

```golang
EndBlock() ValidatorSetChanges
    vsc = GetTendermintUpdates()
    ClearTendermintUpdates()
    return vsc
```

# Inflation

The atom inflation rate is changed once per hour based on the current and
historic bond ratio

```golang
processProvisions():
    hrsPerYr = 8766   // as defined by a julian year of 365.25 days
    
    time = BFTTime()
    if time > pool.InflationLastTime + ProvisionTimeout 
        pool.InflationLastTime = time
        pool.Inflation = nextInflation(hrsPerYr).Round(1000000000)
        
        provisions = pool.Inflation * (pool.TotalSupply / hrsPerYr)
        
        pool.LooseUnbondedTokens += provisions
        feePool += LooseUnbondedTokens
        
        setPool(pool)

nextInflation(hrsPerYr rational.Rat):
    if pool.TotalSupply > 0 
        bondedRatio = pool.BondedPool / pool.TotalSupply
    else 
        bondedRation = 0
   
    inflationRateChangePerYear = (1 - bondedRatio / params.GoalBonded) * params.InflationRateChange
    inflationRateChange = inflationRateChangePerYear / hrsPerYr

    inflation = pool.Inflation + inflationRateChange
    if inflation > params.InflationMax then inflation = params.InflationMax
	
    if inflation < params.InflationMin then inflation = params.InflationMin
	
    return inflation 
```

