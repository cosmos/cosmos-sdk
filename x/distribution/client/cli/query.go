package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
)

// GetCmdQueryParams implements the query params command.
func GetCmdQueryParams(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.ExactArgs(0),
		Short: "Query distribution params",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := queryParams(cliCtx, cdc, queryRoute)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil

		},
	}
	return cmd
}

func queryParams(cliCtx context.CLIContext, cdc *codec.Codec, queryRoute string) ([]byte, error) {
	retCommunityTax, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/params/community_tax", queryRoute), []byte{})
	if err != nil {
		return nil, err
	}

	retBaseProposerReward, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/params/base_proposer_reward", queryRoute), []byte{})
	if err != nil {
		return nil, err
	}

	retBonusProposerReward, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/params/bonus_proposer_reward", queryRoute), []byte{})
	if err != nil {
		return nil, err
	}

	retWithdrawAddrEnabled, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/params/withdraw_addr_enabled", queryRoute), []byte{})
	if err != nil {
		return nil, err
	}

	return codec.MarshalJSONIndent(cdc, NewPrettyParams(retCommunityTax, retBaseProposerReward, retBonusProposerReward, retWithdrawAddrEnabled))
}

// GetCmdQueryOutstandingRewards implements the query outstanding rewards command.
func GetCmdQueryOutstandingRewards(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outstanding-rewards",
		Args:  cobra.ExactArgs(0),
		Short: "Query distribution outstanding (un-withdrawn) rewards",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := queryOutstandingRewards(cliCtx, cdc, queryRoute)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}
	return cmd
}

func queryOutstandingRewards(cliCtx context.CLIContext, cdc *codec.Codec, queryRoute string) ([]byte, error) {
	return cliCtx.QueryWithData(fmt.Sprintf("custom/%s/outstanding_rewards", queryRoute), []byte{})
}

// GetCmdQueryValidatorCommission implements the query validator commission command.
func GetCmdQueryValidatorCommission(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commission [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query distribution validator commission",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := queryValidatorCommission(cliCtx, cdc, queryRoute, distr.NewQueryValidatorCommissionParams(validatorAddr))
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}
	return cmd
}

func queryValidatorCommission(cliCtx context.CLIContext, cdc *codec.Codec, queryRoute string, params distr.QueryValidatorCommissionParams) ([]byte, error) {
	bz, err := cdc.MarshalJSON(params)
	if err != nil {
		return nil, err
	}
	return cliCtx.QueryWithData(fmt.Sprintf("custom/%s/validator_commission", queryRoute), bz)
}

// GetCmdQueryValidatorSlashes implements the query validator slashes command.
func GetCmdQueryValidatorSlashes(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
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

			res, err := queryValidatorSlashes(cliCtx, cdc, queryRoute, distr.NewQueryValidatorSlashesParams(validatorAddr, startHeight, endHeight))
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}
	return cmd
}

func queryValidatorSlashes(cliCtx context.CLIContext, cdc *codec.Codec, queryRoute string, params distr.QueryValidatorSlashesParams) ([]byte, error) {
	bz, err := cdc.MarshalJSON(params)
	if err != nil {
		return nil, err
	}
	return cliCtx.QueryWithData(fmt.Sprintf("custom/%s/validator_slashes", queryRoute), bz)
}

// GetCmdQueryDelegatorRewards implements the query delegator rewards command.
func GetCmdQueryDelegatorRewards(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
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

			res, err := queryDelegationRewards(cliCtx, cdc, queryRoute, distr.NewQueryDelegationRewardsParams(delegatorAddr, validatorAddr))
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}
	return cmd
}

func queryDelegationRewards(cliCtx context.CLIContext, cdc *codec.Codec, queryRoute string, params distr.QueryDelegationRewardsParams) ([]byte, error) {
	bz, err := cdc.MarshalJSON(params)
	if err != nil {
		return nil, err
	}
	return cliCtx.QueryWithData(fmt.Sprintf("custom/%s/delegation_rewards", queryRoute), bz)
}
