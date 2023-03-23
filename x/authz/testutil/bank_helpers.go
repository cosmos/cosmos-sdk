package testutil

import (
	"context"

	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ bank.MsgServer = MockBankKeeper{}

func (k MockBankKeeper) Send(_ context.Context, _ *bank.MsgSend) (*bank.MsgSendResponse, error) {
	return nil, nil
}

func (k MockBankKeeper) MultiSend(_ context.Context, _ *bank.MsgMultiSend) (*bank.MsgMultiSendResponse, error) {
	return nil, nil
}

func (k MockBankKeeper) UpdateParams(_ context.Context, _ *bank.MsgUpdateParams) (*bank.MsgUpdateParamsResponse, error) {
	return nil, nil
}

func (k MockBankKeeper) SetSendEnabled(_ context.Context, _ *bank.MsgSetSendEnabled) (*bank.MsgSetSendEnabledResponse, error) {
	return nil, nil
}
