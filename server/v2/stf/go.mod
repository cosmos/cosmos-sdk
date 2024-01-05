module cosmossdk.io/server/v2/stf

go 1.21

replace (
	cosmossdk.io/core => ../../../core
	cosmossdk.io/server/v2/core => ../core
)

require (
	cosmossdk.io/core v0.11.0
	cosmossdk.io/server/v2/core v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.8.4
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/cosmos/gogoproto v1.4.11 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
