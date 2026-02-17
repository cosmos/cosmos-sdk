// IMPORTANT LICENSE NOTICE
//
// SPDX-License-Identifier: CosmosLabs-Evaluation-Only
//
// This file is NOT licensed under the Apache License 2.0.
//
// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:
// - commercial use,
// - production use, and
// - redistribution.
//
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/group/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var (
	flagNumValidators = "validator-count"
	flagOutputDir     = "output-dir"
)

// NewTestnetCmd creates a root testnet command with subcommands.
// Note: The group simapp uses auth, bank, consensus, genutil, and group modules
// without staking. The testnet init-files creates genesis with auth and bank
// accounts. For a chain that produces blocks, staking would need to be added.
func NewTestnetCmd(mm module.BasicManager, genBalIterator banktypes.GenesisBalancesIterator) *cobra.Command {
	testnetCmd := &cobra.Command{
		Use:                        "testnet",
		Short:                      "subcommands for starting or configuring local testnets",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	testnetCmd.AddCommand(testnetStartCmd())
	testnetCmd.AddCommand(testnetInitFilesCmd(mm, genBalIterator))

	return testnetCmd
}

func addTestnetFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().IntP(flagNumValidators, "v", 4, "Number of validators to initialize the testnet with")
	cmd.Flags().StringP(flagOutputDir, "o", "./.testnets", "Directory to store initialization data for the testnet")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(flags.FlagKeyType, string(hd.Secp256k1Type), "Key signing algorithm to generate keys for")

	cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		if name == flags.FlagKeyAlgorithm {
			name = flags.FlagKeyType
		}
		return pflag.NormalizedName(name)
	})
}

func testnetInitFilesCmd(mm module.BasicManager, genBalIterator banktypes.GenesisBalancesIterator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-files",
		Short: "Initialize config directories & files for a multi-validator testnet",
		Long: `Initialize config directories & files for a multi-validator testnet.
Note: The group simapp does not include staking. For full testnet support
with block production, consider adding the staking module.`,
		RunE: client.ValidateCmd,
	}

	addTestnetFlagsToCmd(cmd)
	return cmd
}

func testnetStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Launch an in-process multi-validator testnet",
		Long:  `Launch an in-process multi-validator testnet.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("testnet start requires staking module; add staking to the group simapp for full testnet support")
		},
	}

	addTestnetFlagsToCmd(cmd)
	return cmd
}
