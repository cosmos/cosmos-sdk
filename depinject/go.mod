module cosmossdk.io/depinject

go 1.23

require (
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/cosmos/gogoproto v1.7.0
	github.com/stretchr/testify v1.10.0
	google.golang.org/grpc v1.70.0
	google.golang.org/protobuf v1.36.4
	gotest.tools/v3 v3.5.2
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/tendermint/go-amino v0.16.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250122153221-138b5a5a4fd4 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// keep grpc 1.67.1 to avoid go minimum version bump (depinject should be compatible with 0.47, 0.50 and 0.52)
replace google.golang.org/grpc => google.golang.org/grpc v1.67.1
