package cometbft

import (
	errorsmod "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (c *Consensus[T]) handleQueryP2P(path []string) *abci.QueryResponse {
	// "/p2p" prefix for p2p queries
	if len(path) < 4 {
		return QueryResult(errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "path should be p2p filter <addr|id> <parameter>"), c.trace)
	}

	resp := &abci.QueryResponse{}

	cmd, typ, arg := path[1], path[2], path[3]
	switch cmd {
	case "filter":
		switch typ {
		case "addr":
			if c.addrPeerFilter != nil {
				return c.addrPeerFilter(arg)
			}

		case "id":
			if c.idPeerFilter != nil {
				return c.idPeerFilter(arg)
			}
		}
	default:
		resp = QueryResult(errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "expected second parameter to be 'filter'"), c.trace)
	}

	return resp
}
