## State

### Minter

The minter is a space for holding current inflation information.

 - Minter: `0x00 -> amino(minter)`

```golang
type Minter struct {
	InflationLastTime time.Time // block time which the last inflation was processed
	Inflation         sdk.Dec   // current annual inflation rate
}
```

### Params

Minting params are held in the global params store. 

 - Params: `mint/params -> amino(params)`

```golang
type Params struct {
	MintDenom           string  // type of coin to mint
	InflationRateChange sdk.Dec // maximum annual change in inflation rate
	InflationMax        sdk.Dec // maximum inflation rate
	InflationMin        sdk.Dec // minimum inflation rate
	GoalBonded          sdk.Dec // goal of percent bonded atoms
}
```

