---
sidebar_position: 1
---

# `x/consensus`

## Abstract 

Functionality to modify CometBFT's ABCI consensus params.

## Contents

* [State](#state)
* [Params](#params)
* [Keepers](#keepers)
* [Messages](#messages)
* [Consensus Messages](#consensus-messages)
* [Events](#events)
    * [Message Events](#message-events)


## State

The `x/consensus` module keeps state of the consensus params from cometbft.:

## Params

The consensus module stores it's params in state with the prefix of `0x05`,
it can be updated with governance or the address with authority.

* Params: `0x05 | ProtocolBuffer(cometbft.ConsensusParams)`

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/release/v0.52.x/x/consensus/proto/cosmos/consensus/v1/consensus.proto#L9-L15
```

## Keepers

The consensus module provides methods to Set and Get consensus params. It is recommended to use the `x/consensus` module keeper to get consensus params instead of accessing them through the context.

## Messages

### UpdateParams

Update consensus params.

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/release/v0.52.x/x/consensus/proto/cosmos/consensus/v1/tx.proto#L23-L44
```

The message will fail under the following conditions:

* The signer is not the set authority 
* Not all values are set

## Events

The consensus module emits the following events:

### Message Events

#### MsgUpdateParams

| Type   | Attribute Key | Attribute Value     |
|--------|---------------|---------------------|
| string | authority     | msg.Signer          |
| string | parameters    | consensus Parameters |
