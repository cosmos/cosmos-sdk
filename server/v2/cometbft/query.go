package cometbft

import (
	"context"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/proto/tendermint/crypto"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/server/v2/cometbft/types"
	cometerrors "cosmossdk.io/server/v2/cometbft/types/errors"
)

func (c *Consensus[T]) handleQueryP2P(path []string) (*abci.ResponseQuery, error) {
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

// handlerQueryApp handles the query requests for the application.
// It expects the path parameter to have at least two elements.
// The second element of the path can be either 'simulate' or 'version'.
// If the second element is 'simulate', it decodes the request data into a transaction,
// simulates the transaction using the application, and returns the simulation result.
// If the second element is 'version', it returns the version of the application.
// If the second element is neither 'simulate' nor 'version', it returns an error indicating an unknown query.
func (c *Consensus[T]) handlerQueryApp(ctx context.Context, path []string, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
	if len(path) < 2 {
		return nil, errorsmod.Wrap(
			cometerrors.ErrUnknownRequest,
			"expected second parameter to be either 'simulate' or 'version', neither was present",
		)
	}

	switch path[1] {
	case "simulate":
		tx, err := c.txCodec.Decode(req.Data)
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to decode tx")
		}

		txResult, _, err := c.app.Simulate(ctx, tx)
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to simulate tx")
		}

		bz, err := intoABCISimulationResponse(txResult, c.cfg.IndexEvents)
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to marshal txResult")
		}

		return &abci.ResponseQuery{
			Codespace: cometerrors.RootCodespace,
			Value:     bz,
			Height:    req.Height,
		}, nil

	case "version":
		return &abci.ResponseQuery{
			Codespace: cometerrors.RootCodespace,
			Value:     []byte(c.cfg.Version),
			Height:    req.Height,
		}, nil
	}

	return nil, errorsmod.Wrapf(cometerrors.ErrUnknownRequest, "unknown query: %s", path)
}

func (c *Consensus[T]) handleQueryStore(path []string, st types.Store, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
	req.Path = "/" + strings.Join(path[1:], "/")
	if req.Height <= 1 && req.Prove {
		return nil, errorsmod.Wrap(
			cometerrors.ErrInvalidRequest,
			"cannot query with proof when height <= 1; please provide a valid height",
		)
	}

	// "/store/<storeName>" for store queries
	storeName := path[1]
	qRes, err := c.store.Query(storeName, uint64(req.Height), req.Data, req.Prove)
	if err != nil {
		return nil, err
	}

	res := &abci.ResponseQuery{
		Codespace: cometerrors.RootCodespace,
		Height:    int64(qRes.Version()),
		Key:       qRes.Key(),
		Value:     qRes.Value(),
	}

	if req.Prove {
		bz, err := qRes.Proof().Marshal()
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to marshal proof")
		}

		res.ProofOps = &crypto.ProofOps{
			Ops: []crypto.ProofOp{
				{
					Type: qRes.ProofType(),
					Key:  qRes.Key(),
					Data: bz,
				},
			},
		}
	}

	return res, nil
}
