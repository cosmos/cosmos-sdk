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

type moduleContainer struct {
	typeRegistry codecTypes.TypeRegistry
	cdc          codec.Codec
	opts         []container.Option
}

func (mc moduleContainer) AddProtoModule(name string, config *codecTypes.Any) error {
	// unpack Any
	msgTyp := proto.MessageType(config.TypeUrl)
	mod := reflect.New(msgTyp.Elem()).Interface().(proto.Message)
	if err := proto.Unmarshal(config.Value, mod); err != nil {
		return err
	}

	return mc.AddModule(name, mod)
}

var moduleKeyType = reflect.TypeOf((*ModuleKey)(nil)).Elem()

func (mc *moduleContainer) AddModule(name string, mod interface{}) error {
	key := &moduleKey{&moduleID{name}}

	// register types
	if typeProvider, ok := mod.(TypeProvider); ok {
		typeProvider.RegisterTypes(mc.typeRegistry)
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

			mc.opts = append(mc.opts, container.Provide(fn.Interface()))
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

func ProvideModules(modules map[string]*codecTypes.Any) container.Option {
	typeRegistry := codecTypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(typeRegistry)
	amino := codec.NewLegacyAmino()
	cdcProvider := container.Provide(func() (
		codecTypes.TypeRegistry,
		codec.Codec,
		codec.ProtoCodecMarshaler,
		codec.BinaryCodec,
		codec.JSONCodec,
		*codec.LegacyAmino,
	) {
		return typeRegistry, cdc, cdc, cdc, cdc, amino
	})

	mc := moduleContainer{
		typeRegistry: typeRegistry,
		cdc:          cdc,
		opts:         []container.Option{cdcProvider},
	}

	for name, mod := range modules {
		err := mc.AddProtoModule(name, mod)
		if err != nil {
			return container.Error(err)
		}
	}

	return container.Options(mc.opts...)
}
