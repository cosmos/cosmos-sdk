package codegen

import (
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"

	"google.golang.org/protobuf/compiler/protogen"
)

var (
	tablePkg = protogen.GoImportPath("github.com/cosmos/cosmos-sdk/orm/model/ormtable")
)

type fileGen struct {
	*protogen.GeneratedFile
	file *protogen.File
}

func (f fileGen) gen() error {
	f.P("package ", f.file.GoPackageName)
	f.genStoreAccessor()
	f.genStoreInterface()
	for _, msg := range f.file.Messages {
		//opts := proto.GetExtension(msg.Desc.Options(), v1alpha1.E_Table).(*v1alpha1.TableDescriptor)
		//if opts == nil {
		//	continue
		//}
		//err := genMsg(gen, opts, msg)
		//if err != nil {
		//	return fmt.Errorf("unable to generate message %s in file %s", msg.Desc.FullName(), f.Desc.Path())
		//}
		tableGen{
			fileGen: f,
			msg:     msg,
		}.gen()
	}
	f.genStoreStruct()
	return nil
}

func (f fileGen) genStoreAccessor() {
	f.P("type ", f.storeAccessorName(), " interface {")
	f.P("Open(", contextPkg.Ident("Context"), ")", f.storeInterfaceName())
	f.P("}")
	f.P()

	// constructor
	f.P("func New", f.storeAccessorName(), "() (", f.storeAccessorName(), ", error) {")
	f.P("}")
	f.P()
}

func (f fileGen) genStoreInterface() {
	f.P("type ", f.storeInterfaceName(), " interface {")
	for _, message := range f.file.Messages {
		f.P(f.messageStoreInterfaceName(message))
	}
	f.P("}")
	f.P()
}

func (f fileGen) genStoreStruct() {
	// struct
	f.P("type ", f.storeStructName(), " struct {")
	for _, message := range f.file.Messages {
		f.P(f.messageTableVar(message), " ", tablePkg.Ident("Table"))
	}
	f.P("}")
}

func (f fileGen) storeAccessorName() string {
	return f.storeInterfaceName() + "Accessor"
}

func (f fileGen) storeInterfaceName() string {
	return strcase.ToCamel(f.fileShortName()) + "Store"
}

func (f fileGen) storeStructName() string {
	return strcase.ToLowerCamel(f.fileShortName()) + "Store"
}

func (f fileGen) fileShortName() string {
	filename := f.file.Proto.GetName()
	shortName := filepath.Base(filename)
	i := strings.Index(shortName, ".")
	if i > 0 {
		return shortName[:i]
	}
	return shortName
}

func (f fileGen) messageStoreInterfaceName(m *protogen.Message) string {
	return m.GoIdent.GoName + "Store"
}

func (f fileGen) messageTableVar(m *protogen.Message) string {
	return f.param(m.GoIdent.GoName + "Table")
}

func (f fileGen) param(name string) string {
	return strcase.ToLowerCamel(name)
}
