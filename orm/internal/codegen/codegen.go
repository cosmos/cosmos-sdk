package codegen

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	v1alpha1 "github.com/cosmos/cosmos-sdk/api/cosmos/orm/v1alpha1"

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
		if proto.GetExtension(message.Desc.Options(), v1alpha1.E_Table).(*v1alpha1.TableDescriptor) != nil {
			return true
		}

		if proto.GetExtension(message.Desc.Options(), v1alpha1.E_Singleton).(*v1alpha1.SingletonDescriptor) != nil {
			return true
		}
	}

	return false
}
