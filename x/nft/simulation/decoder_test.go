package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pointnetwork/cosmos-point-sdk/crypto/keys/ed25519"
	"github.com/pointnetwork/cosmos-point-sdk/simapp"
	sdk "github.com/pointnetwork/cosmos-point-sdk/types"
	"github.com/pointnetwork/cosmos-point-sdk/types/kv"
	"github.com/pointnetwork/cosmos-point-sdk/x/nft"
	"github.com/pointnetwork/cosmos-point-sdk/x/nft/keeper"
	"github.com/pointnetwork/cosmos-point-sdk/x/nft/simulation"
)

var (
	ownerPk1   = ed25519.GenPrivKey().PubKey()
	ownerAddr1 = sdk.AccAddress(ownerPk1.Address())
)

func TestDecodeStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Codec
	dec := simulation.NewDecodeStore(cdc)

	class := nft.Class{
		Id:          "ClassID",
		Name:        "ClassName",
		Symbol:      "ClassSymbol",
		Description: "ClassDescription",
		Uri:         "ClassURI",
	}
	classBz, err := cdc.Marshal(&class)
	require.NoError(t, err)

	nft := nft.NFT{
		ClassId: "ClassID",
		Id:      "NFTID",
		Uri:     "NFTURI",
	}
	nftBz, err := cdc.Marshal(&nft)
	require.NoError(t, err)

	nftOfClassByOwnerValue := []byte{0x01}

	totalSupply := 1
	totalSupplyBz := sdk.Uint64ToBigEndian(1)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: []byte(keeper.ClassKey), Value: classBz},
			{Key: []byte(keeper.NFTKey), Value: nftBz},
			{Key: []byte(keeper.NFTOfClassByOwnerKey), Value: nftOfClassByOwnerValue},
			{Key: []byte(keeper.OwnerKey), Value: ownerAddr1},
			{Key: []byte(keeper.ClassTotalSupply), Value: totalSupplyBz},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectErr   bool
		expectedLog string
	}{
		{"Class", false, fmt.Sprintf("%v\n%v", class, class)},
		{"NFT", false, fmt.Sprintf("%v\n%v", nft, nft)},
		{"NFTOfClassByOwnerKey", false, fmt.Sprintf("%v\n%v", nftOfClassByOwnerValue, nftOfClassByOwnerValue)},
		{"OwnerKey", false, fmt.Sprintf("%v\n%v", ownerAddr1, ownerAddr1)},
		{"ClassTotalSupply", false, fmt.Sprintf("%v\n%v", totalSupply, totalSupply)},
		{"other", true, ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectErr {
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			} else {
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
