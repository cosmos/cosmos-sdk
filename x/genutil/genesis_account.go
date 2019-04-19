package genutil

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// GenesisAccount defines an account initialized at genesis.
type GenesisAccount struct {
	Address       sdk.AccAddress `json:"address"`
	Coins         sdk.Coins      `json:"coins"`
	Sequence      uint64         `json:"sequence_number"`
	AccountNumber uint64         `json:"account_number"`

	// vesting account fields
	OriginalVesting  sdk.Coins `json:"original_vesting"`  // total vesting coins upon initialization
	DelegatedFree    sdk.Coins `json:"delegated_free"`    // delegated vested coins at time of delegation
	DelegatedVesting sdk.Coins `json:"delegated_vesting"` // delegated vesting coins at time of delegation
	StartTime        int64     `json:"start_time"`        // vesting start time (UNIX Epoch time)
	EndTime          int64     `json:"end_time"`          // vesting end time (UNIX Epoch time)
}

func NewGenesisAccount(acc *auth.BaseAccount) GenesisAccount {
	return GenesisAccount{
		Address:       acc.Address,
		Coins:         acc.Coins,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
	}
}

func NewGenesisAccountI(acc auth.Account) GenesisAccount {
	gacc := GenesisAccount{
		Address:       acc.GetAddress(),
		Coins:         acc.GetCoins(),
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence(),
	}

	vacc, ok := acc.(auth.VestingAccount)
	if ok {
		gacc.OriginalVesting = vacc.GetOriginalVesting()
		gacc.DelegatedFree = vacc.GetDelegatedFree()
		gacc.DelegatedVesting = vacc.GetDelegatedVesting()
		gacc.StartTime = vacc.GetStartTime()
		gacc.EndTime = vacc.GetEndTime()
	}

	return gacc
}

// convert GenesisAccount to auth.BaseAccount
func (ga *GenesisAccount) ToAccount() auth.Account {

	bacc := auth.NewBaseAccount(ga.Address, ga.Coins.Sort(),
		nil, ga.AccountNumber, ga.Sequence)

	if !ga.OriginalVesting.IsZero() {

		baseVestingAcc := auth.NewBaseVestingAccount(
			bacc, ga.OriginalVesting, ga.DelegatedFree,
			ga.DelegatedVesting, ga.EndTime)

		switch {
		case ga.StartTime != 0 && ga.EndTime != 0:
			return auth.NewContinuousVestingAccountRaw(baseVestingAcc, ga.StartTime)
		case ga.EndTime != 0:
			return auth.NewDelayedVestingAccountRaw(baseVestingAcc)
		default:
			panic(fmt.Sprintf("invalid genesis vesting account: %+v", ga))
		}
	}

	return bacc
}

// XXX TODO refactor this to only take in Accounts and return Accounts
func addGenesisAccount(
	cdc *codec.Codec, appState app.GenesisState, addr sdk.AccAddress,
	coins, vestingAmt sdk.Coins, vestingStart, vestingEnd int64,
) (app.GenesisState, error) {

	for _, stateAcc := range appState.Accounts {
		if stateAcc.Address.Equals(addr) {
			return appState, fmt.Errorf("the application state already contains account %v", addr)
		}
	}

	acc := auth.NewBaseAccountWithAddress(addr)
	acc.Coins = coins

	if vestingAmt.IsZero() {
		appState.Accounts = append(appState.Accounts, app.NewGenesisAccount(&acc))
		return appState, nil
	}

	bvacc := auth.NewBaseVestingAccount(&acc, vestingAmt, sdk.NewCoins(), sdk.NewCoins(), vestingEnd)
	if bvacc.OriginalVesting.IsAllGT(acc.Coins) {
		return appState, fmt.Errorf("vesting amount cannot be greater than total amount")
	}
	if vestingStart >= vestingEnd {
		return appState, fmt.Errorf("vesting start time must before end time")
	}

	var vacc auth.VestingAccount
	if vestingStart != 0 {
		vacc = auth.NewContinuousVestingAccountRaw(bvacc, vestingStart)
	} else {
		vacc = auth.NewDelayedVestingAccountRaw(bvacc)
	}

	appState.Accounts = append(appState.Accounts, app.NewGenesisAccountI(vacc))
	return appState, nil
}
