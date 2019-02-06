package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
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
			params, err := common.QueryParams(cliCtx, queryRoute)
			if err != nil {
				return err
			}
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
		Long: strings.TrimSpace(`Query validator commission rewards from delegators to that validator:

$ gaiacli query distr commission cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := common.QueryValidatorCommission(cliCtx, cdc, queryRoute, validatorAddr)
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
		Long: strings.TrimSpace(`Query all slashes of a validator for a given block range:

$ gaiacli query distr slashes cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 0 100
`),
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
		Use:   "rewards [delegator-addr] [<validator-addr>]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Query all distribution delegator rewards or rewards from a particular validator",
		Long: strings.TrimSpace(`Query all rewards earned by a delegator, optionally restrict to rewards from a single validator:

$ gaiacli query distr rewards cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
$ gaiacli query distr rewards cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var resp []byte
			var err error
			if len(args) == 2 {
				// query for rewards from a particular delegation
				resp, err = common.QueryDelegationRewards(cliCtx, cdc, queryRoute, args[0], args[1])
			} else {
				// query for delegator total rewards
				resp, err = common.QueryDelegatorTotalRewards(cliCtx, cdc, queryRoute, args[0])
			}
			if err != nil {
				return err
			}

			var result sdk.DecCoins
			cdc.MustUnmarshalJSON(resp, &result)
			return cliCtx.PrintOutput(result)
		},
	}
}
