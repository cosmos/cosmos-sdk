go 1.17

module github.com/cosmos/cosmos-sdk/db

require (
	github.com/dgraph-io/badger/v3 v3.2103.1
	github.com/google/btree v1.0.0
	github.com/stretchr/testify v1.7.0
	github.com/tecbot/gorocksdb v0.0.0-20191217155057-f0fad39f321c
)

require (
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/flatbuffers v1.12.0 // indirect
	github.com/klauspost/compress v1.12.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opencensus.io v0.22.5 // indirect
	golang.org/x/net v0.0.0-20201021035429-f5854403a974 // indirect
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

// FIXME: gorocksdb bindings for OptimisticTransactionDB are not merged upstream, so we use a fork
// See https://github.com/tecbot/gorocksdb/pull/216
replace github.com/tecbot/gorocksdb => github.com/roysc/gorocksdb v1.1.0
