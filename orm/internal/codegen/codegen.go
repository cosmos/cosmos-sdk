package codegen

import (
	"fmt"
	"unicode"

	"github.com/cosmos/cosmos-proto/generator"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"

	v1alpha1 "github.com/cosmos/cosmos-sdk/api/cosmos/orm/v1alpha1"
)

const (
	contextPkg          = protogen.GoImportPath("context")
	protoreflectPackage = protogen.GoImportPath("google.golang.org/protobuf/reflect/protoreflect")
)

func PluginRunner(p *protogen.Plugin) error {
	for _, f := range p.Files {
		if !f.Generate {
			continue
		}

		gen := p.NewGeneratedFile(fmt.Sprintf("%s.cosmsos_orm.go", f.GeneratedFilenamePrefix), f.GoImportPath)
		cgen := &generator.GeneratedFile{
			GeneratedFile: gen,
			Ext:           nil,
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

func genMsg(g *protogen.GeneratedFile, opts *v1alpha1.TableDescriptor, msg *protogen.Message) error {
	return ormClientGenerator{
		GeneratedFile: g,
		table:         opts,
		msg:           msg,
	}.generate()
}

type ormClientGenerator struct {
	*protogen.GeneratedFile
	table *v1alpha1.TableDescriptor
	msg   *protogen.Message
}

func (g ormClientGenerator) generate() error {
	err := g.genIteratorInterface()
	if err != nil {
		return err
	}
	err = g.genIteratorType()
	if err != nil {
		return err
	}
	err = g.genInterface()
	if err != nil {
		return err
	}

	return nil
}

func (g ormClientGenerator) unexported() string {
	s := g.msg.GoIdent.GoName
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a) + "Client"
}

func (g ormClientGenerator) interfaceName() string {
	return fmt.Sprintf("%sStore", g.msg.GoIdent.GoName)
}

func (g ormClientGenerator) genInterface() error {
	g.P("type ", g.interfaceName(), " interface {")
	// check if singleton
	//switch g.table.Singleton {
	//case true:
	//	return g.genSingletonInterface()
	//case false:
	return g.genTableObjectInterface()
	//}
	g.P("}")
	g.P()
	return nil
}

func (g ormClientGenerator) genUnexportedType() error {
	g.P("type ", g.unexported(), " struct {")
	g.P("client ", tablePkg.Ident("Client"))
	g.P("}")
	g.P()
	return nil
}

func (g ormClientGenerator) genSingletonInterface() error {
	g.P("Get(ctx ", contextPkg.Ident("Context"), ") (*", g.QualifiedGoIdent(g.msg.GoIdent), ", error)")
	g.P("Create(ctx ", contextPkg.Ident("Context"), ", ", g.param(g.msg.GoIdent.GoName), " *", g.QualifiedGoIdent(g.msg.GoIdent), ") error")
	g.P("Update(ctx ", contextPkg.Ident("Context"), ",  ", g.param(g.msg.GoIdent.GoName), " * ", g.QualifiedGoIdent(g.msg.GoIdent), ") error")
	g.P("Delete(ctx ", contextPkg.Ident("Context"), ") error")
	g.P("}")
	g.P()
	return nil
}

func (g ormClientGenerator) param(name string) string {
	return strcase.ToLowerCamel(name)
}

func (g ormClientGenerator) genTableObjectInterface() error {
	g.P("Create(", g.param(g.msg.GoIdent.GoName), " *", g.QualifiedGoIdent(g.msg.GoIdent), ") error")
	g.P("Update(", g.param(g.msg.GoIdent.GoName), " *", g.QualifiedGoIdent(g.msg.GoIdent), ") error")
	g.P("Save(", g.param(g.msg.GoIdent.GoName), " *", g.QualifiedGoIdent(g.msg.GoIdent), ") error")
	g.P("Delete(", g.param(g.msg.GoIdent.GoName), " *", g.QualifiedGoIdent(g.msg.GoIdent), ") error")
	err := g.genTableObjectGet()
	if err != nil {
		return err
	}
	err = g.genTableObjectListInterfaceMethods()
	if err != nil {
		return err
	}
	g.P("}")
	g.P()

	return nil
}

func (g ormClientGenerator) genTableObjectGet() error {
	return nil
}

func (g ormClientGenerator) genTableObjectListInterfaceMethods() error {
	//for _, field := range g.table.SecondaryKeys {
	//	fd := g.msg.Desc.Fields().ByName(protoreflect.Name(field.FieldName))
	//	if fd == nil {
	//		return fmt.Errorf("field %s is not part of the message", field.FieldName)
	//	}
	//	goType := getGoType(fd.Kind())
	//	g.P("ListBy", strcase.ToCamel(field.FieldName), "(ctx ", contextPkg.Ident("Context"), ", ", g.param(field.FieldName), " ", goType, ") (", g.iteratorInterfaceName(), ", error)")
	//}
	//
	//g.P("List(ctx ", contextPkg.Ident("Context"), ", options ", clientPkg.Ident("ListOptions"), ") (", g.iteratorInterfaceName(), ", error)")
	return nil
}

func (g ormClientGenerator) genIteratorInterface() error {
	//if g.table.Singleton {
	//	return nil
	//}
	//g.P("type ", g.iteratorInterfaceName(), " interface {")
	//g.P(clientPkg.Ident("ObjectIterator"))
	//g.P("Get() ", "(", g.QualifiedGoIdent(g.msg.GoIdent), ", error)")
	//g.P("}")
	//g.P()
	return nil
}

func (g ormClientGenerator) genIteratorType() error {
	return nil
}

func (g ormClientGenerator) iteratorInterfaceName() string {
	return fmt.Sprintf("%sIterator", g.msg.GoIdent.GoName)
}

func getGoType(kind protoreflect.Kind) string {
	// TODO(Fdymylja): fill all types
	switch kind {
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.Int64Kind:
		return "int64"
	case protoreflect.Int32Kind:
		return "int32"
	case protoreflect.Uint64Kind:
		return "uint64"
	case protoreflect.Uint32Kind:
		return "uint32"
	default:
		panic("unsupported type " + kind.String())
	}
}
