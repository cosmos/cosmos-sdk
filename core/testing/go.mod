module cosmossdk.io/core/testing

go 1.23

require (
	cosmossdk.io/core v1.0.0-alpha.6
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/btree v1.7.0
	go.uber.org/mock v0.5.0
)

require (
	cosmossdk.io/schema v0.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace cosmossdk.io/core => ..
