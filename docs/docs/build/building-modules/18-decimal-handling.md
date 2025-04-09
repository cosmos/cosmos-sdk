---
sidebar_position: 1
---

# Decimal Handling in Cosmos SDK

## Introduction

In the Cosmos SDK, the primary decimal type is `LegacyDec`.  
A newer decimal type called `Dec` was previously introduced with the intention to replace `LegacyDec`, offering enhanced performance and precision. However, this change has **since been reverted**, and `LegacyDec` remains the default and supported decimal type in the SDK.

This document exists for **historical reference** and to assist developers who may encounter the `Dec` type in experimental branches or custom forks.

## Background on `Dec` (Historical Context)

The implementation of `Dec` was adapted from Regen Network's [`regen-ledger`](https://github.com/regen-network/regen-ledger/tree/main/types/math), aiming to:

- Provide flexible precision up to 34 decimal places (vs. `LegacyDec`'s fixed 18).
- Leverage the [`apd`](https://github.com/cockroachdb/apd) library for accurate, arbitrary-precision decimal arithmetic.
- Improve performance and avoid mutations in arithmetic operations.

While these benefits appeared promising, they remained theoretical within the Cosmos SDK context and were ultimately deemed misaligned with the long-term stability and compatibility goals of the ecosystem.

## Why `Dec` Was Proposed (But Not Adopted)

- `LegacyDec` wraps `big.Int` and has limited precision control.
- `Dec` promised more efficient handling, better safety in concurrent environments, and better suited for financial calculations.
- However, introducing `Dec` required data format changes and state migrations, leading to high upgrade complexity.

As of now, these trade-offs were deemed unnecessary and `Dec` has been removed.

## Converting Between `LegacyDec` and `Dec`

> ⚠️ Note: These conversion utilities are preserved for **historical or experimental** purposes only.  
> They should **not be used in production environments**, as `Dec` is no longer officially supported in the SDK.

```go
func LegacyDecToDec(ld LegacyDec) (Dec, error) {
    return NewDecFromString(ld.String())
}
