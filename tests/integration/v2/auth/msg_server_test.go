package auth

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/core/transaction"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/bank/testutil"
	banktypes "cosmossdk.io/x/bank/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestAsyncExec(t *testing.T) {
	t.Parallel()
	s := createTestSuite(t)

	addrs := simtestutil.CreateIncrementalAccounts(2)
	coins := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10)))

	assert.NilError(t, testutil.FundAccount(s.ctx, s.bankKeeper, addrs[0], sdk.NewCoins(sdk.NewInt64Coin("stake", 500))))

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

	msgServer := authkeeper.NewMsgServerImpl(s.authKeeper)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := s.app.RunMsg(
				t,
				s.ctx,
				func(ctx context.Context) (transaction.Msg, error) {
					resp, e := msgServer.NonAtomicExec(ctx, tc.req)
					return resp, e
				},
				integration.WithAutomaticCommit(),
			)
			if tc.expectErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result, ok := res.(*authtypes.MsgNonAtomicExecResponse)
				assert.Assert(t, ok)

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
