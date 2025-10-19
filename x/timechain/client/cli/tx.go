// SPDX-License-Identifier: MPL-2.0
// Copyright © 2025 Timechain-Arweave-LunCoSim Contributors

package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/timechain/types"
)

var (
	DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(30) * time.Minute).Nanoseconds())
)

const (
	flagPacketTimeoutTimestamp = "packet-timeout-timestamp"
	listSeparator              = ","
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdProposeSlot())
	cmd.AddCommand(CmdConfirmSlot())
	cmd.AddCommand(CmdRelayEvent())

	return cmd
}

func CmdProposeSlot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose [slot] [vdf-output] [payload-hash]",
		Short: "Propose a new slot",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// TODO: Implement command
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &types.MsgProposeSlot{})
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdConfirmSlot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm [slot-id] [validator] [sig]",
		Short: "Confirm a slot",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// TODO: Implement command
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &types.MsgConfirmSlot{})
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdRelayEvent() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relay [event] [tss-sig]",
		Short: "Relay a cross-chain event",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// TODO: Implement command
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &types.MsgRelayEvent{})
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
