// contrib/export/v0_34_x_to_v0_36_0
// NOTE: We skip v0.35.0
package main

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/tendermint/tendermint/types"
)

var source = "0_34"
var target = "0_36"

type (
//OldGenesisState json.RawMessage
//NewGenesisState json.RawMessage
)

func migrate(appState json.RawMessage) json.RawMessage {

	// use old types to load into appState
	// use new types to set into newAppState
	newAppState := appState

	return newAppState
}

func main() {
	genDoc, err := types.GenesisDocFromFile(fmt.Sprintf("./genesis_%s.json", source))
	if err != nil {
		panic(err)
	}
	newGenState := migrate(genDoc.AppState)
	genDoc.AppState = newGenState
	if err = genutil.ExportGenesisFile(genDoc, fmt.Sprintf("./genesis_%s.json", target)); err != nil {
		panic(err)
	}
}
