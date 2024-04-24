---
sidebar_position: 1
---
# Decimal Handling in Cosmos SDK

:::note
As part of ongoing improvements to the Cosmos SDK, we have updated our decimal handling from `LegacyDec` to `Dec`. This update is crucial for modules that perform mathematical computations, ensuring higher precision and better performance.
:::

## Introduction

In the Cosmos SDK we have 2 types of decimals LegacyDec and Dec. `LegacyDec` is the old decimal type that was used, which is still available. `Dec` is the new decimal type and is more performant than `LegacyDec`.

## Why the Change?

* **Enhanced Precision**: `Dec` uses the [apd](https://github.com/cockroachdb/apd) library for arbitrary precision decimals, suitable for accurate financial calculations.
* **Immutable Operations**: `Dec` operations are safer for concurrent use as they do not mutate the original values.
* **Better Performance**: `Dec` operations are faster and more efficient than `LegacyDec`.

Benchmarking results below between `LegacyDec` and `Dec`:

```
BenchmarkCompareLegacyDecAndNewDec/LegacyDec-10    	 8621032	       143.8 ns/op	     144 B/op	       3 allocs/op
BenchmarkCompareLegacyDecAndNewDec/NewDec-10       	 5206173	       238.7 ns/op	     176 B/op	       7 allocs/op
BenchmarkCompareLegacyDecAndNewDecQuoInteger/LegacyDec-10         	 5767692	       205.1 ns/op	     232 B/op	       6 allocs/op
BenchmarkCompareLegacyDecAndNewDecQuoInteger/NewDec-10            	23172602	        51.75 ns/op	      16 B/op	       2 allocs/op
BenchmarkCompareLegacyAddAndDecAdd/LegacyDec-10                   	21157941	        56.33 ns/op	      80 B/op	       2 allocs/op
BenchmarkCompareLegacyAddAndDecAdd/NewDec-10                      	24133659	        48.92 ns/op	      48 B/op	       1 allocs/op
BenchmarkCompareLegacySubAndDecMul/LegacyDec-10                   	14256832	        87.47 ns/op	      80 B/op	       2 allocs/op
BenchmarkCompareLegacySubAndDecMul/NewDec-10                      	18273994	        65.68 ns/op	      48 B/op	       1 allocs/op
BenchmarkCompareLegacySubAndDecSub/LegacyDec-10                   	19988325	        64.46 ns/op	      80 B/op	       2 allocs/op
BenchmarkCompareLegacySubAndDecSub/NewDec-10                      	27430347	        42.45 ns/op	       8 B/op	       1 allocs/op
```

## Updating Your Modules

Modules using `LegacyDec` should transition to `Dec` to maintain compatibility with the latest SDK updates. This involves:

1. Updating type declarations from `LegacyDec` to `Dec`.
2. Modifying arithmetic operations to handle the new method signatures and potential errors.

# Example Update

Transitioning an addition operation from `LegacyDec` to `Dec`:

**Before:**

```go
result := legacyDec1.Add(legacyDec2)
```

**After:**

```go
result, err := dec1.Add(dec2)
if err != nil {
    log.Fatalf("Error during addition: %v", err)
}
```

This can be done for all arithmetic operations, including subtraction, multiplication, division, and more.
