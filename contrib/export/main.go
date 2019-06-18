package main

import (
	"flag"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	extypes "github.com/cosmos/cosmos-sdk/contrib/export/types"
	"github.com/cosmos/cosmos-sdk/contrib/export/v036"
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
)

func init() {
	log.SetPrefix("")
	log.SetFlags(0)

	// this flag will be used once we support more than one previous version
	flag.StringVar(&source, "source", "", "The SDK version that exported the genesis")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(),
			`Usage: %s [v0.36] [genesis.json] [-source v0.34]
Migrate the source genesis into the target version and export it as standard output
`, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	if flag.NArg() < 2 || flag.NArg() > 4 {
		log.Println("wrong number of arguments provided")
		flag.Usage()
		os.Exit(1)
	}
	target = flag.Arg(0)
	importGenesis = flag.Arg(1)

	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	// Dump here an example genesis for CCI to test import from old types, and export to new ones
	genDoc, err := types.GenesisDocFromFile(importGenesis)
	if err != nil {
		log.Fatalln(err)
	}
	var initialState extypes.AppMap
	cdc.MustUnmarshalJSON(genDoc.AppState, &initialState)

	if migrationMap[target] == nil {
		log.Fatalln(fmt.Sprintf("Missing migration function for version %s", target))
	}
	newGenState := migrationMap[target](initialState, cdc)
	genDoc.AppState = cdc.MustMarshalJSON(newGenState)

	out, err := cdc.MarshalJSONIndent(genDoc, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(out))
}
