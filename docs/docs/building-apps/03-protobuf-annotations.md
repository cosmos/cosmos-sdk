---
sidebar_position: 1
---

# ProtocolBuffer Annotations

This document explains the various protobuf scalars that have been added to make working with protobuf easier for Cosmos SDK application developers

## Scalars

(cosmos_proto.scalar) = "cosmos.AddressString";

## Interfaces

option (cosmos_proto.implements_interface) = "cosmos.auth.v1beta1.AccountI";

## Amino

### Name

option (amino.name) = "cosmos-sdk/BaseAccount";

### Field Name

(amino.field_name) = "public_key"

### Dont_OmitEmpty 

(amino.dont_omitempty)   = true,

### Encoding 

(amino.encoding)         = "legacy_coins",
