package main

import (
	"google.golang.org/protobuf/compiler/protogen"

	"cosmossdk.io/orm/internal/codegen"
)

func main() {
	protogen.Options{}.Run(codegen.QueryProtoPluginRunner)
}
