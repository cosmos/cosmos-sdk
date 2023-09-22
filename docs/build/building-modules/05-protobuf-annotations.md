---
sidebar_position: 1
---

# ProtocolBuffer Annotations

This document explains the various protobuf scalars that have been added to make working with protobuf easier for Cosmos SDK application developers

## Signer

Signer specifies which field should be used to determine the signer of a message for the Cosmos SDK. This field can be used for clients as well to infer which field should be used to determine the signer of a message.

Read more about the signer field [here](./02-messages-and-queries.md).

```protobuf reference 
https://github.com/cosmos/cosmos-sdk/blob/e6848d99b55a65d014375b295bdd7f9641aac95e/proto/cosmos/bank/v1beta1/tx.proto#L40
```

```proto
option (cosmos.msg.v1.signer) = "from_address";
```

## Scalar

The scalar type defines a way for clients to understand how to construct protobuf messages according to what is expected by the module and sdk.

```proto
(cosmos_proto.scalar) = "cosmos.AddressString"
```

Example of account address string scalar:

```proto reference 
https://github.com/cosmos/cosmos-sdk/blob/e6848d99b55a65d014375b295bdd7f9641aac95e/proto/cosmos/bank/v1beta1/tx.proto#L46
```

Example of validator address string scalar: 

```proto reference 
https://github.com/cosmos/cosmos-sdk/blob/e8f28bf5db18b8d6b7e0d94b542ce4cf48fed9d6/proto/cosmos/distribution/v1beta1/query.proto#L87
```

Example of Decimals scalar: 

```proto reference
https://github.com/cosmos/cosmos-sdk/blob/e8f28bf5db18b8d6b7e0d94b542ce4cf48fed9d6/proto/cosmos/distribution/v1beta1/distribution.proto#L26
```

Example of Int scalar: 

```proto reference
https://github.com/cosmos/cosmos-sdk/blob/e8f28bf5db18b8d6b7e0d94b542ce4cf48fed9d6/proto/cosmos/gov/v1/gov.proto#L137
```

There are a few options for what can be provided as a scalar: cosmos.AddressString, cosmos.ValidatorAddressString, cosmos.ConsensusAddressString, cosmos.Int, cosmos.Dec. 

## Implements_Interface

Implement interface is used to provide information to client tooling like [telescope](https://github.com/cosmology-tech/telescope) on how to encode and decode protobuf messages. 

```proto
option (cosmos_proto.implements_interface) = "cosmos.auth.v1beta1.AccountI";
```

## Amino

The amino codec was removed in 0.50.0, this means there is not a need register `legacyAminoCodec`. To replace the amino codec, Amino protobuf annotations are used to provide information to the amino codec on how to encode and decode protobuf messages. 

:::note
Amino annotations are only used for backwards compatibility with amino. New modules are not required use amino annotations.
:::

The below annotations are used to provide information to the amino codec on how to encode and decode protobuf messages in a backwards compatible manner. 

### Name

Name specifies the amino name that would show up for the user in order for them see which message they are signing.

```proto
option (amino.name) = "cosmos-sdk/BaseAccount";
```

```proto reference
https://github.com/cosmos/cosmos-sdk/blob/e8f28bf5db18b8d6b7e0d94b542ce4cf48fed9d6/proto/cosmos/bank/v1beta1/tx.proto#L41
```

### Field_Name

Field name specifies the amino name that would show up for the user in order for them see which field they are signing.

```proto
uint64 height = 1 [(amino.field_name) = "public_key"];
```

```proto reference
https://github.com/cosmos/cosmos-sdk/blob/e8f28bf5db18b8d6b7e0d94b542ce4cf48fed9d6/proto/cosmos/distribution/v1beta1/distribution.proto#L166
```

### Dont_OmitEmpty 

Dont omitempty specifies that the field should not be omitted when encoding to amino. 

```proto
repeated cosmos.base.v1beta1.Coin amount = 3 [(amino.dont_omitempty)   = true];
```

```proto reference
https://github.com/cosmos/cosmos-sdk/blob/e8f28bf5db18b8d6b7e0d94b542ce4cf48fed9d6/proto/cosmos/bank/v1beta1/bank.proto#L56
```

### Encoding 

Encoding specifies the amino encoding that should be used when encoding to amino. 

```proto
(amino.encoding)         = "legacy_coins",
```

```proto reference
https://github.com/cosmos/cosmos-sdk/blob/e8f28bf5db18b8d6b7e0d94b542ce4cf48fed9d6/proto/cosmos/bank/v1beta1/genesis.proto#L23
```
