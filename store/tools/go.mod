module github.com/cosmos/cosmos-sdk/store/tools

go 1.17

require (
	github.com/confio/ics23/go v0.7.0
	github.com/cosmos/cosmos-sdk/store/tools v0.0.0-00010101000000-000000000000
	github.com/lazyledger/smt v0.2.1-0.20210709230900-03ea40719554
	github.com/tendermint/iavl v0.13.2
	github.com/tendermint/tendermint v0.33.2
	github.com/tendermint/tm-db v0.5.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/etcd-io/bbolt v1.3.3 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.3.4 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20190923125748-758128399b1d // indirect
	github.com/tecbot/gorocksdb v0.0.0-20191217155057-f0fad39f321c // indirect
	github.com/tendermint/go-amino v0.14.1 // indirect
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413 // indirect
	golang.org/x/sys v0.0.0-20200122134326-e047566fdf82 // indirect
)

replace github.com/cosmos/cosmos-sdk/store/tools => ./

