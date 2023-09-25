package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
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

	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}).Codec
	router := baseapp.NewMsgServiceRouter()

	router.SetInterfaceRegistry(cdc.InterfaceRegistry())

	msgSend := banktypes.MsgSend{
		FromAddress: from.String(),
		ToAddress:   to.String(),
		Amount:      coins,
	}

	// router.RegisterService(router, banktypes.MsgServer{})
	handler := router.Handler(&msgSend)

	fmt.Printf("handler: %v\n", handler)

	return nil, nil
	// return ExecTestCLICmd(clientCtx, cli.NewSendTxCmd(ac), args)
}
