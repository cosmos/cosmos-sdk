module cosmossdk.io/x/tx

go 1.23.0

require (
	cosmossdk.io/api v1.0.0-alpha.0
	cosmossdk.io/core v1.1.0-alpha.0
	cosmossdk.io/errors v1.0.2
	cosmossdk.io/math v1.5.3
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/cosmos/gogoproto v1.7.0
	github.com/google/go-cmp v0.7.0
	github.com/google/gofuzz v1.2.0
	github.com/iancoleman/strcase v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.10.0
	github.com/tendermint/go-amino v0.16.0
	google.golang.org/protobuf v1.36.6
	gotest.tools/v3 v3.5.2
	pgregory.net/rapid v1.2.0
)

require (
	buf.build/gen/go/cometbft/cometbft/protocolbuffers/go v1.36.6-20241120201313-68e42a58b301.1 // indirect
	buf.build/gen/go/cosmos/gogo-proto/protocolbuffers/go v1.36.6-20240130113600-88ef6483f90f.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250422160041-2d3770c4ea7f // indirect
	google.golang.org/grpc v1.72.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// retracting released version from unreleased sdk v0.52
retract [v1.0.0, v1.1.0]
