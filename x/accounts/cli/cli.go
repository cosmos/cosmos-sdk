package cli

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/cosmos/gogoproto/jsonpb"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"

	v1 "cosmossdk.io/x/accounts/v1"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

func TxCmd(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                name,
		Short:              "Transactions command for the " + name + " module",
		RunE:               client.ValidateCmd,
		DisableFlagParsing: true,
	}
	cmd.AddCommand(GetTxInitCmd(), GetExecuteCmd())
	return cmd
}

func GetTxInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <account-type> <json-message>",
		Short: "Initialize a new account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sender, err := clientCtx.AddressCodec.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			msg := v1.MsgInit{
				Sender:      sender,
				AccountType: args[0],
				JsonMessage: args[1],
			}

			isGenesis, err := cmd.Flags().GetBool("genesis")
			if err != nil {
				return err
			}

			// in case the genesis flag is provided then the init message is printed.
			if isGenesis {
				return clientCtx.WithOutputFormat(flags.OutputFormatJSON).PrintProto(&msg)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	cmd.Flags().Bool("genesis", false, "if true will print the json init message for genesis")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetExecuteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute <account-address> <execute-msg-type-url> <json-message>",
		Short: "Execute state transition to account",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sender, err := clientCtx.AddressCodec.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			schema, err := getSchemaForAccount(clientCtx, args[0])
			if err != nil {
				return err
			}

			msgBytes, err := handlerMsgBytes(schema.ExecuteHandlers, args[1], args[2])
			if err != nil {
				return err
			}
			msg := v1.MsgExecute{
				Sender:  sender,
				Target:  args[0],
				Message: msgBytes,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func getSchemaForAccount(clientCtx client.Context, addr string) (*v1.SchemaResponse, error) {
	queryClient := v1.NewQueryClient(clientCtx)
	accType, err := queryClient.AccountType(clientCtx.CmdContext, &v1.AccountTypeRequest{
		Address: addr,
	})
	if err != nil {
		return nil, err
	}
	return queryClient.Schema(clientCtx.CmdContext, &v1.SchemaRequest{
		AccountType: accType.AccountType,
	})
}

func handlerMsgBytes(handlersSchema []*v1.SchemaResponse_Handler, msgTypeURL, msgString string) (*codectypes.Any, error) {
	var msgSchema *v1.SchemaResponse_Handler
	for _, handler := range handlersSchema {
		if handler.Request == msgTypeURL {
			msgSchema = handler
			break
		}
	}
	if msgSchema == nil {
		return nil, fmt.Errorf("handler for message type %s not found", msgTypeURL)
	}
	return encodeJSONToProto(msgSchema.Request, msgString)
}

func encodeJSONToProto(name, jsonMsg string) (*codectypes.Any, error) {
	impl := gogoproto.MessageType(name)
	if impl == nil {
		return nil, fmt.Errorf("message type %s not found", name)
	}
	msg := reflect.New(impl.Elem()).Interface().(gogoproto.Message)
	err := jsonpb.Unmarshal(bytes.NewBufferString(jsonMsg), msg)
	if err != nil {
		return nil, fmt.Errorf("provided message is not valid %s: %w", jsonMsg, err)
	}
	return codectypes.NewAnyWithValue(msg)
}
