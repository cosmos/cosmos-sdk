# Update to Cosmos-SDK v0.40

The following document describes the changes, and an approach to update the chain to use Cosmos-SDK v0.40 
aka stargate release.

## Contents
- [Updating a module to use Cosmos-SDK v0.40]()
- [Updating a cosmos-sdk chain to use SDK v0.40]()
- [Upgrading a live chain to v0.40]()

## Updating a module to use Cosmos-SDK v0.40
This section covers the important changes in modules from `v0.39` to `v0.40`
#### client
- `context.CLIContext` is renamed to `client.Context` and moved to `github.com/cosmos/cosmos-sdk/client`
- The global `viper` usage is removed from client and is replaced with Cobra' `cmd.Flags()`. There are two helpers 
to read common flags for CLI txs and queries. 
```go
clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
```
- Flags helper functions `flags.PostCommands(cmds ...*cobra.Command) []*cobra.Command`, `flags.GetCommands(...)` usage
 is now replaced by `flags.AddTxFlagsToCmd(cmd *cobra.Command)` and `flags.AddQueryFlagsToCmd(cmd *cobra.Command)` 
 respectively. 
- New CLI tx commands doesn't take `codec` as an input now. 
```go
// v0.39.x
func SendTxCmd(cdc *codec.Codec) *cobra.Command {
	...
}

// v0.40
func NewSendTxCmd() *cobra.Command {
	...
}
```
The new definition for CLI tx command would look like the following
```go
// NewSendTxCmd returns a CLI command handler for creating a MsgSend transaction.
func NewSendTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "send [from_key_or_address] [to_address] [amount]",
		Short: `Send funds from one account to another. Note, the'--from' flag is
ignored as it is implied from [from_key_or_address].`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])

			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			toAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoins(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgSend(clientCtx.GetFromAddress(), toAddr, coins)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
```