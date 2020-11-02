package sdk

import (
	"context"
	"fmt"

	types2 "github.com/cosmos/cosmos-sdk/x/bank/types"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk/types"
)

func (c Client) GetAuthAccount(ctx context.Context, address string, height int64) (types.AccountResponse, error) {
	// update the context metadata to add the height header information
	ctx = context.WithValue(ctx, grpctypes.GRPCBlockHeightHeader, height)

	// get account information
	rawAccount, err := c.authClient.Account(ctx, &auth.QueryAccountRequest{Address: address})
	if err != nil {
		return types.AccountResponse{}, err
	}

	// decode any to raw account
	var account auth.AccountI
	err = c.encodeConfig.UnpackAny(rawAccount.Account, &account)
	if err != nil {
		return types.AccountResponse{}, err
	}

	// get balance information
	balances, err := c.bankClient.AllBalances(ctx, &types2.QueryAllBalancesRequest{
		Address:    address,
		Pagination: nil,
	})
	if err != nil {
		return types.AccountResponse{}, err
	}

	// transform response
	resp := types.AccountResponse{
		Height: height,
		Result: types.Response{
			Type: "", // type does not apply here as it technically is multiple types
			Value: types.BaseAccount{
				Address: address,
				Coins:   balances.Balances,
				PubKey: types.PublicKey{
					Type:  account.GetPubKey().Type(),
					Value: fmt.Sprintf("%x", account.GetPubKey().Bytes()), // is this correct?
				},
				AccountNumber: sdk.NewIntFromUint64(account.GetAccountNumber()).String(),
				Sequence:      sdk.NewIntFromUint64(account.GetSequence()).String(),
			},
		},
	}

	// success
	return resp, nil
}
