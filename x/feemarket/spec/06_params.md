<!--
order: 6 -->

# Parameters

The `x/feemarket` module contains the following parameters:

| Key                           | Type   | Example     |
| ----------------------------- | ------ | ----------- |
| BlockGasTarget                | uint32 | "2000000"   |
| InitialBaseGasPrices          | Coins  | "1000uatom" |
| BaseGasPriceChangeDenominator | uint32 | 8           |

- `BlockGasTarget`, should be smaller than `ConsensusParams.Block.MaxGas` if the latter one is specified.