package app

import (
	"fmt"
	"reflect"
	"strings"

	proto "github.com/gogo/protobuf/proto"
	"go.uber.org/dig"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/container"
)

func addProtoModule(name string, config *codectypes.Any) container.Option {
	// unpack Any
	msgTyp := proto.MessageType(config.TypeUrl)
	mod := reflect.New(msgTyp.Elem()).Interface().(proto.Message)
	if err := proto.Unmarshal(config.Value, mod); err != nil {
		return container.Error(err)
	}

	return addModule(name, mod)
}

var moduleKeyType = reflect.TypeOf((*ModuleKey)(nil)).Elem()

type codecClosureOutput struct {
	container.Out

	CodecClosure codecClosure `group:"codec"`
}

func addModule(name string, mod interface{}) container.Option {
	var opts []container.Option

	if typeProvider, ok := mod.(TypeProvider); ok {
		opts = append(opts, container.Provide(func() codecClosureOutput {
			return codecClosureOutput{
				CodecClosure: func(registry codectypes.TypeRegistry) {
					typeProvider.RegisterTypes(registry)
				},
			}
		}))
	}

	key := &moduleKey{&moduleID{name}}

	// register types
	// register DI Provide* methods
	modTy := reflect.TypeOf(mod)
	numMethods := modTy.NumMethod()
	for i := 0; i < numMethods; i++ {
		method := modTy.Method(i)
		if strings.HasPrefix(method.Name, "Provide") {
			methTyp := method.Type
			if methTyp.IsVariadic() {
				return container.Error(fmt.Errorf("variadic Provide methods are not supported"))
			}

			m := methTyp.NumIn() - 1
			fmt.Printf("method: %+v, m:%d\n", method, m)
			if m < 0 {
				return container.Error(fmt.Errorf("unexpected number of method arguments %d", m))
			}

			moduleKeyIdx := -1
			var in []reflect.Type
			for i := 0; i < m; i++ {
				ty := methTyp.In(i + 1)
				fmt.Printf("i:%d, ty:%+v\n", i, ty)
				if ty == moduleKeyType {
					if moduleKeyIdx >= 0 {
						modTy := reflect.TypeOf(mod)
						if modTy.Kind() == reflect.Ptr {
							modTy = modTy.Elem()
						}
						return container.Error(fmt.Errorf("%v can only be used as a parameter in a Provide method once, used twice in %s.%s.%s", moduleKeyType,
							modTy.PkgPath(), modTy.Name(), method.Name))
					}
					moduleKeyIdx = i + 1
				} else {
					in = append(in, ty)
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
				fmt.Printf("method: %+v, args:%+v, fnTy: %+v\n", method, args, fnTyp)
				return method.Func.Call(args)
			})

			opts = append(opts, container.Provide(fn.Interface()))
		}
	}

	return container.Options(opts...)
}

type registrar struct {
	ctr *dig.Container
}

var _ container.Registrar = registrar{}

func (s registrar) Provide(fn interface{}) error {
	return s.ctr.Provide(fn)
}

func ComposeModules(modules map[string]*codectypes.Any) container.Option {
	var opts []container.Option
	for name, mod := range modules {
		opts = append(opts, addProtoModule(name, mod))
	}
	return container.Options(opts...)
}

type codecInputs struct {
	container.In

	CodecClosures []codecClosure `group:"codec"`
}

type codecClosure func(codectypes.TypeRegistry)

var CodecProvider = container.Provide(func(inputs codecInputs) (
	codectypes.TypeRegistry,
	codec.Codec,
	codec.ProtoCodecMarshaler,
	codec.BinaryCodec,
	codec.JSONCodec,
	*codec.LegacyAmino,
) {

	typeRegistry := codectypes.NewInterfaceRegistry()
	for _, closure := range inputs.CodecClosures {
		closure(typeRegistry)
	}
	cdc := codec.NewProtoCodec(typeRegistry)
	amino := codec.NewLegacyAmino()
	return typeRegistry, cdc, cdc, cdc, cdc, amino
})
