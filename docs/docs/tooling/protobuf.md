---
sidebar_position: 1
---


# Protocol Buffers

It is known that Cosmos SDK uses protocol buffers extensively, this docuemnt is meant to provide a guide on how it is used in the cosmos-sdk. 

To generate the proto file, the Cosmos-SDK uses a docker image, this image is provided to all to use as well. The latest version is `ghcr.io/cosmos/proto-builder:0.11.0`

Below is the example of the Cosmos-SDK's commands for generating, linting, and formatting protobuf files that can be reused in any applications makefile. 
```
protoVer=0.11.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
containerProtoGen=$(PROJECT_NAME)-proto-gen-$(protoVer)
containerProtoGenSwagger=$(PROJECT_NAME)-proto-gen-swagger-$(protoVer)
containerProtoFmt=$(PROJECT_NAME)-proto-fmt-$(protoVer)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf:1.7.0

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGen}$$"; then docker start -a $(containerProtoGen); else docker run --name $(containerProtoGen) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./scripts/protocgen.sh; fi

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGenSwagger}$$"; then docker start -a $(containerProtoGenSwagger); else docker run --name $(containerProtoGenSwagger) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./scripts/protoc-swagger-gen.sh; fi

proto-format:
	@echo "Formatting Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoFmt}$$"; then docker start -a $(containerProtoFmt); else docker run --name $(containerProtoFmt) -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./ -name "*.proto" -exec clang-format -i {} \; ; fi


proto-lint:
	@$(DOCKER_BUF) lint --error-format=json

proto-check-breaking:
	@$(DOCKER_BUF) breaking --against $(HTTPS_GIT)#branch=main
```

The script used to generate the protobuf files can be found in the `scripts/` directory. 

```sh reference
https://github.com/cosmos/cosmos-sdk/blob/10e8aadcad3a30dda1d6163c39c9f86b4a877e54/scripts/protocgen.sh#L1-L37
```

## Buf

[Buf](https://buf.build) is a protobuf tool that abstracts the needs to use the complicated `protoc` toolchain on top of various other things that ensure you are using protobuf in accordance with the majority of the ecosystem. Within the cosmos-sdk repository there are a few files that have a buf prefix. Lets start with the top level and then dive into the various directories. 

### Workspace

At the root level directory a workspace is defined using [buf workspaces](https://docs.buf.build/configuration/v1/buf-work-yaml). This helps if there are one or more protobuf containing directories in your project. 

Cosmos-SDK example: 
```go reference
https://github.com/notional-labs/cosmos-sdk/blob/78c463c2130b18d823f7713f336a9b76e7b6d8b8/buf.work.yaml#L6-L9
```

### Proto Directory

Next is the `proto/` directory where all of our protobuf files live. In here there are many different buf files defined each serving a different purpose. 

```bash
├── README.md
├── buf.gen.gogo.yaml
├── buf.gen.pulsar.yaml
├── buf.gen.swagger.yaml
├── buf.lock
├── buf.md
├── buf.yaml
├── cosmos
└── tendermint
```

The above diagram all the files and directories within the Cosmos-SDK `proto/` directory. 

#### `buf.gen.gogo.yaml`

`buf.gen.gogo.yaml` defines how the protobuf files should be generated for use with in the module. This file uses [gogoproto](https://github.com/gogo/protobuf), a separate generator from the google go-proto generator that makes working with various objects more ergonomic, and it has more performant encode and decode steps

```go reference
https://github.com/cosmos/cosmos-sdk/blob/78c463c2130b18d823f7713f336a9b76e7b6d8b8/proto/buf.gen.gogo.yaml#L1-l9
```

> Example of how to define `gen` files can be found [here](https://docs.buf.build/tour/generate-go-code)

#### `buf.gen.pulsar.yaml`

`buf.gen.pulsar.yaml` defines how protobuf files should be generated using the [new golang apiv2 of protobuf](https://go.dev/blog/protobuf-apiv2). This generator is used instead of the google go-proto generator because it has some extra helpers for Cosmos-SDK applications and will have more performant encode and decode than the google go-proto generator. You can follow the development of this generator [here](https://github.com/cosmos/cosmos-proto). 

```go reference
https://github.com/cosmos/cosmos-sdk/blob/78c463c2130b18d823f7713f336a9b76e7b6d8b8/proto/buf.gen.pulsar.yaml#L1-L18
```

> Example of how to define `gen` files can be found [here](https://docs.buf.build/tour/generate-go-code)

#### `buf.gen.swagger.yaml`

`buf.gen.swagger.yaml` generates the swagger documentation for the query and messages of the chain. This will only define the REST API end points that were defined in the query and msg servers. You can find examples of this [here](https://github.com/cosmos/cosmos-sdk/blob/78c463c2130b18d823f7713f336a9b76e7b6d8b8/proto/cosmos/bank/v1beta1/query.proto#L19)

```go reference
https://github.com/cosmos/cosmos-sdk/blob/78c463c2130b18d823f7713f336a9b76e7b6d8b8/proto/buf.gen.swagger.yaml#L1-L6
```

> Example of how to define `gen` files can be found [here](https://docs.buf.build/tour/generate-go-code)

#### `buf.lock`

This is a autogenerated file based off the dependencies required by the `.gen` files. There is no need to copy the current one. If you depend on cosmos-sdk proto definitions a new entry for the Cosmos-SDK will need to be provided. The dependency you will need to use is `buf.build/cosmos/cosmos-sdk`.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/78c463c2130b18d823f7713f336a9b76e7b6d8b8/proto/buf.lock#L1-L16
```

#### `buf.yaml`

`buf.yaml` defines the [name of your package](https://github.com/cosmos/cosmos-sdk/blob/78c463c2130b18d823f7713f336a9b76e7b6d8b8/proto/buf.yaml#L3), which [breakage checker](https://docs.buf.build/tour/detect-breaking-changes) to use and how to [lint your protobuf files](https://docs.buf.build/tour/lint-your-api). 

```go reference
https://github.com/cosmos/cosmos-sdk/blob/78c463c2130b18d823f7713f336a9b76e7b6d8b8/proto/buf.yaml#L1-L24
```

We use a variety of linters for the Cosmos-SDK protobuf files. The repo also checks this in ci. 

A reference to the github actions can be found [here](https://github.com/cosmos/cosmos-sdk/blob/78c463c2130b18d823f7713f336a9b76e7b6d8b8/.github/workflows/proto.yml#L1-L32)

```go reference
https://github.com/cosmos/cosmos-sdk/blob/78c463c2130b18d823f7713f336a9b76e7b6d8b8/.github/workflows/proto.yml#L1-L32
```
