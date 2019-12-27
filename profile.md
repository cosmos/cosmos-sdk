slashing profiling results
==========================

original bench
===

benchmark
---

```
goos: linux
goarch: amd64
pkg: github.com/cosmos/cosmos-sdk/x/slashing
BenchmarkBeginBlocker100-8           100          30795066 ns/op         7486165 B/op      90541 allocs/op
PASS
ok      github.com/cosmos/cosmos-sdk/x/slashing	5.915s
```

db size
---

```
14M	/tmp/bench-slashing111093085/test.db
14M	/tmp/bench-slashing111093085
```

missed array stored as a single batch
===

benchcmp to original
---
```
benchmark                      old ns/op     new ns/op     delta
BenchmarkBeginBlocker100-8     30795066      15535438      -49.55%

benchmark                      old allocs     new allocs     delta
BenchmarkBeginBlocker100-8     90541          49348          -45.50%

benchmark                      old bytes     new bytes     delta
BenchmarkBeginBlocker100-8     7486165       3915672       -47.69%
```

db size
---

```
2,3M	/tmp/bench-slashing347257969/test.db
2,3M	/tmp/bench-slashing347257969
```


reuse SignedBlocksWindow and MinSignedBlocksWindow
===

benchcmp to original
---

```
benchmark                      old ns/op     new ns/op     delta
BenchmarkBeginBlocker100-8     30795066      13009027      -57.76%

benchmark                      old allocs     new allocs     delta
BenchmarkBeginBlocker100-8     90541          43555          -51.89%

benchmark                      old bytes     new bytes     delta
BenchmarkBeginBlocker100-8     7486165       3654145       -51.19%
```