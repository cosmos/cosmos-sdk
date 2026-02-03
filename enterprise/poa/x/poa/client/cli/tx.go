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
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	"github.com/cosmos/cosmos-sdk/version"
)

// NewTxCommand returns the root transaction command for the POA module.
func NewTxCommand(pubkeyFactory map[string]func(codec.Codec, []byte) *codectypes.Any) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "PoA transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewUpdateParamsCmd(),
		NewCreateValidatorCmd(pubkeyFactory),
		NewUpdateValidatorsCmd(),
		NewWithdrawFeesCmd(),
	)

	return txCmd
}

// NewUpdateParamsCmd returns a CLI command for updating POA module parameters.
func NewUpdateParamsCmd() *cobra.Command {
	type paramsFlags struct {
		Admin string
	}
	// static assertion: if Params and paramsFlags ever diverge, this file will not compile.
	// DEVELOPERS: If this is not compiling, you need to check the flags below and either add or remove flags so they match
	// the types.Params fields.
	_ = func(p types.Params) {
		_ = types.Params(paramsFlags{})
		_ = paramsFlags(p)
	}

	pflags := paramsFlags{}
	flagAdmin := "admin"

	cmd := &cobra.Command{
		Use:     "update-params",
		Short:   "Updates the parameters of the poa module.",
		Long:    "Updates the parameters of the poa module. You only need to pass in the flags of the values you wish to change. All other parameters will remain unchanged.",
		Example: fmt.Sprintf("%s update-params --admin cosmos1...", version.AppName),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}
			p := res.Params

			anyChanged := false
			if cmd.Flags().Changed(flagAdmin) {
				p.Admin = pflags.Admin
				anyChanged = true
			}

			if !anyChanged {
				cmd.Println("No changes made. Either no flags were passed, or the values were the same as the module's values.")
				return nil
			}
			msg := &types.MsgUpdateParams{
				Params: p,
				Admin:  clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().StringVar(&pflags.Admin, flagAdmin, "", "admin address of the PoA module")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewUpdateValidatorsCmd returns a CLI command for updating POA validators.
func NewUpdateValidatorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-validators",
		Short:   "Updates the validators of the poa module.",
		Long:    "Updates the validators of the poa module.",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("%s update-validators validators_file.json", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			valFile, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read validators_file: %w", err)
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var rawVals []json.RawMessage
			if err := json.Unmarshal(valFile, &rawVals); err != nil {
				return fmt.Errorf("failed to unmarshal validators_file as array: %w", err)
			}

			validators := make([]types.Validator, 0, len(rawVals))

			for i, raw := range rawVals {
				var v types.Validator
				if err := clientCtx.Codec.UnmarshalJSON(raw, &v); err != nil {
					return fmt.Errorf("failed to unmarshal validator at index %d: %w", i, err)
				}
				validators = append(validators, v)
			}

			signerAddr := clientCtx.GetFromAddress()

			msg := &types.MsgUpdateValidators{
				Validators: validators,
				Admin:      signerAddr.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCreateValidatorCmd(pubkeyFactory map[string]func(codec.Codec, []byte) *codectypes.Any) *cobra.Command {
	const flagDescription = "description"
	cmd := &cobra.Command{
		Use:   "create-validator",
		Short: "Create a new validator",
		Long:  "Create a new validator with the specified moniker, pubkey info, and optional validator description. The operator_address will be set to the signer of the tx.",
		Args:  cobra.ExactArgs(3),
		Example: fmt.Sprintf(
			"%s tx poa create-validator moniker pubkey_base64 pubkey_type --description \"My validator\" ",
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			desc, err := cmd.Flags().GetString(flagDescription)
			if err != nil {
				return err
			}

			moniker, pubkeyHex, pubkeyType := args[0], args[1], args[2]

			pubkeyBz, err := base64.StdEncoding.DecodeString(pubkeyHex)
			if err != nil {
				return fmt.Errorf("failed to decode pubkey_hex: %w", err)
			}

			pubkeyType = strings.ToLower(strings.TrimSpace(pubkeyType))
			var pkAny *codectypes.Any
			if factory, ok := pubkeyFactory[pubkeyType]; ok {
				pkAny = factory(clientCtx.Codec, pubkeyBz)
			} else {
				return fmt.Errorf("unknown pubkey type: %s. if your application has custom pubkey types, "+
					"use the WithPubkeyFactory option when constructing the PoA module", pubkeyType)
			}

			msg := &types.MsgCreateValidator{
				PubKey:          pkAny,
				Moniker:         moniker,
				Description:     desc,
				OperatorAddress: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagDescription, "", "validator description")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewWithdrawFeesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "withdraw-fees",
		Short:   "Withdraw allocated fees for a validator",
		Long:    "Withdraw allocated fees for a validator operator",
		Example: fmt.Sprintf("%s tx poa withdraw-fees", version.AppName),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgWithdrawFees{
				Operator: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
