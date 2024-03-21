package accounts

import (
	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	v1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/x/accounts/internal/implementation"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// coinsTransferMsgFunc defines a function that creates a message to send coins from one
// address to the other, and also a message that parses such  response.
// This in most cases will be implemented as a bank.MsgSend creator, but we keep x/accounts independent of bank.
type coinsTransferMsgFunc = func(from, to []byte, coins sdk.Coins) (implementation.ProtoMsg, implementation.ProtoMsg, error)

func defaultCoinsTransferMsgFunc(addrCdc address.Codec) coinsTransferMsgFunc {
	return func(from, to []byte, coins sdk.Coins) (implementation.ProtoMsg, implementation.ProtoMsg, error) {
		fromAddr, err := addrCdc.BytesToString(from)
		if err != nil {
			return nil, nil, err
		}
		toAddr, err := addrCdc.BytesToString(to)
		if err != nil {
			return nil, nil, err
		}
		v2Coins := make([]*v1beta1.Coin, len(coins))
		for i, coin := range coins {
			v2Coins[i] = &v1beta1.Coin{
				Denom:  coin.Denom,
				Amount: coin.Amount.String(),
			}
		}
		return &bankv1beta1.MsgSend{
			FromAddress: fromAddr,
			ToAddress:   toAddr,
			Amount:      v2Coins,
		}, new(bankv1beta1.MsgSendResponse), nil
	}
}
