package service

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"

	"cosmossdk.io/tools/rosetta/lib/errors"
	crgtypes "cosmossdk.io/tools/rosetta/lib/types"
)

// AccountBalance retrieves the account balance of an address
// rosetta requires us to fetch the block information too
func (on OnlineNetwork) AccountBalance(ctx context.Context, request *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {
	var (
		height int64
		block  crgtypes.BlockResponse
		err    error
	)

	switch {
	case request.BlockIdentifier == nil:
		syncStatus, err := on.client.Status(ctx)
		if err != nil {
			return nil, errors.ToRosetta(err)
		}
		block, err = on.client.BlockByHeight(ctx, syncStatus.CurrentIndex)
		if err != nil {
			return nil, errors.ToRosetta(err)
		}
	case request.BlockIdentifier.Hash != nil:
		block, err = on.client.BlockByHash(ctx, *request.BlockIdentifier.Hash)
		if err != nil {
			return nil, errors.ToRosetta(err)
		}
		height = block.Block.Index
	case request.BlockIdentifier.Index != nil:
		height = *request.BlockIdentifier.Index
		block, err = on.client.BlockByHeight(ctx, &height)
		if err != nil {
			return nil, errors.ToRosetta(err)
		}
	}

	accountCoins, err := on.client.Balances(ctx, request.AccountIdentifier.Address, &height)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}

	return &types.AccountBalanceResponse{
		BlockIdentifier: block.Block,
		Balances:        accountCoins,
		Metadata:        nil,
	}, nil
}

// AccountsCoins - relevant only for UTXO based chain
// see https://www.rosetta-api.org/docs/AccountApi.html#accountcoins
func (on OnlineNetwork) AccountCoins(_ context.Context, _ *types.AccountCoinsRequest) (*types.AccountCoinsResponse, *types.Error) {
	return nil, errors.ToRosetta(errors.ErrOffline)
}
