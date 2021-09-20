package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// InitGenesis initializes the bank module's state from a given genesis state.
func (k BaseKeeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	totalSupply := sdk.Coins{}
	genState.Balances = types.SanitizeGenesisBalances(genState.Balances)

	for _, balance := range genState.Balances {
		addr, err := sdk.AccAddressFromBech32(balance.Address)
		if err != nil {
			panic(err)
		}

		if err := k.initBalances(ctx, addr, balance.Coins); err != nil {
			panic(fmt.Errorf("error on setting balances %w", err))
		}

		totalSupply = totalSupply.Add(balance.Coins...)
	}

	if !genState.Supply.Empty() && !genState.Supply.IsEqual(totalSupply) {
		panic(fmt.Errorf("genesis supply is incorrect, expected %v, got %v", genState.Supply, totalSupply))
	}

	for _, supply := range totalSupply {
		k.setSupply(ctx, supply)
	}

	for _, meta := range genState.DenomMetadata {
		k.SetDenomMetaData(ctx, meta)
	}
}

// trackDelegation tracks the delegation of the given account if it is a vesting account
func (k BaseKeeper) trackDelegation(ctx sdk.Context, addr sdk.AccAddress, balance, amt sdk.Coins) error {
	acc := k.ak.GetAccount(ctx, addr)
	if acc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	vacc, ok := acc.(vestexported.VestingAccount)
	if ok {
		// TODO: return error on account.TrackDelegation
		vacc.TrackDelegation(ctx.BlockHeader().Time, balance, amt)
		k.ak.SetAccount(ctx, acc)
	}

	return nil
}

// GetPaginatedTotalSupply queries for the supply, ignoring 0 coins, with a given pagination
func (k BaseKeeper) GetPaginatedTotalSupply(ctx sdk.Context, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	supplyStore := prefix.NewStore(store, types.SupplyKey)

	supply := sdk.NewCoins()

	pageRes, err := query.Paginate(supplyStore, pagination, func(key, value []byte) error {
		var amount sdk.Int
		err := amount.Unmarshal(value)
		if err != nil {
			return fmt.Errorf("unable to convert amount string to Int %v", err)
		}

		// `Add` omits the 0 coins addition to the `supply`.
		supply = supply.Add(sdk.NewCoin(string(key), amount))
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return supply, pageRes, nil
}

// ExportGenesis returns the bank module's genesis state.
func (k BaseKeeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	totalSupply, _, err := k.GetPaginatedTotalSupply(ctx, &query.PageRequest{Limit: query.MaxLimit})
	if err != nil {
		panic(fmt.Errorf("unable to fetch total supply %v", err))
	}

	return types.NewGenesisState(
		k.GetParams(ctx),
		k.GetAccountsBalances(ctx),
		totalSupply,
		k.GetAllDenomMetaData(ctx),
	)
}
