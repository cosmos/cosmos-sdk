package keeper_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/bank/testutil"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/tx/signing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ signing.SignModeHandler = directHandler{}

type directHandler struct{}

func (s directHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_DIRECT
}

func (s directHandler) GetSignBytes(_ context.Context, _ signing.SignerData, _ signing.TxData) ([]byte, error) {
	panic("not implemented")
}

func TestAsyncExec(t *testing.T) {
	t.Parallel()
	f := initFixture(t, nil)

	addrs := simtestutil.CreateIncrementalAccounts(2)
	coins := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10)))

	assert.NilError(t, testutil.FundAccount(f.ctx, f.bankKeeper, addrs[0], sdk.NewCoins(sdk.NewInt64Coin("stake", 500))))

	msg := &banktypes.MsgSend{
		FromAddress: addrs[0].String(),
		ToAddress:   addrs[1].String(),
		Amount:      coins,
	}
	msg2 := &banktypes.MsgSend{
		FromAddress: addrs[1].String(),
		ToAddress:   addrs[0].String(),
		Amount:      coins,
	}
	failingMsg := &banktypes.MsgSend{
		FromAddress: addrs[0].String(),
		ToAddress:   addrs[1].String(),
		Amount:      sdk.NewCoins(sdk.NewCoin("stake", sdkmath.ZeroInt())), // No amount specified
	}

	msgAny, err := codectypes.NewAnyWithValue(msg)
	assert.NilError(t, err)

	msgAny2, err := codectypes.NewAnyWithValue(msg2)
	assert.NilError(t, err)

	failingMsgAny, err := codectypes.NewAnyWithValue(failingMsg)
	assert.NilError(t, err)

	testCases := []struct {
		name      string
		req       *authtypes.MsgNonAtomicExec
		expectErr bool
		expErrMsg string
	}{
		{
			name: "empty signer address",
			req: &authtypes.MsgNonAtomicExec{
				Signer: "",
				Msgs:   []*codectypes.Any{},
			},
			expectErr: true,
			expErrMsg: "empty signer address string is not allowed",
		},
		{
			name: "invalid signer address",
			req: &authtypes.MsgNonAtomicExec{
				Signer: "invalid",
				Msgs:   []*codectypes.Any{},
			},
			expectErr: true,
			expErrMsg: "invalid signer address",
		},
		{
			name: "empty msgs",
			req: &authtypes.MsgNonAtomicExec{
				Signer: addrs[0].String(),
				Msgs:   []*codectypes.Any{},
			},
			expectErr: true,
			expErrMsg: "messages cannot be empty",
		},
		{
			name: "valid msg",
			req: &authtypes.MsgNonAtomicExec{
				Signer: addrs[0].String(),
				Msgs:   []*codectypes.Any{msgAny},
			},
			expectErr: false,
		},
		{
			name: "multiple messages being executed",
			req: &authtypes.MsgNonAtomicExec{
				Signer: addrs[0].String(),
				Msgs:   []*codectypes.Any{msgAny, msgAny},
			},
			expectErr: false,
		},
		{
			name: "multiple messages with different signers",
			req: &authtypes.MsgNonAtomicExec{
				Signer: addrs[0].String(),
				Msgs:   []*codectypes.Any{msgAny, msgAny2},
			},
			expectErr: false,
			expErrMsg: "unauthorized: sender does not match expected sender",
		},
		{
			name: "multi msg with one failing being executed",
			req: &authtypes.MsgNonAtomicExec{
				Signer: addrs[0].String(),
				Msgs:   []*codectypes.Any{msgAny, failingMsgAny},
			},
			expectErr: false,
			expErrMsg: "invalid coins",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.req,
				integration.WithAutomaticFinalizeBlock(),
				integration.WithAutomaticCommit(),
			)
			if tc.expectErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := authtypes.MsgNonAtomicExecResponse{}
				err = f.cdc.Unmarshal(res.Value, &result)
				assert.NilError(t, err)

				if tc.expErrMsg != "" {
					for _, res := range result.Results {
						if res.Error != "" {
							assert.Assert(t, strings.Contains(res.Error, tc.expErrMsg), fmt.Sprintf("res.Error %s does not contain %s", res.Error, tc.expErrMsg))
						}
						continue
					}
				}
			}
		})
	}
}
