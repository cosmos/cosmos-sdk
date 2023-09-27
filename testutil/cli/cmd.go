package cli

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// ExecTestCLICmd builds the client context, mocks the output and executes the command.
func ExecTestCLICmd(clientCtx client.Context, cmd *cobra.Command, extraArgs []string) (testutil.BufferWriter, error) {
	cmd.SetArgs(extraArgs)

	_, out := testutil.ApplyMockIO(cmd)
	clientCtx = clientCtx.WithOutput(out)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	if err := cmd.ExecuteContext(ctx); err != nil {
		return out, err
	}

	return out, nil
}

func MsgSendExec(clientCtx client.Context, from, to, amount fmt.Stringer, ac address.Codec, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{from.String(), to.String(), amount.String()}
	args = append(args, extraArgs...)

	coins, err := sdk.ParseCoinsNormalized(amount.String())
	if err != nil {
		return nil, err
	}

	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	msgSend := banktypes.MsgSend{
		FromAddress: from.String(),
		ToAddress:   to.String(),
		Amount:      coins,
	}

	fee := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(0)))
	txBuilder.SetFeeAmount(fee)    // Arbitrary fee
	txBuilder.SetGasLimit(1000000) // Need at least 100386

	err = txBuilder.SetMsgs(&msgSend)
	if err != nil {
		return nil, err
	}

	txFactory := tx.Factory{}
	txFactory = txFactory.
		WithChainID(clientCtx.ChainID).
		WithKeybase(clientCtx.Keyring).
		WithTxConfig(clientCtx.TxConfig)

	accBytes, err := clientCtx.AddressCodec.StringToBytes(from.String())
	if err != nil {
		return nil, err
	}

	keyRecord, err := clientCtx.Keyring.KeyByAddress(accBytes)
	if err != nil {
		return nil, err
	}

	err = tx.Sign(context.Background(), txFactory, keyRecord.Name, txBuilder, true)
	if err != nil {
		return nil, err
	}

	txBz, err := clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	req := &txtypes.BroadcastTxRequest{
		TxBytes: txBz,
		Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
	}

	serviceClient := txtypes.NewServiceClient(clientCtx)
	resp, err := serviceClient.BroadcastTx(context.Background(), req)

	fmt.Println("Resp", resp, err, resp.TxResponse.Code)

	// return resp, err
	return nil, nil
}

func GenOrBroadcastTestTx(clientCtx client.Context, msg proto.Message, from sdk.AccAddress, generateonly bool) (testutil.BufferWriter, error) {
	txBuilder := clientCtx.TxConfig.NewTxBuilder()

	fee := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10)))
	txBuilder.SetFeeAmount(fee)    // Arbitrary fee
	txBuilder.SetGasLimit(1000000) // Need at least 100386

	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, err
	}

	if generateonly {
		txBz, err := clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
		if err != nil {
			return nil, err
		}

		out := bytes.NewBuffer(txBz)
		return out, nil
	}

	txFactory := tx.Factory{}
	txFactory = txFactory.
		WithChainID(clientCtx.ChainID).
		WithKeybase(clientCtx.Keyring).
		WithTxConfig(clientCtx.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	accBytes, err := clientCtx.AddressCodec.StringToBytes(from.String())
	if err != nil {
		return nil, err
	}

	keyRecord, err := clientCtx.Keyring.KeyByAddress(accBytes)
	if err != nil {
		return nil, err
	}

	err = tx.Sign(context.Background(), txFactory, keyRecord.Name, txBuilder, true)
	if err != nil {
		return nil, err
	}

	txBz, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	req := &txtypes.BroadcastTxRequest{
		TxBytes: txBz,
		Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
	}

	serviceClient := txtypes.NewServiceClient(clientCtx)
	resp, err := serviceClient.BroadcastTx(context.Background(), req)
	if err != nil {
		return nil, err
	}

	respBz, err := clientCtx.Codec.Marshal(resp)
	if err != nil {
		return nil, err
	}

	out := bytes.NewBuffer(respBz)
	return out, nil
}

type SubmitTestTxConfig struct {
	Simulate bool
	GenOnly  bool
	Memo     string
	Gas      uint64
}

func SubmitTestTx(val *network.Validator, msg proto.Message, from sdk.AccAddress, config SubmitTestTxConfig) (testutil.BufferWriter, error) {
	clientCtx := val.ClientCtx
	txBuilder := clientCtx.TxConfig.NewTxBuilder()

	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, err
	}

	fee := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10)))
	txBuilder.SetFeeAmount(fee) // Arbitrary fee

	if config.Gas != 0 {
		txBuilder.SetGasLimit(config.Gas)
	} else {
		txBuilder.SetGasLimit(flags.DefaultGasLimit) // Need at least 100386
	}

	if config.Memo != "" {
		txBuilder.SetMemo(config.Memo)
	}

	if config.GenOnly {
		txBz, err := val.ClientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
		if err != nil {
			return nil, err
		}

		out := bytes.NewBuffer(txBz)
		return out, nil
	}

	acc, err := clientCtx.AccountRetriever.GetAccount(clientCtx, from)
	if err != nil {
		return nil, err
	}

	txFactory := tx.Factory{}
	txFactory = txFactory.
		WithChainID(clientCtx.ChainID).
		WithKeybase(clientCtx.Keyring).
		WithTxConfig(clientCtx.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithAccountNumber(acc.GetAccountNumber()).
		WithSequence(acc.GetSequence())

	if config.Simulate {
		txFactory = txFactory.WithSimulateAndExecute(true)
	}

	accBytes, err := clientCtx.AddressCodec.StringToBytes(from.String())
	if err != nil {
		return nil, err
	}

	keyRecord, err := clientCtx.Keyring.KeyByAddress(accBytes)
	if err != nil {
		return nil, err
	}

	err = tx.Sign(context.Background(), txFactory, keyRecord.Name, txBuilder, true)
	if err != nil {
		return nil, err
	}

	txBz, err := val.ClientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	res, err := val.BroadcastTx(txBz)
	out := bytes.NewBuffer(res)

	return out, err
}
