package wasmtime

import (
	"context"

	"cosmossdk.io/core/intermodule"
	"github.com/bytecodealliance/wasmtime-go/v14"
)

func AddSearchPaths(paths ...string) {

}

type fnType = func(caller *wasmtime.Caller, vals []wasmtime.Val) ([]wasmtime.Val, *wasmtime.Trap)

func wrapClientInvoke(global globalContext, client intermodule.Client) fnType {
	return func(caller *wasmtime.Caller, vals []wasmtime.Val) ([]wasmtime.Val, *wasmtime.Trap) {
		//ctx, ok := global.contexts[vals[0].I32()]
		//if !ok {
		//	panic("invalid context")
		//}
		//
		//target := vals[1].I32()
		//
		//reqPtr := vals[2].Externref()
		//resPtr := vals[3].Externref()

		// context, req, res
		panic("not implemented")
	}
}

type globalContext struct {
	contexts map[int32]context.Context
}
