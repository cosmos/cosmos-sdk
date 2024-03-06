package accounts

import (
	"google.golang.org/protobuf/proto"

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

type gogoProtoPlusV2 interface {
	proto.Message
	implementation.ProtoMsg
}

// protoV2GogoWrapper is a wrapper of a protov2 message into a gogo message.
// this is exceptionally allowed to enable accounts to be decoupled from
// the SDK, since x/accounts can support only protov1 in its APIs.
// But in order to keep it decoupled from the SDK we need to use the API module.
// This is a hack to make an API module type work in x/accounts. Once the SDK
// has protov2 support, we can migrate internal/implementation/encoding.go to
// work with protov2.
type protoV2GogoWrapper struct {
	gogoProtoPlusV2
}

func (h protoV2GogoWrapper) XXX_MessageName() string {
	return string(proto.MessageName(h.gogoProtoPlusV2))
}

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
		return protoV2GogoWrapper{&bankv1beta1.MsgSend{
			FromAddress: fromAddr,
			ToAddress:   toAddr,
			Amount:      v2Coins,
		}}, new(bankv1beta1.MsgSendResponse), nil
	}
}
