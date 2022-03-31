package codegen

import (
	"fmt"

	"github.com/cosmos/cosmos-proto/generator"
	orm "github.com/cosmos/cosmos-sdk/api/cosmos/orm/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

const (
	contextPkg  = protogen.GoImportPath("context")
	ormListPkg  = protogen.GoImportPath("github.com/cosmos/cosmos-sdk/orm/model/ormlist")
	ormdbPkg    = protogen.GoImportPath("github.com/cosmos/cosmos-sdk/orm/model/ormdb")
	ormErrPkg   = protogen.GoImportPath("github.com/cosmos/cosmos-sdk/orm/types/ormerrors")
	ormTablePkg = protogen.GoImportPath("github.com/cosmos/cosmos-sdk/orm/model/ormtable")
)

func PluginRunner(p *protogen.Plugin) error {
	for _, f := range p.Files {
		if !f.Generate {
			continue
		}

		if !hasTables(f) {
			continue
		}

		gen := p.NewGeneratedFile(fmt.Sprintf("%s.cosmos_orm.go", f.GeneratedFilenamePrefix), f.GoImportPath)
		cgen := &generator.GeneratedFile{
			GeneratedFile: gen,
			LocalPackages: map[string]bool{},
		}
		f := fileGen{GeneratedFile: cgen, file: f}
		err := f.gen()
		if err != nil {
			return err
		}

	}

	return nil
}

func hasTables(file *protogen.File) bool {
	for _, message := range file.Messages {
		if proto.GetExtension(message.Desc.Options(), orm.E_Table).(*orm.TableDescriptor) != nil {
			return true
		}

		if proto.GetExtension(message.Desc.Options(), orm.E_Singleton).(*orm.SingletonDescriptor) != nil {
			return true
		}
	}

	return false
}
