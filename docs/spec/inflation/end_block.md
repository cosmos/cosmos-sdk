# End Block

```
EndBlock() 
    processProvisions()
```

# Inflation

Validator provisions are minted on an hourly basis (the first block of a new
hour). The annual target of between 7% and 20%. The long-term target ratio of
bonded tokens to unbonded tokens is 67%.

The target annual inflation rate is recalculated for each provisions cycle. The
inflation is also subject to a rate change (positive or negative) depending on
the distance from the desired ratio (67%). The maximum rate change possible is
defined to be 13% per year, however the annual inflation is capped as between
7% and 20%.

Within the inflation module the tokens are created, and fed to the distribution 
module to be further processed and distributed similarly to fee distribution (with 
the exception that there are no special rewards for the block proposer)

Note that params are global params (TODO: link to the global params spec)

```
processProvisions():
    hrsPerYr = 8766   // as defined by a julian year of 365.25 days
    
    time = BFTTime()
    if time > GetInflationLastTime() + OneHour
        SetInflationLastTime(InflationLastTime + OneHour)
        inflation = nextInflation(hrsPerYr).Round(1000000000)
        SetInflation(inflation)
        
        provisions = inflation * (pool.TotalSupply() / hrsPerYr)
        pool.LooseTokens += provisions
        
        distribution.AddInflation(provisions)

nextInflation(hrsPerYr rational.Rat):
    if pool.TotalSupply() > 0 
        panic("Total supply must be greater than 0")

    bondedRatio = pool.BondedPool / pool.TotalSupply()

    inflationRateChangePerYear = (1 - bondedRatio / params.GoalBonded) * params.InflationRateChange
    inflationRateChange = inflationRateChangePerYear / hrsPerYr

    inflation = GetInflation() + inflationRateChange
    if inflation > params.InflationMax then inflation = params.InflationMax
    if inflation < params.InflationMin then inflation = params.InflationMin
	
    return inflation 
```
