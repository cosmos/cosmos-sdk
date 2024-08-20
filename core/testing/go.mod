module cosmossdk.io/core/testing

go 1.20

replace cosmossdk.io/core => ../

require (
	cosmossdk.io/core v0.12.0
	github.com/spf13/cast v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d
	github.com/tidwall/btree v1.7.0
)

require github.com/golang/snappy v0.0.4 // indirect
