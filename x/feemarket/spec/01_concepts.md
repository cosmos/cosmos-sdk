<!--
order: 1
-->

# Concepts

Adopt EIP 1559 idea into cosmos-sdk.

- Maintain a minimal gas prices in consensus, which is ajusted based on total gas used in the previous block. It's called base gas price, to distinguish with the `minimal-gas-prices` configured in each node.
- In checkTx context, the `minimal-gas-prices` are also repected if it's bigger than the consensus one.