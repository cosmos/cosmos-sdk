module cosmossdk.io/server/v2/stf

go 1.23

replace cosmossdk.io/core => ../../../core

require (
	cosmossdk.io/core v0.11.0
	github.com/cosmos/gogoproto v1.7.0
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/btree v1.7.0
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/cosmos/gogoproto => github.com/cosmos/gogoproto v1.6.1-0.20240809124342-d6a57064ada0
