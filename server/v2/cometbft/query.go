package cometbft

import (
	"context"
	"strings"

	errorsmod "cosmossdk.io/errors"
	cometerrors "cosmossdk.io/server/v2/cometbft/types/errors"
	"cosmossdk.io/server/v2/core/store"
	abci "github.com/cometbft/cometbft/abci/types"
)

func (c *Consensus[T]) handleQueryP2P(path []string) (*abci.QueryResponse, error) {
	// "/p2p" prefix for p2p queries
	if len(path) < 4 {
		return nil, errorsmod.Wrap(cometerrors.ErrUnknownRequest, "path should be p2p filter <addr|id> <parameter>")
	}

	cmd, typ, arg := path[1], path[2], path[3]
	switch cmd {
	case "filter":
		switch typ {
		case "addr":
			if c.cfg.AddrPeerFilter != nil {
				return c.cfg.AddrPeerFilter(arg)
			}

		case "id":
			if c.cfg.IdPeerFilter != nil {
				return c.cfg.IdPeerFilter(arg)
			}
		}
	}

	return nil, errorsmod.Wrap(cometerrors.ErrUnknownRequest, "expected second parameter to be 'filter'")
}

func (c *Consensus[T]) handlerQueryApp(ctx context.Context, path []string, req *abci.QueryRequest) (*abci.QueryResponse, error) {
	if len(path) < 2 {
		return nil, errorsmod.Wrap(
			cometerrors.ErrUnknownRequest,
			"expected second parameter to be either 'simulate' or 'version', neither was present",
		)
	}

	switch path[1] {
	case "simulate":
		txResult, err := c.app.Simulate(ctx, req.Data)
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to simulate tx")
		}

		bz, err := intoABCISimulationResponse(txResult, c.cfg.IndexEvents)
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to marshal txResult")
		}

		return &abci.QueryResponse{
			Codespace: cometerrors.RootCodespace,
			Value:     bz,
			Height:    req.Height,
		}, nil

	case "version":
		return &abci.QueryResponse{
			Codespace: cometerrors.RootCodespace,
			Value:     []byte(c.cfg.Version),
			Height:    req.Height,
		}, nil
	}

	return nil, errorsmod.Wrapf(cometerrors.ErrUnknownRequest, "unknown query: %s", path)
}

func (c *Consensus[T]) handleQueryStore(path []string, store store.Store, req *abci.QueryRequest) (*abci.QueryResponse, error) {
	req.Path = "/" + strings.Join(path[1:], "/")
	if req.Height <= 1 && req.Prove {
		return nil, errorsmod.Wrap(
			cometerrors.ErrInvalidRequest,
			"cannot query with proof when height <= 1; please provide a valid height",
		)
	}

	st, err := store.StateAt(uint64(req.Height))
	if err != nil {
		return nil, err
	}

	// TODO: key format has changed or should we do the same as in baseapp?
	bz, err := st.Get([]byte(req.Path))
	if err != nil {
		return nil, err
	}
	// TODO: also proving is not implemented
	return &abci.QueryResponse{
		Codespace: cometerrors.RootCodespace,
		Value:     bz,
		Height:    req.Height,
		Key:       []byte(req.Path),
	}, nil
}
