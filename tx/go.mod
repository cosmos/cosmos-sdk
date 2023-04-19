module cosmossdk.io/tx

go 1.19

require (
	cosmossdk.io/api v0.3.1
	cosmossdk.io/core v0.3.2
	cosmossdk.io/math v1.0.0
	github.com/cosmos/cosmos-proto v1.0.0-beta.2
	github.com/stretchr/testify v1.8.2
	google.golang.org/protobuf v1.30.0
)

require (
	github.com/cosmos/gogoproto v1.4.8 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20230321023759-10a507213a29 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230202175211-008b39050e57 // indirect
	google.golang.org/grpc v1.54.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// temporary until we tag a new go module
replace cosmossdk.io/core => ../core
