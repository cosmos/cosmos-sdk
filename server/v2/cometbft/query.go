package cometbft

import (
	"context"
	"encoding/json"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/server/v2/core/store"
	abci "github.com/cometbft/cometbft/abci/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (c *Consensus[T]) handleQueryP2P(path []string) (*abci.QueryResponse, error) {
	// "/p2p" prefix for p2p queries
	if len(path) < 4 {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "path should be p2p filter <addr|id> <parameter>")
	}

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
	}

	return nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "expected second parameter to be 'filter'")
}

func (c *Consensus[T]) handlerQueryApp(ctx context.Context, path []string, req *abci.QueryRequest) (*abci.QueryResponse, error) {
	if len(path) < 2 {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrUnknownRequest,
			"expected second parameter to be either 'simulate' or 'version', neither was present",
		)
	}

	switch path[1] {
	case "simulate":
		// TODO: is this context fine?
		txResult, err := c.app.Simulate(ctx, req.Data)
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to simulate tx")
		}

		// TODO: encode txResult, we use codec.ProtoMarshalJSON in baseapp
		bz, err := json.Marshal(txResult)
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to marshal txResult")
		}

		return &abci.QueryResponse{
			Codespace: "sdk", // TODO: put this in a const
			Value:     bz,
			Height:    req.Height,
		}, nil

	case "version":
		return &abci.QueryResponse{
			Codespace: "sdk", // TODO: put this in a const
			Value:     []byte(c.version),
			Height:    req.Height,
		}, nil
	}

	return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query: %s", path)
}

func (c *Consensus[T]) handleQueryStore(path []string, store store.Store, req *abci.QueryRequest) (*abci.QueryResponse, error) {
	req.Path = "/" + strings.Join(path[1:], "/")
	if req.Height <= 1 && req.Prove {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"cannot query with proof when height <= 1; please provide a valid height",
		)
	}

	st, err := store.ReadonlyStateAt(uint64(req.Height))
	if err != nil {
		return nil, err
	}

	// TODO: revisit this, i need to parse the path
	bz, err := st.Get([]byte(req.Path))
	if err != nil {
		return nil, err
	}
	// TODO: also proving is not implemented
	return &abci.QueryResponse{
		Codespace: "sdk", // TODO: put this in a const
		Value:     bz,
		Height:    req.Height,
		Key:       []byte(req.Path),
	}, nil
}
