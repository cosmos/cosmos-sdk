package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func queryTotalSupply(cliCtx context.CLIContext, m codec.Marshaler) error {
	params := types.NewQueryTotalSupplyParams(1, 0) // no pagination
	bz, err := m.MarshalJSON(params)
	if err != nil {
		return err
	}

	res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryTotalSupply), bz)
	if err != nil {
		return err
	}

	var totalSupply sdk.Coins
	err = m.UnmarshalJSON(res, &totalSupply)
	if err != nil {
		return err
	}

	return cliCtx.PrintOutput(totalSupply)
}

func querySupplyOf(cliCtx context.CLIContext, m codec.Marshaler, denom string) error {
	params := types.NewQuerySupplyOfParams(denom)
	bz, err := m.MarshalJSON(params)
	if err != nil {
		return err
	}

	res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QuerySupplyOf), bz)
	if err != nil {
		return err
	}

	var supply sdk.Int
	err = m.UnmarshalJSON(res, &supply)
	if err != nil {
		return err
	}

	return cliCtx.PrintOutput(supply)
}
