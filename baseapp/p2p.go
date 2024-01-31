package baseapp

// This file exists because Tendermint allows the application to control which peers it connects to.
// This is for an interesting idea -- allow the application to control the peer layer/ topology!
// It would be really exciting to mix web of trust and expander-graph style primitives
// for how information gets disseminated.
// However the API surface for this to make sense isn't really well exposed / thought through,
// so this file mostly acts as confusing boilerplate.

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

type peerFilters struct {
	addrPeerFilter sdk.PeerFilter // filter peers by address and port
	idPeerFilter   sdk.PeerFilter // filter peers by node ID
}

// FilterPeerByAddrPort filters peers by address/port.
func (app *BaseApp) FilterPeerByAddrPort(info string) abci.ResponseQuery {
	if app.addrPeerFilter != nil {
		return app.addrPeerFilter(info)
	}

	return abci.ResponseQuery{}
}

// FilterPeerByID filters peers by node ID.
func (app *BaseApp) FilterPeerByID(info string) abci.ResponseQuery {
	if app.idPeerFilter != nil {
		return app.idPeerFilter(info)
	}

	return abci.ResponseQuery{}
}

func handleQueryP2P(app *BaseApp, path []string) abci.ResponseQuery {
	// "/p2p" prefix for p2p queries
	if len(path) < 4 {
		return sdkerrors.QueryResult(
			sdkerrors.Wrap(
				sdkerrors.ErrUnknownRequest, "path should be p2p filter <addr|id> <parameter>",
			),
			app.trace,
		)
	}

	var resp abci.ResponseQuery

	cmd, typ, arg := path[1], path[2], path[3]
	switch cmd {
	case "filter":
		switch typ {
		case "addr":
			resp = app.FilterPeerByAddrPort(arg)

		case "id":
			resp = app.FilterPeerByID(arg)
		}

	default:
		resp = sdkerrors.QueryResult(sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "expected second parameter to be 'filter'"), app.trace)
	}

	return resp
}
