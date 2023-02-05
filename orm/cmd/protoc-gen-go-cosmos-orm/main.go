package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/cosmos/cosmos-sdk/orm/internal/codegen"
)

func main() {
	flag.BoolVar(&codegen.GenQueries, "query-gen", false, "generate queries")
	protogen.Options{ParamFunc: flag.Set}.Run(codegen.ORMPluginRunner)
}
