package accounts

import (
	"context"
	"fmt"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	account_abstractionv1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	banktypes "cosmossdk.io/x/bank/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

func TestMsgServer_ExecuteBundle(t *testing.T) {
	t.Run("bundle success", func(t *testing.T) {
		f := initFixture(t, func(ctx context.Context, msg *account_abstractionv1.MsgAuthenticate) (*account_abstractionv1.MsgAuthenticateResponse, error) {
			return &account_abstractionv1.MsgAuthenticateResponse{}, nil
		})

		recipient := f.mustAddr([]byte("recipient"))
		feeAmt := sdk.NewInt64Coin("atom", 100)
		sendAmt := sdk.NewInt64Coin("atom", 200)

		f.mint(f.mockAccountAddress, feeAmt, sendAmt)

		tx := makeTx(t, &banktypes.MsgSend{
			FromAddress: f.mustAddr(f.mockAccountAddress),
			ToAddress:   recipient,
			Amount:      sdk.NewCoins(sendAmt),
		}, []byte("pass"), &account_abstractionv1.TxExtension{
			AuthenticationGasLimit: 2400,
			BundlerPaymentMessages: []*codectypes.Any{wrapAny(t, &banktypes.MsgSend{
				FromAddress: f.mustAddr(f.mockAccountAddress),
				ToAddress:   f.bundler,
				Amount:      sdk.NewCoins(feeAmt),
			})},
			BundlerPaymentGasLimit: 30000,
			ExecutionGasLimit:      30000,
		})

		bundleResp := f.runBundle(tx)
		require.Len(t, bundleResp.Responses, 1)

		txResp := bundleResp.Responses[0]

		require.Empty(t, txResp.Error)
		require.NotZero(t, txResp.AuthenticationGasUsed)
		require.NotZero(t, txResp.BundlerPaymentGasUsed)
		require.NotZero(t, txResp.ExecutionGasUsed)

		// asses responses
		require.Len(t, txResp.BundlerPaymentResponses, 1)
		require.Equal(t, txResp.BundlerPaymentResponses[0].TypeUrl, "/cosmos.bank.v1beta1.MsgSendResponse")

		require.Len(t, txResp.ExecutionResponses, 1)
		require.Equal(t, txResp.ExecutionResponses[0].TypeUrl, "/cosmos.bank.v1beta1.MsgSendResponse")

		// ensure sends have happened
		require.Equal(t, f.balance(f.bundler, feeAmt.Denom), feeAmt)
		require.Equal(t, f.balance(recipient, sendAmt.Denom), sendAmt)
	})

	t.Run("tx fails at auth step", func(t *testing.T) {
		f := initFixture(t, func(ctx context.Context, msg *account_abstractionv1.MsgAuthenticate) (*account_abstractionv1.MsgAuthenticateResponse, error) {
			return &account_abstractionv1.MsgAuthenticateResponse{}, fmt.Errorf("sentinel")
		})
		recipient := f.mustAddr([]byte("recipient"))
		feeAmt := sdk.NewInt64Coin("atom", 100)
		sendAmt := sdk.NewInt64Coin("atom", 200)
		f.mint(f.mockAccountAddress, feeAmt, sendAmt)

		tx := makeTx(t, &banktypes.MsgSend{
			FromAddress: f.mustAddr(f.mockAccountAddress),
			ToAddress:   recipient,
			Amount:      sdk.NewCoins(sendAmt),
		}, []byte("pass"), &account_abstractionv1.TxExtension{
			AuthenticationGasLimit: 2400,
			BundlerPaymentMessages: []*codectypes.Any{wrapAny(t, &banktypes.MsgSend{
				FromAddress: f.mustAddr(f.mockAccountAddress),
				ToAddress:   f.bundler,
				Amount:      sdk.NewCoins(feeAmt),
			})},
			BundlerPaymentGasLimit: 30000,
			ExecutionGasLimit:      30000,
		})

		bundleResp := f.runBundle(tx)

		require.Len(t, bundleResp.Responses, 1)

		txResp := bundleResp.Responses[0]
		require.NotEmpty(t, txResp.Error)
		require.Contains(t, txResp.Error, "sentinel")
		require.NotZero(t, txResp.AuthenticationGasUsed)
		require.Zero(t, txResp.BundlerPaymentGasUsed)
		require.Zero(t, txResp.ExecutionGasUsed)
		require.Empty(t, txResp.BundlerPaymentResponses)
		require.Empty(t, txResp.ExecutionResponses)

		// ensure auth side effects are not persisted in case of failures
	})

	t.Run("tx fails at pay bundler step", func(t *testing.T) {
		f := initFixture(t, func(ctx context.Context, msg *account_abstractionv1.MsgAuthenticate) (*account_abstractionv1.MsgAuthenticateResponse, error) {
			return &account_abstractionv1.MsgAuthenticateResponse{}, nil
		})

		recipient := f.mustAddr([]byte("recipient"))
		feeAmt := sdk.NewInt64Coin("atom", 100)
		sendAmt := sdk.NewInt64Coin("atom", 200)

		f.mint(f.mockAccountAddress, feeAmt, sendAmt)

		tx := makeTx(t, &banktypes.MsgSend{
			FromAddress: f.mustAddr(f.mockAccountAddress),
			ToAddress:   recipient,
			Amount:      sdk.NewCoins(sendAmt),
		}, []byte("pass"), &account_abstractionv1.TxExtension{
			AuthenticationGasLimit: 2400,
			BundlerPaymentMessages: []*codectypes.Any{
				wrapAny(t, &banktypes.MsgSend{
					FromAddress: f.mustAddr(f.mockAccountAddress),
					ToAddress:   f.bundler,
					Amount:      sdk.NewCoins(feeAmt.AddAmount(feeAmt.Amount.AddRaw(100))),
				}),
				wrapAny(t, &banktypes.MsgSend{
					FromAddress: f.mustAddr(f.mockAccountAddress),
					ToAddress:   f.bundler,
					Amount:      sdk.NewCoins(feeAmt.AddAmount(feeAmt.Amount.AddRaw(30000))),
				}),
			},
			BundlerPaymentGasLimit: 30000,
			ExecutionGasLimit:      30000,
		})

		bundleResp := f.runBundle(tx)
		require.Len(t, bundleResp.Responses, 1)

		txResp := bundleResp.Responses[0]

		require.NotEmpty(t, txResp.Error)
		require.Contains(t, txResp.Error, "bundler payment failed")
		require.NotZero(t, txResp.AuthenticationGasUsed)
		require.NotZero(t, txResp.BundlerPaymentGasUsed)

		require.Empty(t, txResp.BundlerPaymentResponses)
		require.Zero(t, txResp.ExecutionGasUsed)
		require.Empty(t, txResp.ExecutionResponses)

		// ensure bundler payment side effects are not persisted
		require.True(t, f.balance(f.bundler, feeAmt.Denom).IsZero())
	})

	t.Run("tx fails at execution step", func(t *testing.T) {
		f := initFixture(t, func(ctx context.Context, msg *account_abstractionv1.MsgAuthenticate) (*account_abstractionv1.MsgAuthenticateResponse, error) {
			return &account_abstractionv1.MsgAuthenticateResponse{}, nil
		})

		recipient := f.mustAddr([]byte("recipient"))
		feeAmt := sdk.NewInt64Coin("atom", 100)
		sendAmt := sdk.NewInt64Coin("atom", 40000) // this fails

		f.mint(f.mockAccountAddress, feeAmt)

		tx := makeTx(t, &banktypes.MsgSend{
			FromAddress: f.mustAddr(f.mockAccountAddress),
			ToAddress:   recipient,
			Amount:      sdk.NewCoins(sendAmt),
		}, []byte("pass"), &account_abstractionv1.TxExtension{
			AuthenticationGasLimit: 2400,
			BundlerPaymentMessages: []*codectypes.Any{
				wrapAny(t, &banktypes.MsgSend{
					FromAddress: f.mustAddr(f.mockAccountAddress),
					ToAddress:   f.bundler,
					Amount:      sdk.NewCoins(feeAmt),
				}),
			},
			BundlerPaymentGasLimit: 30000,
			ExecutionGasLimit:      30000,
		})

		bundleResp := f.runBundle(tx)
		require.Len(t, bundleResp.Responses, 1)

		txResp := bundleResp.Responses[0]

		require.NotEmpty(t, txResp.Error)
		require.Contains(t, txResp.Error, "execution failed")

		require.NotZero(t, txResp.AuthenticationGasUsed)

		require.NotZero(t, txResp.BundlerPaymentGasUsed)
		require.NotEmpty(t, txResp.BundlerPaymentResponses)
		require.Equal(t, f.balance(f.bundler, feeAmt.Denom), feeAmt) // ensure bundler payment side effects are persisted

		require.NotZero(t, txResp.ExecutionGasUsed)
		require.Empty(t, txResp.ExecutionResponses)

		// ensure execution side effects are not persisted
		// aka recipient must not have money
		require.True(t, f.balance(recipient, feeAmt.Denom).IsZero())
	})
}

func makeTx(t *testing.T, msg gogoproto.Message, sig []byte, xt *account_abstractionv1.TxExtension) []byte {
	t.Helper()
	anyMsg, err := codectypes.NewAnyWithValue(msg)
	require.NoError(t, err)

	anyXt, err := codectypes.NewAnyWithValue(xt)
	require.NoError(t, err)

	tx := &txtypes.Tx{
		Body: &txtypes.TxBody{
			Messages:                    []*codectypes.Any{anyMsg},
			Memo:                        "",
			TimeoutHeight:               0,
			Unordered:                   false,
			TimeoutTimestamp:            nil,
			ExtensionOptions:            []*codectypes.Any{anyXt},
			NonCriticalExtensionOptions: nil,
		},
		AuthInfo: &txtypes.AuthInfo{
			SignerInfos: []*txtypes.SignerInfo{
				{
					PublicKey: nil,
					ModeInfo:  &txtypes.ModeInfo{Sum: &txtypes.ModeInfo_Single_{Single: &txtypes.ModeInfo_Single{Mode: signingtypes.SignMode_SIGN_MODE_UNSPECIFIED}}},
					Sequence:  0,
				},
			},
			Fee: nil,
		},
		Signatures: [][]byte{sig},
	}

	bodyBytes, err := tx.Body.Marshal()
	require.NoError(t, err)

	authInfoBytes, err := tx.AuthInfo.Marshal()
	require.NoError(t, err)

	txRaw, err := (&txtypes.TxRaw{
		BodyBytes:     bodyBytes,
		AuthInfoBytes: authInfoBytes,
		Signatures:    tx.Signatures,
	}).Marshal()
	require.NoError(t, err)
	return txRaw
}

func wrapAny(t *testing.T, msg gogoproto.Message) *codectypes.Any {
	t.Helper()
	any, err := codectypes.NewAnyWithValue(msg)
	require.NoError(t, err)
	return any
}
