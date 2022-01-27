package codegen

import (
	"fmt"

	"github.com/cosmos/cosmos-proto/generator"

	"google.golang.org/protobuf/compiler/protogen"
)

const (
	contextPkg          = protogen.GoImportPath("context")
	protoreflectPackage = protogen.GoImportPath("google.golang.org/protobuf/reflect/protoreflect")
	ormListPkg          = protogen.GoImportPath("github.com/cosmos/cosmos-sdk/orm/model/ormlist")
	ormdbPkg            = protogen.GoImportPath("github.com/cosmos/cosmos-sdk/orm/model/ormdb")
	ormErrPkg           = protogen.GoImportPath("github.com/cosmos/cosmos-sdk/orm/types/ormerrors")
	fmtPkg              = protogen.GoImportPath("fmt")
)

func PluginRunner(p *protogen.Plugin) error {
	for _, f := range p.Files {
		if !f.Generate {
			continue
		}

		gen := p.NewGeneratedFile(fmt.Sprintf("%s.cosmos_orm.go", f.GeneratedFilenamePrefix), f.GoImportPath)
		cgen := &generator.GeneratedFile{
			GeneratedFile: gen,
			Ext:           nil,
			LocalPackages: map[string]bool{},
		}
		f := fileGen{GeneratedFile: cgen, file: f}
		// TODO skip files without tables
		err := f.gen()
		if err != nil {
			return err
		}

	}

	return nil
}
