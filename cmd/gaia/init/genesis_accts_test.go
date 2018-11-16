package init

import (
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestAddGenesisAccount(t *testing.T) {
	cdc := codec.New()
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	type args struct {
		appState app.GenesisState
		addr     sdk.AccAddress
		coins    sdk.Coins
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"valid account",
		args{
			app.GenesisState{},
			addr1,
			sdk.Coins{},
		},
		false},
		{"dup account", args{
			app.GenesisState{Accounts: []app.GenesisAccount{app.GenesisAccount{Address:addr1}}},
			addr1,
			sdk.Coins{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := addGenesisAccount(cdc, tt.args.appState, tt.args.addr, tt.args.coins)
			require.Equal(t, tt.wantErr, (err != nil))
		})
	}
}
