<!--
order: 2
-->

# State

## Minter

The minter is a space for holding current inflation information.

 - Minter: `0x00 -> amino(minter)`

```go
type Minter struct {
	Inflation          sdk.Dec   // current annual inflation rate
	AnnualProvisions   sdk.Dec   // current annual exptected provisions
	LastBlockTimestamp time.Time // the timestamp of last processed block
	AverageBlockTime   time.Duration // the cumulated average of block creation time
}
```

## Params

Minting params are held in the global params store. 

 - Params: `mint/params -> amino(params)`

```go
type Params struct {
	MintDenom           string  // type of coin to mint
	InflationRateChange sdk.Dec // maximum annual change in inflation rate
	InflationMax        sdk.Dec // maximum inflation rate
	InflationMin        sdk.Dec // minimum inflation rate
	GoalBonded          sdk.Dec // goal of percent bonded atoms
}
```
