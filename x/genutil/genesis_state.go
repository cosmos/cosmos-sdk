package genutil

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// State to Unmarshal
type GenesisState struct {
	GenTxs []json.RawMessage `json:"gentxs"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(genTxs []json.RawMessage) GenesisState {
	return GenesisState{
		GenTxs: genTxs,
	}
}

// NewGenesisStateFromStdTx creates a new GenesisState object
// from auth transactions
func NewGenesisStateFromStdTx(genTxs []auth.StdTx) GenesisState {
	genTxsBz := make([]json.RawMessage, len(genTxs))
	for i, genTx := range genTxs {
		genTxsBz[i] = moduleCdc.MustMarshalJSON(genTx)
	}
	return GenesisState{
		GenTxs: genTxsBz,
	}
}

// get the genutil genesis state from the expected app state
func GetGenesisStateFromAppState(cdc *codec.Codec, appState ExpectedAppGenesisState) GenesisState {
	var genesisState GenesisState
	cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	return genesisState
}

// set the genutil genesis state within the expected app state
func SetGenesisStateInAppState(cdc *codec.Codec,
	appState ExpectedAppGenesisState, genesisState GenesisState) ExpectedAppGenesisState {

	genesisStateBz := cdc.MustMarshalJSON(genesisState)
	appState[ModuleName] = genesisStateBz
	return appState
}

// validate GenTx transactions
func ValidateGenesis(genesisState GenesisState) error {
	for i, genTx := range genesisState.GenTxs {
		var tx auth.StdTx
		if err := moduleCdc.UnmarshalJSON(genTx, &tx); err != nil {
			return err
		}

		msgs := tx.GetMsgs()
		if len(msgs) != 1 {
			return errors.New(
				"must provide genesis StdTx with exactly 1 CreateValidator message")
		}

		if _, ok := msgs[0].(staking.MsgCreateValidator); !ok {
			return fmt.Errorf(
				"Genesis transaction %v does not contain a MsgCreateValidator", i)
		}
	}
	return nil
}
