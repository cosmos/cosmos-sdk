<!--
order: 1
-->

# Concepts

Adopt EIP 1559 idea into cosmos-sdk.

- Maintain a minimal gas price in consensus, which is adjusted based on total gas used in the previous block. It's called base gas price, to distinguish with the `minimal-gas-prices` configured in each node.
- In checkTx context, the `minimal-gas-prices` are also respected if it's bigger than the consensus one.