module cosmossdk.io/core

go 1.18

require (
	cosmossdk.io/api v0.1.0-alpha8
	cosmossdk.io/depinject v1.0.0-alpha.4
	cosmossdk.io/math v1.0.0-beta.3
	github.com/cosmos/cosmos-proto v1.0.0-alpha7
	github.com/stretchr/testify v1.8.0
	google.golang.org/protobuf v1.28.0
	gotest.tools/v3 v3.3.0
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20220428152302-39d4317da171 // indirect
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/sys v0.0.0-20211019181941-9d821ace8654 // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/genproto v0.0.0-20211223182754-3ac035c7e7cb // indirect
	google.golang.org/grpc v1.48.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	cosmossdk.io/api => ../api
	cosmossdk.io/depinject => ../depinject
)
