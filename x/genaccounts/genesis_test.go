package genaccounts

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	authType "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestGenesis(t *testing.T) {
	config := setupTestInput()
	genCoin := sdk.NewCoins(sdk.Coin{
		Denom:  "okt",
		Amount: sdk.NewDec(10000),
	})

	_, pubKey, addr := KeyTestPubAddr()

	genesisAccount := NewGenesisAccount(
		authType.NewBaseAccount(addr, genCoin, pubKey, 0, 1))
	gaccs := make(GenesisState, 0)
	gaccs = append(gaccs, genesisAccount)
	InitGenesis(config.ctx, ModuleCdc, config.ak, gaccs)

	genesisState := ExportGenesis(config.ctx, config.ak)

	require.Equal(t, gaccs, genesisState)

	require.NotNil(t, genesisState)
	require.Len(t, genesisState, 1)

	config.ak.IterateAccounts(config.ctx,
		func(account authexported.Account) (stop bool) {
			require.Equal(t, account, genesisAccount.ToAccount())
			return false
		})
}
