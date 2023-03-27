---
sidebar_position: 1
---

# ProtocolBuffer Annotations

This document explains the various protobuf scalars that have been added to make working with protobuf easier for Cosmos SDK application developers

## Scalar

(cosmos_proto.scalar) = "cosmos.AddressString";

## Implements_Interface

Implement interface is 

option (cosmos_proto.implements_interface) = "cosmos.auth.v1beta1.AccountI";

## Amino

### Name

option (amino.name) = "cosmos-sdk/BaseAccount";

### Field_Name

(amino.field_name) = "public_key"

### Dont_OmitEmpty 

(amino.dont_omitempty)   = true,

### Encoding 

(amino.encoding)         = "legacy_coins",
