package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
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

	fmt.Println("txBz", string(txBz))

	req := &txtypes.BroadcastTxRequest{
		TxBytes: txBz,
		Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
	}

	serviceClient := txtypes.NewServiceClient(clientCtx)
	resp, err := serviceClient.BroadcastTx(context.Background(), req)

	fmt.Println("Resp", resp, err)

	// return resp, err
	return nil, nil
}
