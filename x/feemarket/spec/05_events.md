<!--
order: 5 -->

# Events

The `x/feemarket` module emits the following events:

## EndBlocker

| Type       | Attribute Key   | Attribute Value |
| ---------- | --------------- | --------------- |
| block_gas  | height          | {blockHeight}   |
| block_gas  | amount          | {blockGasUsed}  |
| fee_market | base_gas_prices | {baseGasPrices} |