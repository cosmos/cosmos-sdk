go 1.17

module github.com/cosmos/cosmos-sdk/db

require (
	github.com/dgraph-io/badger/v3 v3.2103.1
	github.com/google/btree v1.0.0
	github.com/stretchr/testify v1.7.0
	github.com/tecbot/gorocksdb v0.0.0-20191217155057-f0fad39f321c
)

require (
	github.com/davecgh/go-spew v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)
replace github.com/tecbot/gorocksdb => github.com/roysc/gorocksdb v0.0.0-20210804143633-c0bf0b3635e5 // FIXME

