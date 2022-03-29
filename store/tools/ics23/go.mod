module github.com/cosmos/cosmos-sdk/store/tools/ics23

go 1.17

require (
	github.com/confio/ics23/go v0.7.0
	github.com/cosmos/cosmos-sdk v0.45.0
	github.com/cosmos/iavl v0.18.0
	github.com/lazyledger/smt v0.2.1-0.20210709230900-03ea40719554
	github.com/tendermint/tendermint v0.35.2
	github.com/tendermint/tm-db v0.6.6
)

require (
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cosmos/gogoproto v1.4.0 // indirect
	github.com/dgraph-io/badger/v2 v2.2007.4 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/tecbot/gorocksdb v0.0.0-20191217155057-f0fad39f321c // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220315194320-039c03cc5b86 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220314164441-57ef72a4c106 // indirect
	google.golang.org/grpc v1.45.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)

replace github.com/cosmos/cosmos-sdk/store/tools/ics23 => ./

// TODO: remove once cosmos-sdk 0.46 beta will be released
replace github.com/cosmos/cosmos-sdk => ../../../
