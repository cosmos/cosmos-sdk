package codegen

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-proto/generator"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	ormv1 "cosmossdk.io/api/cosmos/orm/v1"
)

const (
	contextPkg  = protogen.GoImportPath("context")
	ormListPkg  = protogen.GoImportPath("cosmossdk.io/orm/model/ormlist")
	ormErrPkg   = protogen.GoImportPath("cosmossdk.io/orm/types/ormerrors")
	ormTablePkg = protogen.GoImportPath("cosmossdk.io/orm/model/ormtable")
)

func ORMPluginRunner(p *protogen.Plugin) error {
	p.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
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
		fgen := fileGen{GeneratedFile: cgen, file: f}
		err := fgen.gen()
		if err != nil {
			return err
		}
	}

	return nil
}

func QueryProtoPluginRunner(p *protogen.Plugin) error {
	p.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	for _, f := range p.Files {
		if !f.Generate {
			continue
		}

		if !hasTables(f) {
			continue
		}

		out, err := os.OpenFile(fmt.Sprintf("%s_query.proto", f.GeneratedFilenamePrefix), os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0o644)
		if err != nil {
			return err
		}

		err = queryProtoGen{
			File:    f,
			svc:     newWriter(),
			msgs:    newWriter(),
			outFile: out,
			imports: map[string]bool{},
		}.gen()
		if err != nil {
			return err
		}
	}

	return nil
}

func hasTables(file *protogen.File) bool {
	for _, message := range file.Messages {
		if proto.GetExtension(message.Desc.Options(), ormv1.E_Table).(*ormv1.TableDescriptor) != nil {
			return true
		}

		if proto.GetExtension(message.Desc.Options(), ormv1.E_Singleton).(*ormv1.SingletonDescriptor) != nil {
			return true
		}
	}

	return false
}
