# End Block

Validator provisions are minted on an hourly basis (the first block of a new
hour). The annual target of between 7% and 20%. The long-term target ratio of
bonded tokens to unbonded tokens is 67%.

The target annual inflation rate is recalculated for each provisions cycle. The
inflation is also subject to a rate change (positive or negative) depending on
the distance from the target ratio (67%). The maximum rate change possible is
defined to be 13% per year, however the annual inflation is capped as between
7% and 20%.

Within the inflation module the tokens are created, and fed to the distribution 
module to be further processed and distributed similarly to fee distribution (with 
the exception that there are no special rewards for the block proposer)

Note that params are global params (TODO: link to the global params spec)

```
EndBlock(): 

    //process provisions
    hrsPerYr = 8766   // as defined by a julian year of 365.25 days
    precision = 10000

    time = BFTTime() // time is in seconds
    if time > GetInflationLastTime() + 3600 
        SetInflationLastTime(InflationLastTime + 3600)
        inflation = nextInflation(hrsPerYr).Round(precision)
        SetInflation(inflation)
        
        provisions = inflation * (pool.TotalSupply() / hrsPerYr)
        pool.LooseTokens += provisions
        
        distribution.AddInflation(provisions)

nextInflation(hrsPerYr rational.Rat):

    bondedRatio = pool.BondedPool / pool.TotalSupply()

    inflationRateChangePerYear = (1 - bondedRatio / params.GoalBonded) * params.InflationRateChange
    inflationRateChange = inflationRateChangePerYear / hrsPerYr

    inflation = GetInflation() + inflationRateChange
    switch inflation
        case > params.InflationMax
            return params.InflationMax
        case < params.InflationMin
            return params.InflationMin
        default 	
            return inflation 
```
