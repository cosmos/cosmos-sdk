module cosmossdk.io/x/tx

go 1.23

require (
	cosmossdk.io/api v0.7.6
	cosmossdk.io/core v0.11.0
	cosmossdk.io/errors v1.0.1
	cosmossdk.io/math v1.5.0
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/cosmos/gogoproto v1.7.0
	github.com/google/go-cmp v0.7.0
	github.com/google/gofuzz v1.2.0
	github.com/iancoleman/strcase v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.10.0
	github.com/tendermint/go-amino v0.16.0
	google.golang.org/protobuf v1.36.5
	gotest.tools/v3 v3.5.2
	pgregory.net/rapid v1.1.0
)

require (
	github.com/cockroachdb/apd/v3 v3.2.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/exp v0.0.0-20240404231335-c0f41cb1a7a0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250212204824-5a70512c5d8b // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250212204824-5a70512c5d8b // indirect
	google.golang.org/grpc v1.71.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace cosmossdk.io/api => ../../api

// NOTE: we do not want to replace to the development version of cosmossdk.io/api yet
// Until https://github.com/cosmos/cosmos-sdk/issues/19228 is resolved
// We are tagging x/tx v0.14+ from main and v0.13 from release/v0.50.x and must keep using released versions of x/tx dependencies
