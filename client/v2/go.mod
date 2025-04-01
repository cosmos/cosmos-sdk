module cosmossdk.io/client/v2

go 1.22.7

toolchain go1.23.4

require (
	cosmossdk.io/api v0.3.1
	cosmossdk.io/core v0.3.2
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/iancoleman/strcase v0.2.0
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	google.golang.org/grpc v1.68.0
	google.golang.org/protobuf v1.34.2
	gotest.tools/v3 v3.5.1
)

require (
	cosmossdk.io/depinject v1.0.0-alpha.4 // indirect
	github.com/cosmos/gogoproto v1.4.10 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/exp v0.0.0-20230711153332-06a737ee72cb // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4 // indirect
)

replace (
	cosmossdk.io/api => ../../api
	cosmossdk.io/core => ../../core
	cosmossdk.io/depinject => ../../depinject
)
