package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

// InitGenesis updates this keeper's store using the provided GenesisState.
func (k Keeper) InitGenesis(ctx sdk.Context, genesisState *quarantine.GenesisState) {
	for _, toAddrStr := range genesisState.QuarantinedAddresses {
		toAddr := sdk.MustAccAddressFromBech32(toAddrStr)
		if err := k.SetOptIn(ctx, toAddr); err != nil {
			panic(err)
		}
	}

	for _, qar := range genesisState.AutoResponses {
		toAddr := sdk.MustAccAddressFromBech32(qar.ToAddress)
		fromAddr := sdk.MustAccAddressFromBech32(qar.FromAddress)
		k.SetAutoResponse(ctx, toAddr, fromAddr, qar.Response)
	}

	totalQuarantined := sdk.Coins{}
	for _, qf := range genesisState.QuarantinedFunds {
		toAddr := sdk.MustAccAddressFromBech32(qf.ToAddress)
		qr := quarantine.NewQuarantineRecord(qf.UnacceptedFromAddresses, qf.Coins, qf.Declined)
		k.SetQuarantineRecord(ctx, toAddr, qr)
		totalQuarantined = totalQuarantined.Add(qf.Coins...)
	}

	if !totalQuarantined.IsZero() {
		qFundHolderBalance := k.bankKeeper.GetAllBalances(ctx, k.fundsHolder)
		if _, hasNeg := qFundHolderBalance.SafeSub(totalQuarantined...); hasNeg {
			panic(fmt.Errorf("quarantine fund holder account %q does not have enough funds %q to cover quarantined funds %q",
				k.fundsHolder.String(), qFundHolderBalance.String(), totalQuarantined.String()))
		}
	}
}

// ExportGenesis reads this keeper's entire state and returns it as a GenesisState.
func (k Keeper) ExportGenesis(ctx sdk.Context) *quarantine.GenesisState {
	qAddrs := k.GetAllQuarantinedAccounts(ctx)
	autoResps := k.GetAllAutoResponseEntries(ctx)
	qFunds := k.GetAllQuarantinedFunds(ctx)

	return quarantine.NewGenesisState(qAddrs, autoResps, qFunds)
}

// GetAllQuarantinedAccounts gets the bech32 string of every account that have opted into quarantine.
// This is designed for use with ExportGenesis. See also IterateQuarantinedAccounts.
func (k Keeper) GetAllQuarantinedAccounts(ctx sdk.Context) []string {
	var rv []string
	k.IterateQuarantinedAccounts(ctx, func(toAddr sdk.AccAddress) bool {
		rv = append(rv, toAddr.String())
		return false
	})
	return rv
}

// GetAllAutoResponseEntries gets an AutoResponseEntry entry for every quarantine auto-response that has been set.
// This is designed for use with ExportGenesis. See also IterateAutoResponses.
func (k Keeper) GetAllAutoResponseEntries(ctx sdk.Context) []*quarantine.AutoResponseEntry {
	var rv []*quarantine.AutoResponseEntry
	k.IterateAutoResponses(ctx, nil, func(toAddr, fromAddr sdk.AccAddress, resp quarantine.AutoResponse) bool {
		rv = append(rv, quarantine.NewAutoResponseEntry(toAddr, fromAddr, resp))
		return false
	})
	return rv
}

// GetAllQuarantinedFunds gets a QuarantinedFunds entry for each QuarantineRecord.
// This is designed for use with ExportGenesis. See also IterateQuarantineRecords.
func (k Keeper) GetAllQuarantinedFunds(ctx sdk.Context) []*quarantine.QuarantinedFunds {
	var rv []*quarantine.QuarantinedFunds
	k.IterateQuarantineRecords(ctx, nil, func(toAddr, _ sdk.AccAddress, funds *quarantine.QuarantineRecord) bool {
		rv = append(rv, funds.AsQuarantinedFunds(toAddr))
		return false
	})
	return rv
}
