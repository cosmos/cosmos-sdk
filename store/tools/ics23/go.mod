module github.com/cosmos/cosmos-sdk/store/tools/ics23

go 1.18

require (
	github.com/celestiaorg/smt v0.3.0
	github.com/confio/ics23/go v0.7.0
	github.com/cosmos/cosmos-sdk v0.46.1
	github.com/cosmos/iavl v0.19.1
	github.com/tendermint/tendermint v0.34.21
	github.com/tendermint/tm-db v0.6.7
)

require (
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cosmos/gorocksdb v1.2.0 // indirect
	github.com/dgraph-io/badger/v2 v2.2007.4 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/gogo/protobuf v1.3.3 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sasha-s/go-deadlock v0.2.1-0.20190427202633-1595213edefa // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e // indirect
	golang.org/x/net v0.0.0-20220726230323-06994584191e // indirect
	golang.org/x/sys v0.0.0-20220727055044-e65921a090b8 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)

replace github.com/cosmos/cosmos-sdk/store/tools/ics23 => ./

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
