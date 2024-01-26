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
	github.com/tidwall/btree v1.7.0
	google.golang.org/protobuf v1.32.0
)

require (
	github.com/cosmos/gogoproto v1.4.11 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/kr/text v0.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231009173412-8bfb1ae86b6c // indirect
	google.golang.org/grpc v1.60.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
