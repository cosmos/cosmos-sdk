module cosmossdk.io/x/upgrade

go 1.19

require (
	cosmossdk.io/api v0.2.6
	cosmossdk.io/core v0.5.1
	cosmossdk.io/depinject v1.0.0-alpha.3
	cosmossdk.io/store v0.0.0-20230201234215-6b256ce7c087
	github.com/cosmos/cosmos-db v0.0.0-20230119180254-161cf3632b7c
	github.com/cosmos/cosmos-proto v1.0.0-beta.1
	github.com/cosmos/cosmos-sdk v0.0.0-20230201234215-6b256ce7c087
	github.com/cosmos/gogoproto v1.4.4
	github.com/golang/protobuf v1.5.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/go-getter v1.6.2
	github.com/spf13/cast v1.5.0
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.1
	github.com/tendermint/tendermint v0.37.0-rc2
	google.golang.org/genproto v0.0.0-20230125152338-dcaf20b6aeaa
	google.golang.org/grpc v1.52.3
	google.golang.org/protobuf v1.28.1
)

replace (
	// Fix upstream GHSA-h395-qcrw-5vmq vulnerability.
	// TODO Remove it: https://github.com/cosmos/cosmos-sdk/issues/10409
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.8.1
)
