package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// State to Unmarshal
type GenesisState struct {
	Accounts []GenesisAccount           `json:"accounts"`
	Modules  map[string]json.RawMessage `json:"modules"`
	GenTxs   []json.RawMessage          `json:"gentxs"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(accounts []GenesisAccount, modules map[string]json.RawMessage, genTxs []json.RawMessage) GenesisState {
	return GenesisState{
		Accounts: accounts,
		Modules:  modules,
		GenTxs:   genTxs,
	}
}

// Sanitize sorts accounts and coin sets.
func (gs GenesisState) Sanitize() {
	sort.Slice(gs.Accounts, func(i, j int) bool {
		return gs.Accounts[i].AccountNumber < gs.Accounts[j].AccountNumber
	})

	for _, acc := range gs.Accounts {
		acc.Coins = acc.Coins.Sort()
	}
}

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

// XXX move to init unused besides there and in testing

// Create the core parameters for genesis initialization for gaia
// note that the pubkey input is this machines pubkey
func GaiaAppGenState(cdc *codec.Codec, genDoc tmtypes.GenesisDoc, appGenTxs []json.RawMessage) (
	genesisState GenesisState, err error) {

	if err = cdc.UnmarshalJSON(genDoc.AppState, &genesisState); err != nil {
		return genesisState, err
	}

	// if there are no gen txs to be processed, return the default empty state
	if len(appGenTxs) == 0 {
		return genesisState, errors.New("there must be at least one genesis tx")
	}

	stakingData := genesisState.StakingData
	for i, genTx := range appGenTxs {
		var tx auth.StdTx
		if err := cdc.UnmarshalJSON(genTx, &tx); err != nil {
			return genesisState, err
		}

		msgs := tx.GetMsgs()
		if len(msgs) != 1 {
			return genesisState, errors.New(
				"must provide genesis StdTx with exactly 1 CreateValidator message")
		}

		if _, ok := msgs[0].(staking.MsgCreateValidator); !ok {
			return genesisState, fmt.Errorf(
				"Genesis transaction %v does not contain a MsgCreateValidator", i)
		}
	}

	for _, acc := range genesisState.Accounts {
		for _, coin := range acc.Coins {
			if coin.Denom == genesisState.StakingData.Params.BondDenom {
				stakingData.Pool.NotBondedTokens = stakingData.Pool.NotBondedTokens.
					Add(coin.Amount) // increase the supply
			}
		}
	}

	genesisState.StakingData = stakingData
	genesisState.GenTxs = appGenTxs

	return genesisState, nil
}

// GaiaAppGenState but with JSON
func GaiaAppGenStateJSON(cdc *codec.Codec, genDoc tmtypes.GenesisDoc, appGenTxs []json.RawMessage) (
	appState json.RawMessage, err error) {

	// create the final app state
	genesisState, err := GaiaAppGenState(cdc, genDoc, appGenTxs)
	if err != nil {
		return nil, err
	}
	return codec.MarshalJSONIndent(cdc, genesisState)
}

// NewDefaultGenesisState generates the default state for gaia.
func NewDefaultGenesisState() GenesisState {

	moduleGS := make(map[string]json.RawMessage)
	for _, mb := range moduleBasics {
		moduleGS[mb.Name()] = mb.DefaultGenesisState(cdc)
	}

	return GenesisState{
		Accounts: nil,
		Modules:  moduleGS,
		//AuthData:     auth.DefaultGenesisState(),
		//BankData:     bank.DefaultGenesisState(),
		//StakingData:  staking.DefaultGenesisState(),
		//MintData:     mint.DefaultGenesisState(),
		//DistrData:    distr.DefaultGenesisState(),
		//GovData:      gov.DefaultGenesisState(),
		//CrisisData:   crisis.DefaultGenesisState(),
		//SlashingData: slashing.DefaultGenesisState(),
		GenTxs: nil,
	}
}

// validateGenesisStateAccounts performs validation of genesis accounts. It
// ensures that there are no duplicate accounts in the genesis state and any
// provided vesting accounts are valid.
func validateGenesisStateAccounts(accs []GenesisAccount) error {
	addrMap := make(map[string]bool, len(accs))
	for _, acc := range accs {
		addrStr := acc.Address.String()

		// disallow any duplicate accounts
		if _, ok := addrMap[addrStr]; ok {
			return fmt.Errorf("duplicate account found in genesis state; address: %s", addrStr)
		}

		// validate any vesting fields
		if !acc.OriginalVesting.IsZero() {
			if acc.EndTime == 0 {
				return fmt.Errorf("missing end time for vesting account; address: %s", addrStr)
			}

			if acc.StartTime >= acc.EndTime {
				return fmt.Errorf(
					"vesting start time must before end time; address: %s, start: %s, end: %s",
					addrStr,
					time.Unix(acc.StartTime, 0).UTC().Format(time.RFC3339),
					time.Unix(acc.EndTime, 0).UTC().Format(time.RFC3339),
				)
			}
		}

		addrMap[addrStr] = true
	}

	return nil
}
