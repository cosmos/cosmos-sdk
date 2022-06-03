module cosmossdk.io/core

go 1.18

require (
	cosmossdk.io/api v0.1.0-alpha8
	github.com/cosmos/cosmos-proto v1.0.0-alpha7
	github.com/cosmos/cosmos-sdk/depinject v1.0.0-alpha.4
	google.golang.org/protobuf v1.28.0
	gotest.tools/v3 v3.2.0
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/exp v0.0.0-20220428152302-39d4317da171 // indirect
	golang.org/x/net v0.0.0-20220520000938-2e3eb7b945c2 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220519153652-3a47de7e79bd // indirect
	google.golang.org/grpc v1.47.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	github.com/cosmos/cosmos-sdk/api => ../api
	github.com/cosmos/cosmos-sdk/depinject => ../depinject
)
