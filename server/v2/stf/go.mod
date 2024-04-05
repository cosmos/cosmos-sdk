module cosmossdk.io/server/v2/stf

go 1.21

replace (
	cosmossdk.io/collections => ../../../collections
	cosmossdk.io/core => ../../../core
)

require (
	cosmossdk.io/collections v0.4.0
	cosmossdk.io/core v0.12.0
	github.com/cosmos/gogoproto v1.4.12
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/btree v1.7.0
	golang.org/x/exp v0.0.0-20240314144324-c7f7c6466f7f
	google.golang.org/protobuf v1.33.0
)

require (
	cosmossdk.io/log v1.3.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rs/zerolog v1.32.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240221002015-b0ce06bbee7c // indirect
	google.golang.org/grpc v1.62.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
