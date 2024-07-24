package genutil

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	apibanktypes "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/core/genesis"
	"cosmossdk.io/math"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// SetGenTxsInAppGenesisState - sets the genesis transactions in the app genesis state
func SetGenTxsInAppGenesisState(
	cdc codec.JSONCodec, txJSONEncoder sdk.TxEncoder, appGenesisState map[string]json.RawMessage, genTxs []sdk.Tx,
) (map[string]json.RawMessage, error) {
	genesisState := types.GetGenesisStateFromAppState(cdc, appGenesisState)
	genTxsBz := make([]json.RawMessage, 0, len(genTxs))

	for _, genTx := range genTxs {
		txBz, err := txJSONEncoder(genTx)
		if err != nil {
			return appGenesisState, err
		}

		genTxsBz = append(genTxsBz, txBz)
	}

	genesisState.GenTxs = genTxsBz
	return types.SetGenesisStateInAppState(cdc, appGenesisState, genesisState), nil
}

// GetGenesisStateFromAppState returns x/bank GenesisState given raw application
// genesis state.
func GetBankGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *apibanktypes.GenesisState {
	var genesisState apibanktypes.GenesisState

	if appState["bank"] != nil {
		cdc.MustUnmarshalJSON(appState["bank"], &genesisState)
	}

	return &genesisState
}

// ValidateAccountInGenesis checks that the provided account has a sufficient
// balance in the set of genesis accounts.
func ValidateAccountInGenesis(
	appGenesisState map[string]json.RawMessage,
	addr string, coins sdk.Coins, cdc codec.JSONCodec,
) error {
	var stakingData stakingtypes.GenesisState
	cdc.MustUnmarshalJSON(appGenesisState[stakingtypes.ModuleName], &stakingData)
	bondDenom := stakingData.Params.BondDenom

	var err error

	accountIsInGenesis := false

	balances := GetBankGenesisStateFromAppState(cdc, appGenesisState).Balances
	for _, bal := range balances {
		accAddress := bal.Address
		accCoins := bal.Coins
		// ensure that account is in genesis
		if strings.EqualFold(accAddress, addr) {
			// ensure account contains enough funds of default bond denom
			bd := math.Int{}
			for _, coin := range accCoins {
				if coin.Denom == bondDenom {
					bd, _ = math.NewIntFromString(coin.Amount)
				}
			}
			if coins.AmountOf(bondDenom).GT(bd) {
				err = fmt.Errorf(
					"account %s has a balance in genesis, but it only has %v%s available to stake, not %v%s",
					addr, bd, bondDenom, coins.AmountOf(bondDenom), bondDenom,
				)

				return err
			}

			accountIsInGenesis = true
			return nil
		}
	}

	if err != nil {
		return err
	}

	if !accountIsInGenesis {
		return fmt.Errorf("account %s does not have a balance in the genesis state", addr)
	}

	return nil
}

// DeliverGenTxs iterates over all genesis txs, decodes each into a Tx and
// invokes the provided deliverTxfn with the decoded Tx. It returns the result
// of the staking module's ApplyAndReturnValidatorSetUpdates.
func DeliverGenTxs(
	ctx context.Context, genTxs []json.RawMessage,
	stakingKeeper types.StakingKeeper, deliverTx genesis.TxHandler,
	txEncodingConfig client.TxEncodingConfig,
) ([]module.ValidatorUpdate, error) {
	for _, genTx := range genTxs {
		tx, err := txEncodingConfig.TxJSONDecoder()(genTx)
		if err != nil {
			return nil, fmt.Errorf("failed to decode GenTx '%s': %w", genTx, err)
		}

		bz, err := txEncodingConfig.TxEncoder()(tx)
		if err != nil {
			return nil, fmt.Errorf("failed to encode GenTx '%s': %w", genTx, err)
		}

		err = deliverTx.ExecuteGenesisTx(bz)
		if err != nil {
			return nil, fmt.Errorf("failed to execute DeliverTx for '%s': %w", genTx, err)
		}
	}

	return stakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
}
