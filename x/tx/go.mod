module cosmossdk.io/x/tx

go 1.23

require (
	cosmossdk.io/api v0.8.2
	cosmossdk.io/errors v1.0.1
	cosmossdk.io/math v1.5.0
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/cosmos/gogoproto v1.7.0
	github.com/google/go-cmp v0.6.0
	github.com/google/gofuzz v1.2.0
	github.com/iancoleman/strcase v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.10.0
	github.com/tendermint/go-amino v0.16.0
	google.golang.org/protobuf v1.36.2
	gotest.tools/v3 v3.5.1
	pgregory.net/rapid v1.1.0
)

require (
	buf.build/gen/go/cometbft/cometbft/protocolbuffers/go v1.36.2-20241120201313-68e42a58b301.1 // indirect
	buf.build/gen/go/cosmos/gogo-proto/protocolbuffers/go v1.36.2-20240130113600-88ef6483f90f.1 // indirect
	github.com/cockroachdb/apd/v3 v3.2.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241015192408-796eee8c2d53 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250106144421-5f5ef82da422 // indirect
	google.golang.org/grpc v1.69.4 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// NOTE: we do not want to replace to the development version of cosmossdk.io/api yet
// Until https://github.com/cosmos/cosmos-sdk/issues/19228 is resolved
// We are tagging x/tx from main and must keep using released versions of x/tx dependencies
