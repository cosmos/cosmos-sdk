package types

import abci "github.com/tendermint/tendermint/abci/types"

type CustomQueryable interface {
	Query(ctx Context, path []string, req abci.RequestQuery) (res []byte, err Error)
}
