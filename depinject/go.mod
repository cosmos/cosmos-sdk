module cosmossdk.io/depinject

go 1.23.0

toolchain go1.23.7

replace (
	cosmossdk.io/api => ../api
	github.com/cosmos/cosmos-sdk => ../.
)

require (
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/cosmos/cosmos-sdk v0.50.13
	github.com/cosmos/gogoproto v1.7.0
	github.com/stretchr/testify v1.10.0
	google.golang.org/grpc v1.71.0
	google.golang.org/protobuf v1.36.6
	gotest.tools/v3 v3.5.2
	sigs.k8s.io/yaml v1.4.0
)

require (
	cosmossdk.io/api v0.7.6 // indirect
	cosmossdk.io/core v0.11.0 // indirect
	cosmossdk.io/errors v1.0.2 // indirect
	cosmossdk.io/x/tx v0.13.7 // indirect
	github.com/cometbft/cometbft v0.38.17 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/oasisprotocol/curve25519-voi v0.0.0-20230904125328-1f23a7beb09a // indirect
	github.com/petermattis/goid v0.0.0-20240813172612-4fcff4a6cae7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/sasha-s/go-deadlock v0.3.5 // indirect
	github.com/tendermint/go-amino v0.16.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// keep grpc 1.67.1 to avoid go minimum version bump (depinject should be compatible with 0.47, 0.50 and 0.52)
replace google.golang.org/grpc => google.golang.org/grpc v1.67.1
