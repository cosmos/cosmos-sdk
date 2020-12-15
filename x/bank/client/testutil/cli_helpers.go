package testutil

import (
	"context"
	"fmt"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
	grpc "google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func MsgSendExec(clientCtx client.Context, from, to, amount fmt.Stringer, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{from.String(), to.String(), amount.String()}
	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, bankcli.NewSendTxCmd(), args)
}

func QueryBalancesExec(clientCtx client.Context, address fmt.Stringer, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{address.String(), fmt.Sprintf("--%s=json", cli.OutputFlag)}
	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, bankcli.GetBalancesCmd(), args)
}

// serviceMsgClientConn is an instance of grpc.ClientConn that is used to test building
// transactions with MsgClient's. It is intended to be replaced by the work in
// https://github.com/cosmos/cosmos-sdk/issues/7541 when that is ready.
type serviceMsgClientConn struct {
	msgs []sdk.Msg
}

func (t *serviceMsgClientConn) Invoke(_ context.Context, method string, args, _ interface{}, _ ...grpc.CallOption) error {
	req, ok := args.(sdk.MsgRequest)
	if !ok {
		return fmt.Errorf("%T should implement %T", args, (*sdk.MsgRequest)(nil))
	}

	err := req.ValidateBasic()
	if err != nil {
		return err
	}

	t.msgs = append(t.msgs, sdk.ServiceMsg{
		MethodName: method,
		Request:    req,
	})

	return nil
}

func (t *serviceMsgClientConn) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("not supported")
}

var _ gogogrpc.ClientConn = &serviceMsgClientConn{}

// newSendTxMsgServiceCmd is just for the purpose of testing ServiceMsg's in an end-to-end case. It is effectively
// NewSendTxCmd but using MsgClient.
func newSendTxMsgServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "send [from_key_or_address] [to_address] [amount]",
		Short: `Send funds from one account to another. Note, the'--from' flag is
ignored as it is implied from [from_key_or_address].`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			toAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgSend(clientCtx.GetFromAddress(), toAddr, coins)
			svcMsgClientConn := &serviceMsgClientConn{}
			bankMsgClient := types.NewMsgClient(svcMsgClientConn)
			_, err = bankMsgClient.Send(context.Background(), msg)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), svcMsgClientConn.msgs...)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// ServiceMsgSendExec is a temporary method to test Msg services in CLI using
// x/bank's Msg/Send service. After https://github.com/cosmos/cosmos-sdk/issues/7541
// is merged, this method should be removed, and we should prefer MsgSendExec
// instead.
func ServiceMsgSendExec(clientCtx client.Context, from, to, amount fmt.Stringer, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{from.String(), to.String(), amount.String()}
	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, newSendTxMsgServiceCmd(), args)
}
