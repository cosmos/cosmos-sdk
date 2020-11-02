package rosetta

import (
	"context"

	"github.com/tendermint/cosmos-rosetta-gateway/rosetta"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinbase/rosetta-sdk-go/types"
)

func (l launchpad) AccountBalance(ctx context.Context, request *types.AccountBalanceRequest) (
	*types.AccountBalanceResponse, *types.Error) {

	if l.properties.OfflineMode {
		return nil, ErrEndpointDisabledOfflineMode
	}

	var reqHeight int64
	if request.BlockIdentifier != nil {
		reqHeight = *request.BlockIdentifier.Index
	}
	resp, err := l.cosmos.GetAuthAccount(ctx, request.AccountIdentifier.Address, reqHeight)
	if err != nil {
		return nil, rosetta.WrapError(ErrNodeConnection, err.Error())
	}

	block, err := l.tendermint.Block(uint64(resp.Height))
	if err != nil {
		return nil, rosetta.WrapError(ErrNodeConnection, err.Error())
	}

	return &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: resp.Height,
			Hash:  block.BlockID.Hash,
		},
		Balances: convertCoinsToRosettaBalances(resp.Result.Value.Coins),
		Coins: []*types.Coin{
			{
				CoinIdentifier: &types.CoinIdentifier{Identifier: "atom"},
				Amount:         convertCoinsToRosettaBalances(resp.Result.Value.Coins)[1],
			},
		},
	}, nil
}

func convertCoinsToRosettaBalances(coins []sdk.Coin) []*types.Amount {
	amounts := make([]*types.Amount, len(coins))

	for i, coin := range coins {
		amounts[i] = &types.Amount{
			Value: coin.Amount.String(),
			Currency: &types.Currency{
				Symbol: coin.Denom,
			},
		}
	}

	return amounts
}
