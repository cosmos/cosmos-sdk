package main

import (
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/pointnetwork/cosmos-point-sdk/orm/internal/codegen"
)

func main() {
	protogen.Options{}.Run(codegen.PluginRunner)
}
