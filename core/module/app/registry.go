package app

import "github.com/cosmos/cosmos-sdk/core/module"

var defaultRegistry = module.NewRegistry((interface{})(nil), (*Handler)(nil))

func DefaultRegistry() *module.Registry {
	return defaultRegistry
}

//var registry map[reflect.Type]Handler
//
//func RegisterAppModule(constructor interface{}) {
//	typ := reflect.TypeOf(constructor)
//	if typ.Kind() != reflect.Func {
//		panic("TODO")
//	}
//
//	if typ.NumIn() < 1 {
//		panic("TODO")
//	}
//
//	typ.In(1)
//
//}
//
