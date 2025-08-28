package testutil

import (
	"context"

	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ bank.MsgServer = MockBankKeeper{}

func (k MockBankKeeper) Send(context.Context, *bank.MsgSend) (*bank.MsgSendResponse, error) {
	return nil, nil
}

func (k MockBankKeeper) MultiSend(context.Context, *bank.MsgMultiSend) (*bank.MsgMultiSendResponse, error) {
	return nil, nil
}

func (k MockBankKeeper) UpdateParams(context.Context, *bank.MsgUpdateParams) (*bank.MsgUpdateParamsResponse, error) {
	return nil, nil
}

func (k MockBankKeeper) SetSendEnabled(context.Context, *bank.MsgSetSendEnabled) (*bank.MsgSetSendEnabledResponse, error) {
	return nil, nil
}
