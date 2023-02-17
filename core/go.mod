module cosmossdk.io/core

go 1.19

require (
	cosmossdk.io/api v0.2.6
	cosmossdk.io/depinject v1.0.0-alpha.3
	cosmossdk.io/math v1.0.0-beta.6
	github.com/cosmos/cosmos-proto v1.0.0-beta.1
	github.com/stretchr/testify v1.8.1
	google.golang.org/grpc v1.51.0
	google.golang.org/protobuf v1.28.1
	gotest.tools/v3 v3.4.0
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/cosmos/gogoproto v1.4.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
<<<<<<< HEAD
	golang.org/x/exp v0.0.0-20221019170559-20944726eadf // indirect
	golang.org/x/net v0.2.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	google.golang.org/genproto v0.0.0-20221118155620-16455021b5e6 // indirect
=======
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20200815110645-5c35d600f0ca // indirect
	golang.org/x/exp v0.0.0-20230203172020-98cc5a0785f9 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20230202175211-008b39050e57 // indirect
>>>>>>> 77d347b48 (build(deps): Bump github.com/hashicorp/go-getter from 1.6.2 to 1.7.0 and go version to 1.20.1 (#15051))
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// temporary until we tag a new go module
replace cosmossdk.io/math => ../math
