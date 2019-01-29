package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// GetCmdQueryParams implements the query params command.
func GetCmdQueryParams(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query distribution params",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/params/community_tax", queryRoute)
			retCommunityTax, err := cliCtx.QueryWithData(route, []byte{})
			if err != nil {
				return err
			}

			route = fmt.Sprintf("custom/%s/params/base_proposer_reward", queryRoute)
			retBaseProposerReward, err := cliCtx.QueryWithData(route, []byte{})
			if err != nil {
				return err
			}

			route = fmt.Sprintf("custom/%s/params/bonus_proposer_reward", queryRoute)
			retBonusProposerReward, err := cliCtx.QueryWithData(route, []byte{})
			if err != nil {
				return err
			}

			route = fmt.Sprintf("custom/%s/params/withdraw_addr_enabled", queryRoute)
			retWithdrawAddrEnabled, err := cliCtx.QueryWithData(route, []byte{})
			if err != nil {
				return err
			}

			params := NewPrettyParams(retCommunityTax, retBaseProposerReward,
				retBonusProposerReward, retWithdrawAddrEnabled)

			return cliCtx.PrintOutput(params)
		},
	}
}

// GetCmdQueryOutstandingRewards implements the query outstanding rewards command.
func GetCmdQueryOutstandingRewards(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "outstanding-rewards",
		Args:  cobra.NoArgs,
		Short: "Query distribution outstanding (un-withdrawn) rewards",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/outstanding_rewards", queryRoute)
			res, err := cliCtx.QueryWithData(route, []byte{})
			if err != nil {
				return err
			}

			var outstandingRewards types.OutstandingRewards
			cdc.MustUnmarshalJSON(res, &outstandingRewards)
			return cliCtx.PrintOutput(outstandingRewards)
		},
	}
}

// GetCmdQueryValidatorCommission implements the query validator commission command.
func GetCmdQueryValidatorCommission(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "commission [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query distribution validator commission",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(distr.NewQueryValidatorCommissionParams(validatorAddr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/validator_commission", queryRoute)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var valCom types.ValidatorAccumulatedCommission
			cdc.MustUnmarshalJSON(res, &valCom)
			return cliCtx.PrintOutput(valCom)
		},
	}
}

// GetCmdQueryValidatorSlashes implements the query validator slashes command.
func GetCmdQueryValidatorSlashes(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "slashes [validator] [start-height] [end-height]",
		Args:  cobra.ExactArgs(3),
		Short: "Query distribution validator slashes",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			startHeight, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("start-height %s not a valid uint, please input a valid start-height", args[1])
			}

			endHeight, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("end-height %s not a valid uint, please input a valid end-height", args[2])
			}

			params := distr.NewQueryValidatorSlashesParams(validatorAddr, startHeight, endHeight)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/validator_slashes", queryRoute), bz)
			if err != nil {
				return err
			}

			var slashes types.ValidatorSlashEvents
			cdc.MustUnmarshalJSON(res, &slashes)
			return cliCtx.PrintOutput(slashes)
		},
	}
}

// GetCmdQueryDelegatorRewards implements the query delegator rewards command.
func GetCmdQueryDelegatorRewards(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "rewards [delegator] [validator]",
		Args:  cobra.ExactArgs(2),
		Short: "Query distribution delegator rewards",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			validatorAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			params := distr.NewQueryDelegationRewardsParams(delegatorAddr, validatorAddr)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/delegation_rewards", queryRoute)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var coins sdk.DecCoins
			cdc.MustUnmarshalJSON(res, &coins)
			return cliCtx.PrintOutput(coins)
		},
	}
}
