package utils

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// TODO: remove hardcoded storename
const storeName = "params"

// Query parameters from node with CLIContext
func QueryParams(cliCtx context.CLIContext, subStoreName string, ps params.ParamStruct) error {
	m := make(map[string][]byte)

	for _, p := range ps.KeyFieldPairs() {
		key := p.Key
		bz, err := cliCtx.QueryStore([]byte(subStoreName+"/"+key), storeName)
		if err != nil {
			return err
		}
		m[key] = bz
	}

	return params.UnmarshalParamsFromMap(m, cliCtx.Codec, ps)
}
