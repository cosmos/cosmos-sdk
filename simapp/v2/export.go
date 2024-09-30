package simapp

import (
	"context"
	"fmt"

	stakingtypes "cosmossdk.io/x/staking/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v2 "github.com/cosmos/cosmos-sdk/x/genutil/v2"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *SimApp[T]) ExportAppStateAndValidators(jailAllowedAddrs []string) (v2.ExportedApp, error) {
	// as if they could withdraw from the start of the next block
	ctx := context.Background()

	latestHeight, err := app.LoadLatestHeight()
	if err != nil {
		return v2.ExportedApp{}, err
	}

	genesis, err := app.ExportGenesis(ctx, latestHeight)
	if err != nil {
		return v2.ExportedApp{}, err
	}

	// get the current bonded validators
	resp, err := app.Query(ctx, 0, latestHeight, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded,
	})

	vals, ok := resp.(*stakingtypes.QueryValidatorsResponse)
	if !ok {
		return v2.ExportedApp{}, fmt.Errorf("invalid response, expected QueryValidatorsResponse")
	}

	// convert to genesis validator
	var genesisVals []sdk.GenesisValidator
	for _, val := range vals.Validators {
		pk, err := val.ConsPubKey()
		if err != nil {
			return v2.ExportedApp{}, err
		}
		jsonPk, err := cryptocodec.PubKeyFromProto(pk)
		if err != nil {
			return v2.ExportedApp{}, err
		}

		genesisVals = append(genesisVals, sdk.GenesisValidator{
			Address: sdk.ConsAddress(pk.Address()).Bytes(),
			PubKey:  jsonPk,
			Power:   val.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)),
			Name:    val.Description.Moniker,
		})
	}

	return v2.ExportedApp{
		AppState:   genesis,
		Height:     int64(latestHeight),
		Validators: genesisVals,
	}, err
}
