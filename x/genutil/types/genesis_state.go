package types

import (
	"encoding/json"
	"fmt"

	tmos "github.com/tendermint/tendermint/libs/os"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(genTxs []json.RawMessage) *GenesisState {
	// Ensure genTxs is never nil, https://github.com/cosmos/cosmos-sdk/issues/5086
	if len(genTxs) == 0 {
		genTxs = make([]json.RawMessage, 0)
	}
	return &GenesisState{
		GenTxs: genTxs,
	}
}

// DefaultGenesisState returns the genutil module's default genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		GenTxs: []json.RawMessage{},
	}
}

// NewGenesisStateFromTx creates a new GenesisState object
// from auth transactions
func NewGenesisStateFromTx(txJSONEncoder sdk.TxEncoder, genTxs []sdk.Tx) *GenesisState {
	genTxsBz := make([]json.RawMessage, len(genTxs))
	for i, genTx := range genTxs {
		var err error
		genTxsBz[i], err = txJSONEncoder(genTx)
		if err != nil {
			panic(err)
		}
	}
	return NewGenesisState(genTxsBz)
}

// GetGenesisStateFromAppState gets the genutil genesis state from the expected app state
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}
	return &genesisState
}

// SetGenesisStateInAppState sets the genutil genesis state within the expected app state
func SetGenesisStateInAppState(
	cdc codec.JSONCodec, appState map[string]json.RawMessage, genesisState *GenesisState,
) map[string]json.RawMessage {
	genesisStateBz := cdc.MustMarshalJSON(genesisState)
	appState[ModuleName] = genesisStateBz
	return appState
}

// GenesisStateFromGenDoc creates the core parameters for genesis initialization
// for the application.
//
// NOTE: The pubkey input is this machines pubkey.
func GenesisStateFromGenDoc(genDoc tmtypes.GenesisDoc) (genesisState map[string]json.RawMessage, err error) {
	if err = json.Unmarshal(genDoc.AppState, &genesisState); err != nil {
		return genesisState, err
	}
	return genesisState, nil
}

// GenesisStateFromGenFile creates the core parameters for genesis initialization
// for the application.
//
// NOTE: The pubkey input is this machines pubkey.
func GenesisStateFromGenFile(genFile string) (genesisState map[string]json.RawMessage, genDoc *tmtypes.GenesisDoc, err error) {
	if !tmos.FileExists(genFile) {
		return genesisState, genDoc,
			fmt.Errorf("%s does not exist, run `init` first", genFile)
	}

	genDoc, err = tmtypes.GenesisDocFromFile(genFile)
	if err != nil {
		return genesisState, genDoc, err
	}

	genesisState, err = GenesisStateFromGenDoc(*genDoc)
	return genesisState, genDoc, err
}

// ValidateGenesis validates GenTx transactions
func ValidateGenesis(genesisState *GenesisState, txJSONDecoder sdk.TxDecoder) error {
	for _, genTx := range genesisState.GenTxs {
		_, err := ValidateAndGetGenTx(genTx, txJSONDecoder)
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateAndGetGenTx validates the genesis transaction and returns GenTx if valid
// it cannot verify the signature as it is stateless validation
func ValidateAndGetGenTx(genTx json.RawMessage, txJSONDecoder sdk.TxDecoder) (sdk.Tx, error) {
	tx, err := txJSONDecoder(genTx)
	if err != nil {
		return tx, fmt.Errorf("failed to decode gentx: %s, error: %s", genTx, err)
	}

	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		return tx, fmt.Errorf("unexpected number of GenTx messages; got: %d, expected: 1", len(msgs))
	}

	// TODO: abstract back to staking
	if _, ok := msgs[0].(*stakingtypes.MsgCreateValidator); !ok {
		return tx, fmt.Errorf("unexpected GenTx message type; expected: MsgCreateValidator, got: %T", msgs[0])
	}

	if err := msgs[0].ValidateBasic(); err != nil {
		return tx, fmt.Errorf("invalid GenTx '%s': %s", msgs[0], err)
	}

	return tx, nil
}
