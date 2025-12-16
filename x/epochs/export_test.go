package epochs

import "github.com/cosmos/cosmos-sdk/x/epochs/keeper"

func (am AppModule) Keeper() *keeper.Keeper {
	return am.keeper
}
