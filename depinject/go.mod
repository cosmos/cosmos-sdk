module github.com/cosmos/cosmos-sdk/depinject/v2

go 1.23.0

require (
	cosmossdk.io/depinject v1.2.0
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/cosmos/gogoproto v1.7.0
	github.com/stretchr/testify v1.10.0
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6
	gotest.tools/v3 v3.5.2
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/tendermint/go-amino v0.16.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250422160041-2d3770c4ea7f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// keep grpc 1.67.1 to avoid go minimum version bump (depinject should be compatible with 0.47, 0.50 and 0.53)
replace google.golang.org/grpc => google.golang.org/grpc v1.67.1
