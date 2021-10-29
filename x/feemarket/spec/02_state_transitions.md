<!--
order: 2
-->

# State Transitions

This document describes the state transition operations pertaining to the base gas prices.

## Block Gas Used

The total gas used by current block is stored in chain state at the `EndBlock` event.

It's initialized to `-1` in `InitGenesis`.

## Base Gas Prices

### Init

The base gas prices are initialized to the parameter `InitialBasePrices` in `InitGenesis` event.

### Adjust

Base gas prices are adjusted in the `EndBlock` event according to the total gas used in the previous block.

```golang
parentGP := GetState(BaseGasPrices)
parentGasUsed := GetState(BlockGasUsed)

var expectedGP sdk.Coins
if parentGasUsed == -1:
  expectedGP = parentGP
else if parentGasUsed == params.BlockGasTarget:
	expectedGP = parentGP
else if parentGasUsed > params.BlockGasTarget:
	delta = parentGasUsed - params.BlockGasTarget
	delta = max(parentBaseFee * delta / params.BlockGasTarget / params.BaseGasPriceChangeDenominator, 1)
	expectedGP = parentBaseFee + delta
else:
	delta =  params.BlockGasTarget - parentGasUsed
	delta = parentGP * delta / params.BlockGasTarget / params.BaseGasPriceChangeDenominator
	expectedGP = parentGP - delta

SetState(BaseGasPrices, expectedGP)
```

