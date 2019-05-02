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

// XXX TODO
func ValidateAccountInGenesis(appGenesisState ExpectedAppGenesisState, iterateGenAcc IterateGenesisAccountsFn,
	key sdk.AccAddress, coins sdk.Coins, cdc *codec.Codec) error {

	accountIsInGenesis := false

	// TODO refactor out bond denom to common state area
	stakingDataBz := appGenesisState[staking.ModuleName]
	var stakingData staking.GenesisState
	cdc.MustUnmarshalJSON(stakingDataBz, &stakingData)
	bondDenom := stakingData.Params.BondDenom

	genUtilDataBz := appGenesisState[staking.ModuleName]
	var genesisState GenesisState
	cdc.MustUnmarshalJSON(genUtilDataBz, &genesisState)

	var err error
	iterateGenAcc(cdc, appGenesisState,
		func(acc auth.Account) (stop bool) {
			accAddress := acc.GetAddress()
			accCoins := acc.GetCoins()

			// Ensure that account is in genesis
			if accAddress.Equals(key) {

				// Ensure account contains enough funds of default bond denom
				if coins.AmountOf(bondDenom).GT(accCoins.AmountOf(bondDenom)) {
					err = fmt.Errorf(
						"account %v is in genesis, but it only has %v%v available to stake, not %v%v",
						key.String(), accCoins.AmountOf(bondDenom), bondDenom, coins.AmountOf(bondDenom), bondDenom,
					)
					return true
				}
				accountIsInGenesis = true
				return true
			}
			return false
		},
	)
	if err != nil {
		return err
	}

	if !accountIsInGenesis {
		return fmt.Errorf("account %s in not in the app_state.accounts array of genesis.json", key)
	}

	return nil
}

// XXX TODO
func MakeOutputFilepath(rootDir, nodeID string) (string, error) {
	writePath := filepath.Join(rootDir, "config", "gentx")
	if err := common.EnsureDir(writePath, 0700); err != nil {
		return "", err
	}
	return filepath.Join(writePath, fmt.Sprintf("gentx-%v.json", nodeID)), nil
}

// XXX TODO
func ReadUnsignedGenTxFile(cdc *codec.Codec, r io.Reader) (auth.StdTx, error) {
	var stdTx auth.StdTx
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return stdTx, err
	}
	err = cdc.UnmarshalJSON(bytes, &stdTx)
	return stdTx, err
}

// XXX TODO
func WriteSignedGenTx(cdc *codec.Codec, outputDocument string, tx auth.StdTx) error {
	outputFile, err := os.OpenFile(outputDocument, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	defer outputFile.Close()
	if err != nil {
		return err
	}
	json, err := cdc.MarshalJSON(tx)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(outputFile, "%s\n", json)
	return err
}

type deliverTxfn func([]byte) abci.ResponseDeliverTx

// deliver a genesis transaction
func DeliverGenTxs(ctx sdk.Context, cdc *codec.Codec, genTxs []json.RawMessage,
	stakingKeeper StakingKeeper, deliverTx deliverTxfn) []abci.ValidatorUpdate {

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
