package main

import (
	"flag"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	extypes "github.com/cosmos/cosmos-sdk/contrib/export/types"
	"github.com/cosmos/cosmos-sdk/contrib/export/v036"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/tendermint/tendermint/types"
	"log"
	"os"
	"path/filepath"
)

var (
	migrationMap = extypes.MigrationMap{
		"v0.36": v036.Migrate,
	}
	source        string
	target        string
	importGenesis string
	exportGenesis string
)

func init() {
	log.SetPrefix("")
	log.SetFlags(0)

	// this flag seems unnecessary, we can reintriduce it once we support multiple versions migration at once
	//flag.StringVar(&source, "s", "034", "SDK version that exported the genesis")
	flag.StringVar(&target, "v", "036", "Goal SDK version that will import the genesis")
	flag.StringVar(&importGenesis, "g", "genesis.json", "Source genesis file")
	flag.StringVar(&exportGenesis, "o", "genesis.json", "Output genesis file")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			`Usage: %s [-t 036] [-g genesis.json] [-o genesis.json]
Migrate the source genesis into the target version

`, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	// This function should be modularized, for now we read and dump genesis committed to git,
	// to simplify the creation of a CCI script that tests three different things:
	// - Reading old types,
	// - converting to new ones
	// - reading the committed new genesis and see if all works

	// We could add an invariant test for genesis, conversion should be identical to exporting the target genesis once imported

	cdc := codec.New()
	// Dump here an example genesis for CCI to test import from old types, and export to new ones
	genDoc, err := types.GenesisDocFromFile(importGenesis)
	if err != nil {
		panic(err)
	}
	var initialState extypes.AppMap
	cdc.MustUnmarshalJSON(genDoc.AppState, &initialState)

	if migrationMap[target] == nil {
		panic(fmt.Sprintf("Missing migration function for version %s", target))
	}
	newGenState := migrationMap[target](initialState, cdc)

	genDoc.AppState = cdc.MustMarshalJSON(newGenState)
	// Keep dumping updated genesis to test import of a new genesis directly
	if err = genutil.ExportGenesisFile(genDoc, exportGenesis); err != nil {
		panic(err)
	}
}
