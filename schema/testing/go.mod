module cosmossdk.io/schema/testing

go 1.23

require (
	cosmossdk.io/schema v0.0.0
	github.com/cockroachdb/apd/v3 v3.2.1
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/btree v1.7.0
	gotest.tools/v3 v3.5.1
	pgregory.net/rapid v1.1.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace cosmossdk.io/schema => ./..
