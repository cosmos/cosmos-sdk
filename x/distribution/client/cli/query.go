package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

const (
	ReportDateFormat = "2006-01-02"
)

type ValidatorQuerier interface {
	GetValidatorCommission(height int64, validatorAddress sdk.ValAddress) (sdk.DecCoins, error)
	GetValidatorRewards(height int64, validatorAddress sdk.ValAddress) (sdk.DecCoins, error)
}

type ValidatorQuerierCli struct {
	cliCtx     context.CLIContext
	cdc        *codec.Codec
	queryRoute string
}

func (q ValidatorQuerierCli) GetValidatorCommission(height int64, validatorAddress sdk.ValAddress) (sdk.DecCoins, error) {
	originalHeight := q.cliCtx.Height
	defer func() {
		q.cliCtx.Height = originalHeight
	}()
	q.cliCtx.Height = height
	res, err := common.QueryValidatorCommission(q.cliCtx, q.queryRoute, validatorAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to query validator commission")
	}

	var validatorCommission types.ValidatorAccumulatedCommission
	q.cdc.MustUnmarshalJSON(res, &validatorCommission)
	return validatorCommission, nil
}

func (q ValidatorQuerierCli) GetValidatorRewards(height int64, validatorAddress sdk.ValAddress) (sdk.DecCoins, error) {
	originalHeight := q.cliCtx.Height
	defer func() {
		q.cliCtx.Height = originalHeight
	}()
	delAddr := sdk.AccAddress(validatorAddress).String()
	res, err := common.QueryDelegationRewards(q.cliCtx, q.queryRoute, delAddr, validatorAddress.String())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to query validator reward")
	}

	var validatorRewards sdk.DecCoins
	q.cdc.MustUnmarshalJSON(res, &validatorRewards)
	return validatorRewards, nil
}

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	distQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the distribution module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	distQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryParams(queryRoute, cdc),
		GetCmdQueryValidatorOutstandingRewards(queryRoute, cdc),
		GetCmdQueryValidatorCommission(queryRoute, cdc),
		GetCmdQueryValidatorSlashes(queryRoute, cdc),
		GetCmdQueryDelegatorRewards(queryRoute, cdc),
		GetCmdQueryCommunityPool(queryRoute, cdc),
		GetCmdQueryGenerateReport(queryRoute, cdc),
	)...)

	return distQueryCmd
}

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

// GetCmdQueryValidatorOutstandingRewards implements the query validator outstanding rewards command.
func GetCmdQueryValidatorOutstandingRewards(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "validator-outstanding-rewards [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query distribution outstanding (un-withdrawn) rewards for a validator and all their delegations",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query distribution outstanding (un-withdrawn) rewards
for a validator and all their delegations.

Example:
$ %s query distr validator-outstanding-rewards cosmosvaloper1lwjmdnks33xwnmfayc64ycprww49n33mtm92ne
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			resp, err := common.QueryValidatorOutstandingReward(cliCtx, queryRoute, valAddr)
			if err != nil {
				return err
			}

			var outstandingRewards types.ValidatorOutstandingRewards
			if err := cdc.UnmarshalJSON(resp, &outstandingRewards); err != nil {
				return err
			}

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
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query validator commission rewards from delegators to that validator.

Example:
$ %s query distr commission cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := common.QueryValidatorCommission(cliCtx, queryRoute, validatorAddr)
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
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all slashes of a validator for a given block range.

Example:
$ %s query distr slashes cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 0 100
`,
				version.ClientName,
			),
		),
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

			params := types.NewQueryValidatorSlashesParams(validatorAddr, startHeight, endHeight)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/validator_slashes", queryRoute), bz)
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
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all rewards earned by a delegator, optionally restrict to rewards from a single validator.

Example:
$ %s query distr rewards cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
$ %s query distr rewards cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.ClientName, version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			if len(args) == 2 {
				// query for rewards from a particular delegation
				resp, err := common.QueryDelegationRewards(cliCtx, queryRoute, args[0], args[1])
				if err != nil {
					return err
				}

				var result sdk.DecCoins
				cdc.MustUnmarshalJSON(resp, &result)
				return cliCtx.PrintOutput(result)
			}

			// query for delegator total rewards
			resp, err := common.QueryDelegatorTotalRewards(cliCtx, queryRoute, args[0])
			if err != nil {
				return err
			}

			var result types.QueryDelegatorTotalRewardsResponse
			cdc.MustUnmarshalJSON(resp, &result)
			return cliCtx.PrintOutput(result)
		},
	}
}

// GetCmdQueryCommunityPool returns the command for fetching community pool info
func GetCmdQueryCommunityPool(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "community-pool",
		Args:  cobra.NoArgs,
		Short: "Query the amount of coins in the community pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all coins in the community pool which is under Governance control.

Example:
$ %s query distr community-pool
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/community_pool", queryRoute), nil)
			if err != nil {
				return err
			}

			var result sdk.DecCoins
			cdc.MustUnmarshalJSON(res, &result)
			return cliCtx.PrintOutput(result)
		},
	}
}

// findBlockFromDate finds the block height right after the timestamp
func findBlockHeightFromDate(startTime time.Time, node rpcclient.Client, maxHeight int64) (*int64, error) {
	minHeight := int64(1)
	searchUnix := startTime.Unix()

	for minHeight <= maxHeight {
		currentHeight := minHeight + (maxHeight-minHeight)/2
		previousHeight := currentHeight - 1
		currentBlock, err := node.Block(&currentHeight)
		if err != nil {
			return nil, sdkerrors.Wrap(err, fmt.Sprintf("unable to get block %d from node", currentHeight))
		}
		previousBlock, err := node.Block(&previousHeight)
		if err != nil {
			return nil, sdkerrors.Wrap(err, fmt.Sprintf("unable to get block %d from node", previousHeight))
		}

		currentUnix := currentBlock.BlockMeta.Header.Time.Unix()
		previousUnix := previousBlock.BlockMeta.Header.Time.Unix()

		switch {
		case currentUnix >= searchUnix && previousUnix < searchUnix:
			return &currentHeight, nil
		case currentUnix < searchUnix:
			minHeight = currentHeight + 1
		default:
			maxHeight = currentHeight - 1
		}
	}

	return nil, errors.New("unable to find block height with search time")
}

func calculateValidatorRewardsAndCommissionPerDay(
	node rpcclient.Client,
	validatorAddress sdk.ValAddress,
	querier ValidatorQuerier,
	endDate time.Time,
	startHeight int64,
	lastBlockHeight int64) (aggByDay [][]sdk.DecCoins, days []time.Time, error error) {
	commission := sdk.DecCoins{}
	reward := sdk.DecCoins{}
	var lastCommission sdk.DecCoins
	var lastReward sdk.DecCoins
	var currentDate time.Time
	if startHeight > 0 {
		// We want to start from the previous day's last block, so we can calculate the diff
		// and account the difference for the starting day's sum.
		startHeight--
	}
	currentHeight := startHeight

	for currentHeight <= lastBlockHeight {
		currentBlock, err := node.Block(&currentHeight)
		if err != nil {
			error = sdkerrors.Wrap(err, "unable to get block")
			return
		}

		if currentBlock.BlockMeta.Header.Time.After(endDate) {
			break
		}

		if currentDate.IsZero() {
			currentDate = currentBlock.BlockMeta.Header.Time
		}

		if currentBlock.BlockMeta.Header.Time.Day() != currentDate.Day() {
			// Once we see a block in the next day, record the sum differences
			// for the previous day
			aggByDay = append(aggByDay, []sdk.DecCoins{commission, reward})
			days = append(days, currentDate)
			currentDate = currentBlock.BlockMeta.Header.Time
			commission = sdk.DecCoins{}
			reward = sdk.DecCoins{}
		}

		validatorCommission, err := querier.GetValidatorCommission(currentHeight, validatorAddress)
		if err != nil {
			error = sdkerrors.Wrap(err, "unable to get validator commission")
			return
		}

		if lastCommission != nil {
			commission = commission.Add(validatorCommission.Sub(lastCommission))
		}
		lastCommission = validatorCommission

		validatorRewards, err := querier.GetValidatorRewards(currentHeight, validatorAddress)
		if err != nil {
			error = sdkerrors.Wrap(err, "unable to get validator rewards")
		}

		if lastReward != nil {
			reward = reward.Add(validatorRewards.Sub(lastReward))
		}
		lastReward = validatorRewards
		currentHeight++
	}

	if !commission.IsZero() || !reward.IsZero() {
		aggByDay = append(aggByDay, []sdk.DecCoins{commission, reward})
		days = append(days, currentDate)
	}

	return
}

// GetCmdQueryCommunityPool returns the command for fetching community pool info
func GetCmdQueryGenerateReport(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "gen-report [start-date] [end-date] [addr] [addr]...",
		Args:  cobra.MinimumNArgs(3),
		Short: "Generate a report of rewards",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Generate a report of validator rewards that happened for particular addresses and time span.

Example:
$ %s query distr gen-report 2006-01-02 2006-02-02 cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			startDate, err := time.Parse(ReportDateFormat, args[0])
			if err != nil {
				return sdkerrors.Wrap(err, "unable to parse start time")
			}

			node, err := cliCtx.GetNode()
			if err != nil {
				return sdkerrors.Wrap(err, "unable to get node")
			}

			startHeight, err := findBlockHeightFromDate(startDate, node, cliCtx.Height)
			if err != nil {
				return sdkerrors.Wrap(err, "unable to find block height from start time")
			}

			endDate, err := time.Parse(ReportDateFormat, args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "unable to parse end time")
			}

			var validatorAddresses []sdk.ValAddress
			for _, bech32 := range args[2:] {
				validatorAddr, err := sdk.ValAddressFromBech32(bech32)
				if err != nil {
					return sdkerrors.Wrap(err, "unable to parse validator address: "+bech32)
				}

				validatorAddresses = append(validatorAddresses, validatorAddr)
			}

			reportByValidator := map[string][][]sdk.DecCoins{}
			var blockDays []time.Time
			querier := ValidatorQuerierCli{
				cliCtx:     cliCtx,
				cdc:        cdc,
				queryRoute: queryRoute,
			}
			for _, validatorAddress := range validatorAddresses {
				aggByDay, days, err :=
					calculateValidatorRewardsAndCommissionPerDay(
						node, validatorAddress, querier, endDate, *startHeight, cliCtx.Height)
				if err != nil {
					return err
				}
				reportByValidator[validatorAddress.String()] = aggByDay
				blockDays = days
			}

			fmt.Println("date,type,amount,currency")
			for i, day := range blockDays {
				finalCommission := sdk.DecCoins{}
				finalReward := sdk.DecCoins{}
				for _, results := range reportByValidator {
					finalCommission.Add(results[i][0])
					finalReward.Add(results[i][1])
				}
				for _, coin := range finalCommission {
					fmt.Printf("%d/%d/%d,commission,%s,%s\n", day.Year(), day.Month(), day.Day(), coin.Amount.String(), coin.Denom)
				}
				for _, coin := range finalReward {
					fmt.Printf("%d/%d/%d,reward,%s,%s\n", day.Year(), day.Month(), day.Day(), coin.Amount.String(), coin.Denom)
				}
			}

			return nil
		},
	}
}
