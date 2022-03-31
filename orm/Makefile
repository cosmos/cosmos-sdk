codegen:
	go install ./cmd/protoc-gen-go-cosmos-orm
	(cd internal; buf generate)

build:
	go build ./...

test:
	go test ./...
