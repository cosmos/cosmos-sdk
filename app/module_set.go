package app

import (
	"fmt"
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

var moduleKeyType = reflect.TypeOf((*ModuleKey)(nil)).Elem()

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
			methTyp := method.Type
			if methTyp.IsVariadic() {
				return fmt.Errorf("variadic Provide methods are not supported")
			}

			m := methTyp.NumIn() - 1
			if m < 0 {
				return fmt.Errorf("unexpected number of method arguments %d", m)
			}

			moduleKeyIdx := -1
			var in []reflect.Type
			for i := 0; i < m; i++ {
				ty := methTyp.In(i + 1)
				if ty == moduleKeyType {
					if moduleKeyIdx >= 0 {
						modTy := reflect.TypeOf(mod)
						if modTy.Kind() == reflect.Ptr {
							modTy = modTy.Elem()
						}
						return fmt.Errorf("%v can only be used as a parameter in a Provide method once, used twice in %s.%s.%s", moduleKeyType,
							modTy.PkgPath(), modTy.Name(), method.Name)
					}
					moduleKeyIdx = i + 1
				} else {
					in = append(in)
				}
			}

			n := methTyp.NumOut()
			out := make([]reflect.Type, n)
			for i := 0; i < n; i++ {
				out[i] = methTyp.Out(i)
			}

			fnTyp := reflect.FuncOf(in, out, false)

			fn := reflect.MakeFunc(fnTyp, func(args []reflect.Value) (results []reflect.Value) {
				args = append([]reflect.Value{reflect.ValueOf(mod)}, args...)
				if moduleKeyIdx >= 0 {
					args0 := append(args[:moduleKeyIdx], reflect.ValueOf(key))
					args = append(args0, args[moduleKeyIdx:]...)
				}
				fmt.Printf("modTy: %s, method: %+v, idx:%d, args:%+v\n", reflect.TypeOf(mod).Elem().PkgPath(), method, moduleKeyIdx, args)
				return method.Func.Call(args)
			})

			err := mc.Provide(fn.Interface())
			if err != nil {
				return err
			}
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
