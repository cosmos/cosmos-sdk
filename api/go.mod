module cosmossdk.io/api

go 1.19

require (
<<<<<<< HEAD
	github.com/cosmos/cosmos-proto v1.0.0-alpha8
	github.com/cosmos/gogoproto v1.4.3
	google.golang.org/genproto v0.0.0-20221014213838-99cd37c6964a
	google.golang.org/grpc v1.50.1
=======
	github.com/cosmos/cosmos-proto v1.0.0-beta.1
	github.com/cosmos/gogoproto v1.4.4
	google.golang.org/genproto v0.0.0-20230125152338-dcaf20b6aeaa
	google.golang.org/grpc v1.52.3
>>>>>>> d0a5bd1a0 (fix(reflection): Fix gogoproto import paths (#14838))
	google.golang.org/protobuf v1.28.1
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/net v0.0.0-20221017152216-f25eb7ecb193 // indirect
	golang.org/x/sys v0.0.0-20221013171732-95e765b1cc43 // indirect
	golang.org/x/text v0.3.8 // indirect
)
