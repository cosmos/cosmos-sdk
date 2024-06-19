module cosmossdk.io/indexer/testing

require (
	cosmossdk.io/indexer/base v0.0.0
	github.com/brianvoe/gofakeit/v7 v7.0.3
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/btree v1.7.0
	pgregory.net/rapid v1.1.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace cosmossdk.io/indexer/base => ../base

go 1.22
