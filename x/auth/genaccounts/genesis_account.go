package genaccounts

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// GenesisAccount is a struct for account initialization used exclusively during genesis
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

// validate the the VestingAccount parameters are possible
func (ga GenesisAccount) Validate() error {
	if !ga.OriginalVesting.IsZero() {
		if ga.OriginalVesting.IsAnyGT(ga.Coins) {
			return errors.New("vesting amount cannot be greater than total amount")
		}
		if ga.StartTime >= ga.EndTime {
			return errors.New("vesting start-time cannot be before end-time")
		}
	}
	return nil
}

// NewGenesisAccount creates a new GenesisAccount object
func NewGenesisAccountRaw(address sdk.AccAddress, coins,
	vestingAmount sdk.Coins, vestingStartTime, vestingEndTime int64) GenesisAccount {

	return GenesisAccount{
		Address:          address,
		Coins:            coins,
		Sequence:         0,
		AccountNumber:    0, // ignored set by the account keeper during InitGenesis
		OriginalVesting:  vestingAmount,
		DelegatedFree:    sdk.Coins{}, // ignored
		DelegatedVesting: sdk.Coins{}, // ignored
		StartTime:        vestingStartTime,
		EndTime:          vestingEndTime,
	}
}

func NewGenesisAccount(acc *auth.BaseAccount) GenesisAccount {
	return GenesisAccount{
		Address:       acc.Address,
		Coins:         acc.Coins,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
	}
}

func NewGenesisAccountI(acc auth.Account) (GenesisAccount, error) {
	gacc := GenesisAccount{
		Address:       acc.GetAddress(),
		Coins:         acc.GetCoins(),
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence(),
	}

	if err := gacc.Validate(); err != nil {
		return gacc, err
	}

	vacc, ok := acc.(auth.VestingAccount)
	if ok {
		gacc.OriginalVesting = vacc.GetOriginalVesting()
		gacc.DelegatedFree = vacc.GetDelegatedFree()
		gacc.DelegatedVesting = vacc.GetDelegatedVesting()
		gacc.StartTime = vacc.GetStartTime()
		gacc.EndTime = vacc.GetEndTime()
	}

	return gacc, nil
}

// convert GenesisAccount to auth.Account
func (ga *GenesisAccount) ToAccount() auth.Account {
	bacc := auth.NewBaseAccount(ga.Address, ga.Coins.Sort(), nil, ga.AccountNumber, ga.Sequence)

	if !ga.OriginalVesting.IsZero() {
		baseVestingAcc := auth.NewBaseVestingAccount(
			bacc, ga.OriginalVesting, ga.DelegatedFree,
			ga.DelegatedVesting, ga.EndTime,
		)

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

//___________________________________
type GenesisAccounts []GenesisAccount

// genesis accounts contain an address
func (gaccs GenesisAccounts) Contains(acc sdk.AccAddress) bool {
	for _, gacc := range gaccs {
		if gacc.Address.Equals(acc) {
			return true
		}
	}
	return false
}
