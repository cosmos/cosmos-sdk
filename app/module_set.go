package app

import (
	"reflect"
	"strings"

	proto "github.com/gogo/protobuf/proto"
	"go.uber.org/dig"

	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/container"
)

type ModuleContainer struct {
	*dig.Container
	typeRegistry codecTypes.TypeRegistry
	codec        codec.Codec
	modules      map[ModuleKey]interface{}
}

func NewModuleContainer() *ModuleContainer {
	typeRegistry := codecTypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(typeRegistry)
	amino := codec.NewLegacyAmino()
	ctr := dig.New()
	err := ctr.Provide(func() (
		codecTypes.TypeRegistry,
		codec.Codec,
		codec.ProtoCodecMarshaler,
		codec.BinaryCodec,
		codec.JSONCodec,
		*codec.LegacyAmino,
	) {
		return typeRegistry, cdc, cdc, cdc, cdc, amino
	})
	if err != nil {
		panic(err)
	}

	return &ModuleContainer{
		Container:    ctr,
		typeRegistry: typeRegistry,
		codec:        cdc,
		modules:      map[ModuleKey]interface{}{},
	}
}

func (mc ModuleContainer) AddProtoModule(name string, config *codecTypes.Any) error {
	// unpack Any
	msgTyp := proto.MessageType(config.TypeUrl)
	mod := reflect.New(msgTyp.Elem()).Interface().(proto.Message)
	if err := proto.Unmarshal(config.Value, mod); err != nil {
		return err
	}

	return mc.AddModule(name, mod)
}

func (mc *ModuleContainer) AddModule(name string, mod interface{}) error {
	key := &moduleKey{&moduleID{name}}
	mc.modules[key] = mod

	// register types
	if typeProvider, ok := mod.(TypeProvider); ok {
		typeProvider.RegisterTypes(mc.typeRegistry)
	}

	// register DI providers
	if provisioner, ok := mod.(Provisioner); ok {
		registrar := registrar{
			ctr: mc.Container,
		}
		err := provisioner.Provision(nil, registrar)
		if err != nil {
			return err
		}
	}

	// register DI Provide* methods
	modTy := reflect.TypeOf(mod)
	numMethods := modTy.NumMethod()
	for i := 0; i < numMethods; i++ {
		method := modTy.Method(i)
		if strings.HasPrefix(method.Name, "Provide") {
			// TODO wrap function and provide
		}
	}

	return nil
}

type registrar struct {
	ctr *dig.Container
}

var _ container.Registrar = registrar{}

func (s registrar) Provide(fn interface{}) error {
	return s.ctr.Provide(fn)
}
