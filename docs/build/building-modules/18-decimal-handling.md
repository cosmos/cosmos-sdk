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

* Historically we have wrapped a `big.Int` to represent decimals in the Cosmos SDK and never had a decimal type. Finally, we have a decimal type that is more efficient and accurate.
* `Dec` uses the [apd](https://github.com/cockroachdb/apd) library for arbitrary precision decimals, suitable for accurate financial calculations.
* `Dec` operations are safer for concurrent use as they do not mutate the original values.
* `Dec` operations are faster and more efficient than `LegacyDec`.

## Using `Dec` in Modules that haven't used `LegacyDec`

If you are creating a new module or updating an existing module that has not used `LegacyDec`, you can directly use `Dec`.
Ensure proper error handling.
  
```
-- math.NewLegacyDecFromInt64(100)
++ math.NewDecFromInt64(100)

-- math.LegacyNewDecWithPrec(100, 18)
++ math.NewDecWithPrec(100, 18)

-- math.LegacyNewDecFromStr("100")
++ math.NewDecFromString("100")

-- math.LegacyNewDecFromStr("100.000000000000000000").Quo(math.LegacyNewDecFromInt(2))
++ foo, err := math.NewDecFromString("100.000000000000000000")
++ foo.Quo(math.NewDecFromInt(2))

-- math.LegacyNewDecFromStr("100.000000000000000000").Add(math.LegacyNewDecFromInt(2))
++ foo, err := math.NewDecFromString("100.000000000000000000")
++ foo.Add(math.NewDecFromInt(2))

-- math.LegacyNewDecFromStr("100.000000000000000000").Sub(math.LegacyNewDecFromInt(2))
++ foo, err := math.NewDecFromString("100.000000000000000000")
++ foo.Sub(math.NewDecFromInt(2))
```

## Modules migrating from `LegacyDec` to `Dec`

When migrating from `LegacyDec` to `Dec`, you need to update your module to use the new decimal type. **These types are state breaking changes and require a migration.**

## Precision Handling

The reason for the state breaking change is the difference in precision handling between the two decimal types:

* **LegacyDec**: Fixed precision of 18 decimal places.
* **Dec**: Flexible precision up to 34 decimal places using the apd library.

## Impact of Precision Change

The increase in precision from 18 to 34 decimal places allows for more detailed decimal values but requires data migration. This change in how data is formatted and stored is a key aspect of why the transition is considered state-breaking.

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

