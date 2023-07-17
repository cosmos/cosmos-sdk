package service

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"

	"cosmossdk.io/tools/rosetta/lib/errors"
)

func (on OnlineNetwork) NetworkList(_ context.Context, _ *types.MetadataRequest) (*types.NetworkListResponse, *types.Error) {
	return &types.NetworkListResponse{NetworkIdentifiers: []*types.NetworkIdentifier{on.network}}, nil
}

func (on OnlineNetwork) NetworkOptions(_ context.Context, _ *types.NetworkRequest) (*types.NetworkOptionsResponse, *types.Error) {
	return on.networkOptions, nil
}

func (on OnlineNetwork) NetworkStatus(ctx context.Context, _ *types.NetworkRequest) (*types.NetworkStatusResponse, *types.Error) {
	syncStatus, err := on.client.Status(ctx)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}

	block, err := on.client.BlockByHeight(ctx, syncStatus.CurrentIndex)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}

	oldestBlockIdentifier, err := on.client.OldestBlock(ctx)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}

	genesisBlock, err := on.client.GenesisBlock(ctx)
	if err != nil {
		genesisBlock, err = on.client.InitialHeightBlock(ctx)
		if err != nil {
			genesisBlock = oldestBlockIdentifier
		}
	}

	peers, err := on.client.Peers(ctx)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}

	return &types.NetworkStatusResponse{
		CurrentBlockIdentifier: block.Block,
		CurrentBlockTimestamp:  block.MillisecondTimestamp,
		GenesisBlockIdentifier: genesisBlock.Block,
		OldestBlockIdentifier:  oldestBlockIdentifier.Block,
		SyncStatus:             syncStatus,
		Peers:                  peers,
	}, nil
}
