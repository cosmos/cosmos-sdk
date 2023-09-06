---
sidebar_position: 1
---

# ProtocolBuffer Annotations

This document explains the various protobuf scalars that have been added to make working with protobuf easier for Cosmos SDK application developers

## Signer

Signer specifies which field should be used to determine the signer of a message for the Cosmos SDK. This field can be used for clients as well to infer which field should be used to determine the signer of a message.

Read more about the signer field [here](../building-modules/02-messages-and-queries.md).

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

There are a few options for what can be provided as a scalar: cosmos.AddressString, cosmos.ValidatorAddressString, cosmos.ConsensusAddressString, cosmos.Int, cosmos.Dec. 

## Implements_Interface

Implement interface is used to provide information to client tooling like [telescope](https://github.com/cosmology-tech/telescope) on how to encode and decode protobuf messages. 

```proto
option (cosmos_proto.implements_interface) = "cosmos.auth.v1beta1.AccountI";
```

## Amino

The amino codec was removed in 0.50.0, this means there is not a need register `legacyAminoCodec`. To replace the amino codec, Amino protobuf annotations are used to provide information to the amino codec on how to encode and decode protobuf messages. 

:::Note
Amino annotations are only used for backwards compatibility with amino. New modules should not use amino annotations. The 
:::

The below annotations are used to provide information to the amino codec on how to encode and decode protobuf messages in a backwards compatible manner. 

### Name

Name specifies the amino name that would show up for the user in order for them see which message they are signing.

```proto
option (amino.name) = "cosmos-sdk/BaseAccount";
```

### Field_Name

Field name specifies the amino name that would show up for the user in order for them see which field they are signing.

```proto
(amino.field_name) = "public_key"
```

### Dont_OmitEmpty 

Dont omitempty specifies that the field should not be omitted when encoding to amino. 

```proto
(amino.dont_omitempty)   = true,
```

### Encoding 

Encoding specifies the amino encoding that should be used when encoding to amino. 

```proto
(amino.encoding)         = "legacy_coins",
```
