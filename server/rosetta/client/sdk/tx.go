package sdk

import (
	"context"
	"github.com/cosmos/cosmos-sdk/x/auth/client"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetTx gets the transaction given its hex hash
func (c Client) GetTx(_ context.Context, hash string) (sdk.TxResponse, error) {
	tx, err := client.QueryTx(c.clientCtx, hash)
	if err != nil {
		return sdk.TxResponse{}, err
	}
	return *tx, nil
}

// PostTx broadcasts raw transaction bytes
func (c Client) PostTx(_ context.Context, bz []byte) (sdk.TxResponse, error) {
	resp, err := c.clientCtx.BroadcastTx(bz)
	if err != nil {
		return sdk.TxResponse{}, err
	}
	return *resp, nil
}
