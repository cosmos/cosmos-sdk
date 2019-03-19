# State

## ConstantFee

The ConstantFee params are held in the global params store. 

 - Params: `mint/params -> amino(params)`

```golang
type Params struct {
	MintDenom           string  // type of coin to mint
	InflationRateChange sdk.Dec // maximum annual change in inflation rate
	InflationMax        sdk.Dec // maximum inflation rate
	InflationMin        sdk.Dec // minimum inflation rate
	GoalBonded          sdk.Dec // goal of percent bonded atoms
	BlocksPerYear       uint64   // expected blocks per year
}
```
