package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// GenNFTGenesisState generates a random GenesisState for nft
func GenNFTGenesisState(cdc *codec.Codec, r *rand.Rand, accs []simulation.Account, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	const (
		Kitties = "crypto-kitties"
		Doggos  = "crypto-doggos"
	)

	collections := types.NewCollections(types.NewCollection(Kitties, types.NFTs{}), types.NewCollection(Doggos, types.NFTs{}))
	var ownerships []types.Owner

	for _, acc := range accs {
		if r.Intn(100) < 50 {
			baseNFT := types.NewBaseNFT(
				simulation.RandStringOfLength(r, 10), // id
				acc.Address,
				simulation.RandStringOfLength(r, 45), // tokenURI
			)

			var idCollection types.IDCollection
			var err error
			if r.Intn(100) < 50 {
				collections[0], err = collections[0].AddNFT(&baseNFT)
				if err != nil {
					panic(err)
				}
				idCollection = types.NewIDCollection(Kitties, []string{baseNFT.ID})
			} else {
				collections[1], err = collections[1].AddNFT(&baseNFT)
				if err != nil {
					panic(err)
				}
				idCollection = types.NewIDCollection(Doggos, []string{baseNFT.ID})
			}
			ownership := types.NewOwner(acc.Address, idCollection)
			ownerships = append(ownerships, ownership)
		}
	}

	nftGenesis := types.NewGenesisState(ownerships, collections)

	fmt.Printf("Selected randomly generated NFT parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, nftGenesis))
	genesisState[types.ModuleName] = cdc.MustMarshalJSON(nftGenesis)
}
