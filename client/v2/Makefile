codegen:
	@(cd internal; buf generate --template testpbpulsar/buf.gen.yaml)
	@(cd internal; buf generate --template testpbgogo/buf.gen.yaml)
	@(cd internal/testpbgogo; rm *_grpc.pb.go; rm *.pulsar.go)
	@(cd internal; mv github.com/cosmos/cosmos-sdk/client/v2/internal/testpbgogo/* testpbgogo)
	@(cd internal; rm -r github.com)