codegen:
	go install ./cmd/protoc-gen-go-cosmos-orm
	(cd internal; buf generate; buf generate --template buf.gen.orm.yaml --path bankexample)