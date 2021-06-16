package internal

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/gogo/protobuf/proto"
)

type ModuleContainer struct {
	*container.Container
	typeRegistry codecTypes.TypeRegistry
	codec        codec.Codec
	modules      map[container.Scope]interface{}
}

func NewModuleContainer() *ModuleContainer {
	typeRegistry := codecTypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(typeRegistry)
	amino := codec.NewLegacyAmino()
	ctr := container.NewContainer()
	err := ctr.Provide(func() (
		codecTypes.TypeRegistry,
		codec.Codec,
		codec.BinaryCodec,
		codec.JSONCodec,
		*codec.LegacyAmino,
	) {
		return typeRegistry, cdc, cdc, cdc, amino
	})
	if err != nil {
		panic(err)
	}

	return &ModuleContainer{
		Container:    ctr,
		typeRegistry: typeRegistry,
		codec:        cdc,
		modules:      map[container.Scope]interface{}{},
	}
}

func (mc ModuleContainer) AddModule(scope container.Scope, config *codecTypes.Any) error {
	// unpack Any
	msgTyp := proto.MessageType(config.TypeUrl)
	mod := reflect.New(msgTyp).Interface().(proto.Message)
	if err := proto.Unmarshal(config.Value, mod); err != nil {
		return err
	}

	mc.modules[scope] = mod

	// register types
	if typeProvider, ok := mod.(app.TypeProvider); ok {
		typeProvider.RegisterTypes(mc.typeRegistry)
	}

	// register DI providers
	if provisioner, ok := mod.(app.Provisioner); ok {
		registrar := scopedRegistrar{
			ctr:   mc.Container,
			scope: scope,
		}
		err := provisioner.Provision(nil, registrar)
		if err != nil {
			return err
		}
	}

	return nil
}

type scopedRegistrar struct {
	ctr   *container.Container
	scope container.Scope
}

var _ container.Registrar = scopedRegistrar{}

func (s scopedRegistrar) Provide(fn interface{}) error {
	return s.ctr.ProvideWithScope(fn, s.scope)
}
