package stf

import (
	"context"

	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
)

var _ header.Service = (*HeaderService)(nil)

type HeaderService struct{}

func (h HeaderService) HeaderInfo(ctx context.Context) header.Info {
	exCtx, err := getExecutionCtxFromContext(ctx)
	if err != nil {
		panic(err)
	}

	return exCtx.headerInfo
}

const headerInfoPrefix = 0x37

// setHeaderInfo sets the header info in the state to be used by queries in the future.
func (s STF[T]) setHeaderInfo(state store.WriterMap, headerInfo header.Info) error {
	runtimeStore, err := state.GetWriter(Identity)
	if err != nil {
		return err
	}
	bz, err := headerInfo.Bytes()
	if err != nil {
		return err
	}
	err = runtimeStore.Set([]byte{headerInfoPrefix}, bz)
	if err != nil {
		return err
	}
	return nil
}

// getHeaderInfo gets the header info from the state. It should only be used for queries
func (s STF[T]) getHeaderInfo(state store.WriterMap) (i header.Info, err error) {
	runtimeStore, err := state.GetWriter(Identity)
	if err != nil {
		return header.Info{}, err
	}
	v, err := runtimeStore.Get([]byte{headerInfoPrefix})
	if err != nil {
		return header.Info{}, err
	}
	if v == nil {
		return header.Info{}, nil
	}

	err = i.FromBytes(v)
	return i, err
}
