module cosmossdk.io/tx

go 1.19

require (
	cosmossdk.io/api v0.2.4
	cosmossdk.io/core v0.3.2
	cosmossdk.io/math v1.0.0-beta.3
	github.com/cosmos/cosmos-proto v1.0.0-alpha8
	github.com/stretchr/testify v1.8.1
	google.golang.org/protobuf v1.28.1
)

require (
	github.com/cosmos/gogoproto v1.4.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.1.0 // indirect
	golang.org/x/sys v0.1.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	google.golang.org/genproto v0.0.0-20221027153422-115e99e71e1c // indirect
	google.golang.org/grpc v1.50.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// temporary until we tag a new go module
replace (
	cosmossdk.io/core => ../core
	cosmossdk.io/math => ../math
)
