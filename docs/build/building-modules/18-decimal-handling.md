---
sidebar_position: 1
---
# Decimal Handling in Cosmos SDK

## Introduction

In the Cosmos SDK, there are two types of decimals: `LegacyDec` and `Dec`. `LegacyDec` is the older decimal type that is still available for use, while `Dec` is the newer, more performant decimal type. The implementation of `Dec` is adapted from Regen Network's `regen-ledger`, specifically from [this module](https://github.com/regen-network/regen-ledger/tree/main/types/math). Migrating from `LegacyDec` to `Dec` involves state-breaking changes, specifically:

* **Data Format**: The internal representation of decimals changes, affecting how data is stored and processed.
* **Precision Handling**: `Dec` supports flexible precision up to 34 decimal places, unlike `LegacyDec` which has a fixed precision of 18 decimal places.

These changes require a state migration to update existing decimal values to the new format. It is recommended to use `Dec` for new modules to leverage its enhanced performance and flexibility.

## Why the Change?

* **Enhanced Precision**: `Dec` uses the [apd](https://github.com/cockroachdb/apd) library for arbitrary precision decimals, suitable for accurate financial calculations.
* **Immutable Operations**: `Dec` operations are safer for concurrent use as they do not mutate the original values.
* **Better Performance**: `Dec` operations are faster and more efficient than `LegacyDec`.`

## Using `Dec` in Modules that haven't used `LegacyDec`

If you are creating a new module or updating an existing module that has not used `LegacyDec`, you can directly use `Dec` without any changes.

As an example we will use `DecCoin` which is a common type used in the Cosmos SDK.


```protobuf
message DecCoin {
  option (gogoproto.equal) = true;

  string denom  = 1;
  string amount = 2 [
    (cosmos_proto.scalar)  = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.Dec",
    (gogoproto.nullable)   = false
  ];
}
```

How you can implement `Dec` in your module:

```go
import (
	"cosmossdk.io/math"
)

example := math.NewDecFromInt64(100)
```

## Modules migrating from `LegacyDec` to `Dec`

When migrating from `LegacyDec` to `Dec`, you need to update your module to use the new decimal type. **These types are state breaking changes and require a migration.**

## Precision Handling

The reason for the state breaking change is the difference in precision handling between the two decimal types:

* **LegacyDec**: Fixed precision of 18 decimal places.
* **Dec**: Flexible precision up to 34 decimal places using the apd library.

## Byte Representation Changes Example

The change in precision handling directly impacts the byte representation of decimal values:

**Legacy Dec Byte Representation:**
`2333435363738393030303030303030303030303030303030303030`

This example includes the value 123456789 followed by 18 zeros to maintain the fixed precision.

**New Dec Byte Representation:**
`0a03617364121031323334353637383900000000000000`

This example shows the value 123456789 without additional padding, reflecting the flexible precision handling of the new Dec type.

## Impact of Precision Change

The increase in precision from 18 to 34 decimal places allows for more detailed decimal values but requires data migration. This change in how data is formatted and stored is a key aspect of why the transition is considered state-breaking.

## Example of State-Breaking Change

The protobuf definitions for DecCoin illustrate the change in the custom type for the amount field.

**Before:**

```protobuf
message DecCoin {
  option (gogoproto.equal) = true;

  string denom  = 1;
  string amount = 2 [
    (cosmos_proto.scalar)  = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable)   = false
  ];
}
```

**After:**

```protobuf
message DecCoin {
  option (gogoproto.equal) = true;

  string denom  = 1;
  string amount = 2 [
    (cosmos_proto.scalar)  = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.Dec",
    (gogoproto.nullable)   = false
  ];
}
```

## Converting `LegacyDec` to `Dec` without storing the data

If you would like to convert a `LegacyDec` to a `Dec` without a state migration changing how the data is handled internally within the application logic and not how it's stored or represented. You can use the following methods.

```go
func LegacyDecToDec(ld LegacyDec) (Dec, error) {
    return NewDecFromString(ld.String())
}
```

```go
func DecToLegacyDec(ld Dec) (LegacyDec, error) {
    return LegacyDecFromString(ld.String())
}
```

