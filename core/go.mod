module cosmossdk.io/core

go 1.18

require (
	cosmossdk.io/api v0.1.0-alpha8
	github.com/cosmos/cosmos-proto v1.0.0-alpha7
	github.com/cosmos/cosmos-sdk/depinject v1.0.0-alpha.4
	github.com/gogo/protobuf v1.3.2
	github.com/tendermint/tm-db v0.6.7
	google.golang.org/grpc v1.46.2
	google.golang.org/protobuf v1.28.0
	gotest.tools/v3 v3.2.0
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/DataDog/zstd v1.4.1 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cosmos/gorocksdb v1.2.0 // indirect
	github.com/dgraph-io/badger/v2 v2.2007.2 // indirect
	github.com/dgraph-io/ristretto v0.0.3-0.20200630154024-f66de99634de // indirect
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20200815110645-5c35d600f0ca // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	golang.org/x/exp v0.0.0-20220428152302-39d4317da171 // indirect
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/sys v0.0.0-20211019181941-9d821ace8654 // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/genproto v0.0.0-20211223182754-3ac035c7e7cb // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	github.com/cosmos/cosmos-sdk/api => ../api
	github.com/cosmos/cosmos-sdk/depinject => ../depinject
)
