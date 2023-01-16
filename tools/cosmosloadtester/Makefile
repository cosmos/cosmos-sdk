GO ?= "go"
LDFLAGS ?='-s -w'
GOOS ?= "linux"
GOARCH ?= "amd64"
CGO_ENABLED ?= 0

.PHONY: pb server ui

pb:
	(cd proto && buf mod update)
	(cd proto && buf generate --template buf.gen.yaml)
	(cd proto && buf generate --template buf.gen.ts.grpcweb.yaml --include-imports)
	npx --yes swagger-typescript-api -p ./proto/orijtech/cosmosloadtester/v1/loadtest_service.swagger.json -o ./ui/src/gen -n LoadtestApi.ts
server:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -trimpath -ldflags $(LDFLAGS) -o bin/server ./cmd/server
ui:
	(cd ui && npm install)
	(cd ui && npm run build)
