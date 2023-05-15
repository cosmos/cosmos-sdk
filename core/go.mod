module cosmossdk.io/core

go 1.19

require (
	cosmossdk.io/api v0.3.1
	cosmossdk.io/depinject v1.0.0-alpha.3
	cosmossdk.io/math v1.0.1
	github.com/cosmos/cosmos-proto v1.0.0-beta.2
	github.com/stretchr/testify v1.8.1
	google.golang.org/grpc v1.55.0
	google.golang.org/protobuf v1.30.0
	gotest.tools/v3 v3.4.0
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/cosmos/gogoproto v1.4.10 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	golang.org/x/exp v0.0.0-20230203172020-98cc5a0785f9 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// temporary until we tag a new go module
replace cosmossdk.io/math => ../math
