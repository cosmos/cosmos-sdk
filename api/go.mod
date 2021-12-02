module github.com/cosmos/cosmos-sdk/api

go 1.17

require (
	github.com/cosmos/cosmos-proto v1.0.0-alpha3
	github.com/gogo/protobuf v1.3.3
	google.golang.org/protobuf v1.27.1
)

require google.golang.org/genproto v0.0.0-20211129164237-f09f9a12af12 // indirect

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
