package simulation

import (
	"math/rand"

	"cosmossdk.io/core/address"
	"cosmossdk.io/x/nft" //nolint:staticcheck // deprecated and to be removed

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// genClasses returns a slice of nft class.
func genClasses(r *rand.Rand, accounts []simtypes.Account) []*nft.Class {
	classes := make([]*nft.Class, len(accounts)-1)
	for i := 0; i < len(accounts)-1; i++ {
		classes[i] = &nft.Class{
			Id:          simtypes.RandStringOfLength(r, 10),
			Name:        simtypes.RandStringOfLength(r, 10),
			Symbol:      simtypes.RandStringOfLength(r, 10),
			Description: simtypes.RandStringOfLength(r, 10),
			Uri:         simtypes.RandStringOfLength(r, 10),
		}
	}
	return classes
}

// genNFT returns a slice of nft.
func genNFT(r *rand.Rand, classID string, accounts []simtypes.Account, ac address.Codec) []*nft.Entry {
	entries := make([]*nft.Entry, len(accounts)-1)
	for i := 0; i < len(accounts)-1; i++ {
		owner := accounts[i]
		oast, err := ac.BytesToString(owner.Address.Bytes())
		if err != nil {
			panic(err)
		}
		entries[i] = &nft.Entry{
			Owner: oast,
			Nfts: []*nft.NFT{
				{
					ClassId: classID,
					Id:      simtypes.RandStringOfLength(r, 10),
					Uri:     simtypes.RandStringOfLength(r, 10),
				},
			},
		}
	}
	return entries
}

// RandomizedGenState generates a random GenesisState for nft.
func RandomizedGenState(simState *module.SimulationState, ac address.Codec) {
	var classes []*nft.Class
	simState.AppParams.GetOrGenerate(
		"nft", &classes, simState.Rand,
		func(r *rand.Rand) { classes = genClasses(r, simState.Accounts) },
	)

	var entries []*nft.Entry
	simState.AppParams.GetOrGenerate(
		"nft", &entries, simState.Rand,
		func(r *rand.Rand) {
			class := classes[r.Int63n(int64(len(classes)))]
			entries = genNFT(r, class.Id, simState.Accounts, ac)
		},
	)

	nftGenesis := &nft.GenesisState{
		Classes: classes,
		Entries: entries,
	}
	simState.GenState[nft.ModuleName] = simState.Cdc.MustMarshalJSON(nftGenesis)
}
