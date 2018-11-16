# Begin-Block

## Inflation 

Inflation occurs at the beginning of each block, however inflation parameters
are only calculated once per hour.

### NextInflationRate

The target annual inflation rate is recalculated at the first block of each new
hour. The inflation is also subject to a rate change (positive or negative)
depending on the distance from the desired ratio (67%). The maximum rate change
possible is defined to be 13% per year, however the annual inflation is capped
as between 7% and 20%.

NextInflationRate(params Params, bondedRatio sdk.Dec) (inflation sdk.Dec) {
	inflationRateChangePerYear = (1 - bondedRatio/params.GoalBonded) * params.InflationRateChange
	inflationRateChange = inflationRateChangePerYear/hrsPerYr

	// increase the new annual inflation for this next cycle
	inflation += inflationRateChange
	if inflation > params.InflationMax {
		inflation = params.InflationMax
	}
	if inflation < params.InflationMin {
		inflation = params.InflationMin
	}

	return inflation

### NextHourlyProvisions

Rather than directly using the inflation rate to calculate minted tokens for each block, 
the minted tokens is calculated once per hour, and then applied to each _________________________________________________________

the provisions are calculated once per hour and further divided up each block based
on this hour value.
