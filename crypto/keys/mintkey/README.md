To run Bcrypt benchmarks:

```bash
go test -bench .
```

On the test machine (midrange ThinkPad; i7 6600U), this results in:

```bash
goos: linux
goarch: amd64
pkg: github.com/cosmos/cosmos-sdk/crypto/keys/mintkey
BenchmarkBcryptGenerateFromPassword/benchmark-security-param-9-4         	      50	  34609268 ns/op
BenchmarkBcryptGenerateFromPassword/benchmark-security-param-10-4        	      20	  67874471 ns/op
BenchmarkBcryptGenerateFromPassword/benchmark-security-param-11-4        	      10	 135515404 ns/op
BenchmarkBcryptGenerateFromPassword/benchmark-security-param-12-4        	       5	 274824600 ns/op
BenchmarkBcryptGenerateFromPassword/benchmark-security-param-13-4        	       2	 547012903 ns/op
BenchmarkBcryptGenerateFromPassword/benchmark-security-param-14-4        	       1	1083685904 ns/op
BenchmarkBcryptGenerateFromPassword/benchmark-security-param-15-4        	       1	2183674041 ns/op
PASS
ok  	github.com/cosmos/cosmos-sdk/crypto/keys/mintkey	12.093s
```

Benchmark results are in nanoseconds, so security parameter 12 takes about a quarter of a second to generate the Bcrypt key, security param 13 takes half a second, and so on.
