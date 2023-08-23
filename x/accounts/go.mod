module cosmossdk.io/x/accounts

go 1.21

require (
	github.com/cosmos/gogoproto v1.4.11
	github.com/stretchr/testify v1.8.4
	google.golang.org/grpc v1.57.0
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20230811145659-89c5cff77bcb // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230815205213-6bfd019c3878 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// REMOVE post v1 release.
replace cosmossdk.io/api => ../../api
