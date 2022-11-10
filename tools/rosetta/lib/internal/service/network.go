package service

import (
	"context"

	"cosmossdk.io/tools/rosetta/lib/errors"
	"github.com/coinbase/rosetta-sdk-go/types"
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

	// if genesis block could not be reached before
	if on.genesisBlockIdentifier == nil {
		var genesisBlockHeight int64 = 1
		genesisBlockIdentifier, err := on.client.BlockByHeight(ctx, &genesisBlockHeight)
		if err == nil {
			// once genesis is queryable, updates it to avoid future calls
			on.genesisBlockIdentifier = genesisBlockIdentifier.Block // once genesis is queryable, updates OnlineNetwork genesisBlockIdentifier
		}
	}

	var oldestBlockHeight int64 = -1
	oldestBlockIdentifier, err := on.client.BlockByHeight(ctx, &oldestBlockHeight)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}

	peers, err := on.client.Peers(ctx)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}

	return &types.NetworkStatusResponse{
		CurrentBlockIdentifier: block.Block,
		CurrentBlockTimestamp:  block.MillisecondTimestamp,
		GenesisBlockIdentifier: on.genesisBlockIdentifier,
		OldestBlockIdentifier:  oldestBlockIdentifier.Block,
		SyncStatus:             syncStatus,
		Peers:                  peers,
	}, nil
}
