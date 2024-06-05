module cosmossdk.io/tx

go 1.19

require (
	cosmossdk.io/api v0.3.1
	cosmossdk.io/core v0.3.2
	cosmossdk.io/math v1.3.0
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/stretchr/testify v1.9.0
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/cosmos/gogoproto v1.5.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20240123012728-ef4313101c80 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240123012728-ef4313101c80 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240123012728-ef4313101c80 // indirect
	google.golang.org/grpc v1.62.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// temporary until we tag a new go module
replace cosmossdk.io/core => ../core
