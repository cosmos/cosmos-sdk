package main

import (
	"github.com/cosmos/cosmos-sdk/orm/internal/codegen"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	protogen.Options{}.Run(codegen.QueryProtoPluginRunner)
}
