<!--
order: 3
-->

# State Transitions

## Send Fungible Tokens

A successful fungible token send has two state transitions depending if the
transfer is a movement forward or backwards in the token's timeline:

1. Sender chain is the source chain, *i.e* a transfer to any chain other than the one it was previously received from is a movement forwards in the token's timeline. This results in the following state transitions:

- The coins are transferred to an escrow address (i.e locked) on the sender chain
- The coins are transferred to the receiving chain through IBC TAO logic.

2. Sender chain is the sink chain, *i.e* the token is sent back to the chain it previously received from. This is a backwards movement in the token's timeline. This results in the following state transitions:

- The coins (vouchers) are burned on the sender chain
- The coins transferred to the receiving chain though IBC TAO logic.

## Receive Fungible Tokens

A successful fungible token receive has two state transitions depending if the
transfer is a movement forward or backwards in the token's timeline:

1. Receiver chain is the source chain. This is a backwards movement in the token's timeline. This results in the following state transitions:

- The leftmost port and channel identifier pair is removed from the token denomination prefix.
- The tokens are unescrowed and sent to the receiving address.

2. Receiver chain is the sink chain. This is a movement forwards in the token's timeline. This results in the following state transitions:

- Token vouchers are minted by prefixing the destination port and channel identifiers to the trace information.
- The receiving chain stores the new trace information in the store (if not set already).
- The vouchers are sent to the receiving address.
