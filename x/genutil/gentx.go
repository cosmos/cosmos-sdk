package genutil

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

var (
	defaultTokens                  = sdk.TokensFromTendermintPower(100)
	defaultAmount                  = defaultTokens.String() + sdk.DefaultBondDenom
	defaultCommissionRate          = "0.1"
	defaultCommissionMaxRate       = "0.2"
	defaultCommissionMaxChangeRate = "0.01"
	defaultMinSelfDelegation       = "1"
)

func accountInGenesis(genesisState app.GenesisState, key sdk.AccAddress, coins sdk.Coins, cdc *codec.Codec) error {
	accountIsInGenesis := false

	stakingDataBz := genesisState.Modules[staking.ModuleName]
	var stakingData staking.GenesisState
	cdc.MustUnmarshalJSON(stakingDataBz, &stakingData)
	bondDenom := stakingData.Params.BondDenom

	// Check if the account is in genesis
	for _, acc := range genesisState.Accounts {
		// Ensure that account is in genesis
		if acc.Address.Equals(key) {

			// Ensure account contains enough funds of default bond denom
			if coins.AmountOf(bondDenom).GT(acc.Coins.AmountOf(bondDenom)) {
				return fmt.Errorf(
					"account %v is in genesis, but it only has %v%v available to stake, not %v%v",
					key.String(), acc.Coins.AmountOf(bondDenom), bondDenom, coins.AmountOf(bondDenom), bondDenom,
				)
			}
			accountIsInGenesis = true
			break
		}
	}

	if accountIsInGenesis {
		return nil
	}

	return fmt.Errorf("account %s in not in the app_state.accounts array of genesis.json", key)
}

func makeOutputFilepath(rootDir, nodeID string) (string, error) {
	writePath := filepath.Join(rootDir, "config", "gentx")
	if err := common.EnsureDir(writePath, 0700); err != nil {
		return "", err
	}
	return filepath.Join(writePath, fmt.Sprintf("gentx-%v.json", nodeID)), nil
}

func readUnsignedGenTxFile(cdc *codec.Codec, r io.Reader) (auth.StdTx, error) {
	var stdTx auth.StdTx
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return stdTx, err
	}
	err = cdc.UnmarshalJSON(bytes, &stdTx)
	return stdTx, err
}

// nolint: errcheck
func writeSignedGenTx(cdc *codec.Codec, outputDocument string, tx auth.StdTx) error {
	outputFile, err := os.OpenFile(outputDocument, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	json, err := cdc.MarshalJSON(tx)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(outputFile, "%s\n", json)
	return err
}

// deliver a genesis transaction
func DeliverGenTxs(ctx sdk.Context, cdc *codec.Codec, genTxs []json.RawMessage,
	stakingKeeper StakingKeeper, deliverTx func([]byte) abci.ResponseDeliverTx) []abci.ValidatorUpdate {

	for _, genTx := range genTxs {
		var tx auth.StdTx
		cdc.MustUnmarshalJSON(genTx, &tx)
		bz := cdc.MustMarshalBinaryLengthPrefixed(tx)
		res := deliverTx(bz)
		if !res.IsOK() {
			panic(res.Log)
		}
	}
	return stakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
}
