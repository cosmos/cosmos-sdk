<!--
order: 3
-->

# AnteHandlers

## Decorators

### `MempoolFeeDecorator`

The `MempoolFeeDecorator` in `x/auth` module should check the `BaseGasPrices` along with the `minimal-gas-prices`.

```golang
gas := tx.GetGas()
baseGP := GetState(BaseGasPrices)
minGP := ctx.MinGasPrices()  // it's zero in deliverTx context

requiredFees := make(sdk.Coins, 0)
if baseGP.IsZero() {
  // base gas prices are not enabled, check the minimal-gas-prices only
  for i, gp := range minGasPrices {
    fee := gp.Amount.Mul(gas).Ceil().RoundInt()
    requiredFees = append(requiredFees, sdk.NewCoin(gp.Denom, fee))
  }
} else {
  // check `tx.GetFee() > max(baseGP, minGP) * tx.GetGas()`
  // ignore the token types in `minimal-gas-prices` which are not specified in `BaseGasPrices`
  for i, gp := range baseGP {
    fee := gp.Amount.Mul(gas).Ceil().RoundInt()
    amt := minGasPrices.AmountOf(gp.Denom)
    if amt > fee {
      fee = amt
    }
    requiredFees = append(requiredFees, sdk.NewCoin(gp.Denom, fee))
  }
}

feeCoins := tx.GetFee()
if !feeCoins.IsAnyGTE(requiredFees) {
  return fmt.Errorf("insufficient fees; got: %s required: %s", feeCoins, requiredFees)
}
```
