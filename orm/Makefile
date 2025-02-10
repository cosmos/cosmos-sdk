codegen:
	go install ./cmd/protoc-gen-go-cosmos-orm
	go install ./cmd/protoc-gen-go-cosmos-orm-proto
	# generate .proto files first
	(cd internal; buf generate --template buf.proto.gen.yaml)
	# generate go code
	(cd internal; buf generate)