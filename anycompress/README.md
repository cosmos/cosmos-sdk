## anycompress

A transparent layer over [tendermint/tm-db.DB]() that wraps a database and intercepts .Set and .Get and compresses and decompresses google.protobuf.Any to reduce on the amount of redundant stored, regardless of the amount of compression by databases. The benefits of using anycompress come from the fac that database compression rarely happens in blocks after rows are stored, because post storage compressions reduce random accesses and don't fit well with the underlying data structures.

### Examining benchmarks and savings
```shell
$ go run -tags rocksdb,boltdb,leveldb cmd/anycompress/main.go -n 1000 -type rocksdb

time to generate  800000 key-value pairs: 4.009872537s
Wrote MemoryProfile to disk mem-1598472491
Wrote CPUProfile to disk cpu-1598472466
Wrote MemoryProfile to disk mem-1598472536
Wrote CPUProfile to disk cpu-1598472516
Wrote MemoryProfile to disk mem-1598472581
rocksdb savings: 53.068%
Original:      736.308MiB (772075015B) - 2m15.992923978s
AnyCompressed: 345.565MiB (362351474B) - 2m26.682974595s
TimeSpent: 2m26.683016742s
```
