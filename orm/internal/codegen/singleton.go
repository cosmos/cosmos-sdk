package codegen

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/dynamicpb"

	ormv1 "cosmossdk.io/api/cosmos/orm/v1"
	"cosmossdk.io/orm/model/ormtable"
)

type singletonGen struct {
	fileGen
	msg      *protogen.Message
	table    *ormv1.SingletonDescriptor
	ormTable ormtable.Table
}

func newSingletonGen(fileGen fileGen, msg *protogen.Message, table *ormv1.SingletonDescriptor) (*singletonGen, error) {
	s := &singletonGen{fileGen: fileGen, msg: msg, table: table}
	var err error
	s.ormTable, err = ormtable.Build(ormtable.Options{
		MessageType:         dynamicpb.NewMessageType(msg.Desc),
		SingletonDescriptor: table,
	})
	return s, err
}

func (s singletonGen) gen() {
	s.genInterface()
	s.genStruct()
	s.genInterfaceGuard()
	s.genMethods()
	s.genConstructor()
}

func (s singletonGen) genInterface() {
	s.P("// singleton store")
	s.P("type ", s.messageTableInterfaceName(s.msg), " interface {")
	s.P("Get(ctx ", contextPkg.Ident("Context"), ") (*", s.msg.GoIdent.GoName, ", error)")
	s.P("Save(ctx ", contextPkg.Ident("Context"), ", ", s.param(s.msg.GoIdent.GoName), "*", s.msg.GoIdent.GoName, ") error")
	s.P("}")
	s.P()
}

func (s singletonGen) genStruct() {
	s.P("type ", s.messageTableReceiverName(s.msg), " struct {")
	s.P("table ", ormTablePkg.Ident("Table"))
	s.P("}")
	s.P()
}

func (s singletonGen) genInterfaceGuard() {
	s.P("var _ ", s.messageTableInterfaceName(s.msg), " = ", s.messageTableReceiverName(s.msg), "{}")
}

func (s singletonGen) genMethods() {
	receiver := fmt.Sprintf("func (x %s) ", s.messageTableReceiverName(s.msg))
	varName := s.param(s.msg.GoIdent.GoName)
	// Get
	s.P(receiver, "Get(ctx ", contextPkg.Ident("Context"), ") (*", s.msg.GoIdent.GoName, ", error) {")
	s.P(varName, " := &", s.msg.GoIdent.GoName, "{}")
	s.P("_, err := x.table.Get(ctx, ", varName, ")")
	s.P("return ", varName, ", err")
	s.P("}")
	s.P()

	// Save
	s.P(receiver, "Save(ctx ", contextPkg.Ident("Context"), ", ", varName, " *", s.msg.GoIdent.GoName, ") error {")
	s.P("return x.table.Save(ctx, ", varName, ")")
	s.P("}")
	s.P()
}

func (s singletonGen) genConstructor() {
	iface := s.messageTableInterfaceName(s.msg)
	s.P("func New", iface, "(db ", ormTablePkg.Ident("Schema"), ") (", iface, ", error) {")
	s.P("table := db.GetTable(&", s.msg.GoIdent.GoName, "{})")
	s.P("if table == nil {")
	s.P("return nil, ", ormErrPkg.Ident("TableNotFound.Wrap"), "(string((&", s.msg.GoIdent.GoName, "{}).ProtoReflect().Descriptor().FullName()))")
	s.P("}")
	s.P("return &", s.messageTableReceiverName(s.msg), "{table}, nil")
	s.P("}")
}
