module cosmossdk.io/core

go 1.23

require (
	cosmossdk.io/api v0.3.1
	cosmossdk.io/depinject v1.0.0-alpha.4
	cosmossdk.io/math v1.3.0
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/stretchr/testify v1.9.0
	google.golang.org/grpc v1.68.0
	google.golang.org/protobuf v1.34.2
	gotest.tools/v3 v3.5.1
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/cosmos/gogoproto v1.4.10 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20230711153332-06a737ee72cb // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	cosmossdk.io/api => ../api
	cosmossdk.io/depinject => ../depinject
	cosmossdk.io/math => ../math
)
