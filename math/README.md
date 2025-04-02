# Cosmos SDK Math

This library contains types for preforming arbitrary precision integer and decimal arithmetic
which is essentially for performing accurate financial calculations. The main types provided
are `Int`, `Uint`, `Dec` and `LegacyDec`.

All of these types can be used as [gogoproto custom types](https://pkg.go.dev/github.com/cosmos/gogoproto/gogoproto).
Although gogoproto is a legacy library, it is still used in most Cosmos SDK modules and a sufficient
replacement for its custom types functionality does not exist yet.

## `Int` and `Uint`

The `Int` and `Uint` types are wrappers around the [`big.Int`](https://pkg.go.dev/math/big#Int) type. They provide a convenient API
for working with signed and unsigned integers that can be used as gogoproto custom types.

## `Dec` and `LegacyDec`

The `Dec` and `LegacyDec` types provide two fundamentally different approaches to representing
and performing high-precision decimal arithmetic. `Dec` is a wrapper around [`apd.Decimal`](https://pkg.go.dev/github.com/cockroachdb/apd/v3) and implements [General Decimal Arithmetic](https://speleotrove.com/decimal/)
with a precision of 34 significant decimal digits.

`LegacyDec` on the other hand is a wrapper around [`big.Int`](https://pkg.go.dev/math/big#Int) which provides a
fixed 18 digits of precision after the decimal point.

More documentation on using these types and how to migrate code from using `LegacyDec` to `Dec` can be found in
the [Decimal Handling in Cosmos SDK](https://docs.cosmos.network/main/build/building-modules/decimal-handling) documentation.
It is important to note that the serialization formats of `Dec` and `LegacyDec` are different,
so it is NOT SAFE to simply replace `LegacyDec` with `Dec` in your code and **state migrations
and API changes are needed**.
