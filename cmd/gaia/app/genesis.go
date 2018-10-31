package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/stake"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	// bonded tokens given to genesis validators/accounts
	freeFermionVal  = int64(100)
	freeFermionsAcc = sdk.NewInt(150)
)

// State to Unmarshal
type GenesisState struct {
	Accounts     []GenesisAccount      `json:"accounts"`
	StakeData    stake.GenesisState    `json:"stake"`
	MintData     mint.GenesisState     `json:"mint"`
	DistrData    distr.GenesisState    `json:"distr"`
	GovData      gov.GenesisState      `json:"gov"`
	SlashingData slashing.GenesisState `json:"slashing"`
	GenTxs       []json.RawMessage     `json:"gentxs"`
}

func NewGenesisState(accounts []GenesisAccount, stakeData stake.GenesisState, mintData mint.GenesisState,
	distrData distr.GenesisState, govData gov.GenesisState, slashingData slashing.GenesisState) GenesisState {

	return GenesisState{
		Accounts:     accounts,
		StakeData:    stakeData,
		MintData:     mintData,
		DistrData:    distrData,
		GovData:      govData,
		SlashingData: slashingData,
	}
}

// GenesisAccount doesn't need pubkey or sequence
type GenesisAccount struct {
	Address sdk.AccAddress `json:"address"`
	Coins   sdk.Coins      `json:"coins"`
}

func NewGenesisAccount(acc *auth.BaseAccount) GenesisAccount {
	return GenesisAccount{
		Address: acc.Address,
		Coins:   acc.Coins,
	}
}

func NewGenesisAccountI(acc auth.Account) GenesisAccount {
	return GenesisAccount{
		Address: acc.GetAddress(),
		Coins:   acc.GetCoins(),
	}
}

// convert GenesisAccount to auth.BaseAccount
func (ga *GenesisAccount) ToAccount() (acc *auth.BaseAccount) {
	return &auth.BaseAccount{
		Address: ga.Address,
		Coins:   ga.Coins.Sort(),
	}
}

// get app init parameters for server init command
func GaiaAppInit() server.AppInit {

	return server.AppInit{
		AppGenState: GaiaAppGenStateJSON,
	}
}

// Create the core parameters for genesis initialization for gaia
// note that the pubkey input is this machines pubkey
func GaiaAppGenState(cdc *codec.Codec, genDoc tmtypes.GenesisDoc, appGenTxs []json.RawMessage) (
	genesisState GenesisState, err error) {

	if err = cdc.UnmarshalJSON(genDoc.AppState, &genesisState); err != nil {
		return
	}

	// if there are no gen txs to be processed, return the default empty state
	if len(appGenTxs) == 0 {
		err = errors.New("there must be at least one genesis tx")
		return
	}

	stakeData := genesisState.StakeData
	for _, genTx := range appGenTxs {
		var tx auth.StdTx
		err = cdc.UnmarshalJSON(genTx, &tx)
		if err != nil {
			return
		}
		msgs := tx.GetMsgs()
		if len(msgs) != 1 {
			err = errors.New("must provide genesis StdTx with exactly 1 CreateValidator message")
			return
		}
		_ = msgs[0].(stake.MsgCreateValidator)
	}

	for _, acc := range genesisState.Accounts {
		// create the genesis account, give'm few steaks and a buncha token with there name
		for _, coin := range acc.Coins {
			if coin.Denom == "steak" {
				stakeData.Pool.LooseTokens = stakeData.Pool.LooseTokens.
					Add(sdk.NewDecFromInt(coin.Amount)) // increase the supply
			}
		}
	}
	genesisState.StakeData = stakeData
	genesisState.GenTxs = appGenTxs
	return genesisState, nil
}

// DefaultGenesisState generates the default state for gaia.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Accounts:     nil,
		StakeData:    stake.DefaultGenesisState(),
		MintData:     mint.DefaultGenesisState(),
		DistrData:    distr.DefaultGenesisState(),
		GovData:      gov.DefaultGenesisState(),
		SlashingData: slashing.DefaultGenesisState(),
		GenTxs:       nil,
	}
}

// GaiaValidateGenesisState ensures that the genesis state obeys the expected invariants
// TODO: No validators are both bonded and jailed (#2088)
// TODO: Error if there is a duplicate validator (#1708)
// TODO: Ensure all state machine parameters are in genesis (#1704)
func GaiaValidateGenesisState(genesisState GenesisState) (err error) {
	err = validateGenesisStateAccounts(genesisState.Accounts)
	if err != nil {
		return
	}
	// skip stakeData validation as genesis is created from txs
	if len(genesisState.GenTxs) > 0 {
		return nil
	}
	return stake.ValidateGenesis(genesisState.StakeData)
}

// Ensures that there are no duplicate accounts in the genesis state,
func validateGenesisStateAccounts(accs []GenesisAccount) (err error) {
	addrMap := make(map[string]bool, len(accs))
	for i := 0; i < len(accs); i++ {
		acc := accs[i]
		strAddr := string(acc.Address)
		if _, ok := addrMap[strAddr]; ok {
			return fmt.Errorf("Duplicate account in genesis state: Address %v", acc.Address)
		}
		addrMap[strAddr] = true
	}
	return
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

// CollectStdTxs processes and validates application's genesis StdTxs and returns the list of
// validators,  appGenTxs, and persistent peers required to generate genesis.json.
func CollectStdTxs(cdc *codec.Codec, moniker string, genTxsDir string, genDoc tmtypes.GenesisDoc) (
	appGenTxs []auth.StdTx, persistentPeers string, err error) {
	var fos []os.FileInfo
	fos, err = ioutil.ReadDir(genTxsDir)
	if err != nil {
		return
	}

	// prepare a map of all accounts in genesis state to then validate
	// against the validators addresses
	var appState GenesisState
	err = cdc.UnmarshalJSON(genDoc.AppState, &appState)
	if err != nil {
		return
	}
	addrMap := make(map[string]GenesisAccount, len(appState.Accounts))
	for i := 0; i < len(appState.Accounts); i++ {
		acc := appState.Accounts[i]
		strAddr := string(acc.Address)
		addrMap[strAddr] = acc
	}

	var addresses []string
	for _, fo := range fos {
		filename := filepath.Join(genTxsDir, fo.Name())
		if !fo.IsDir() && (filepath.Ext(filename) != ".json") {
			continue
		}

		// get the genStdTx
		var jsonRawTx []byte
		jsonRawTx, err = ioutil.ReadFile(filename)
		if err != nil {
			return
		}
		var genStdTx auth.StdTx
		err = cdc.UnmarshalJSON(jsonRawTx, &genStdTx)
		if err != nil {
			return
		}
		appGenTxs = append(appGenTxs, genStdTx)

		nodeAddr := genStdTx.GetMemo()
		if len(nodeAddr) == 0 {
			err = fmt.Errorf("couldn't find node's address in %s", fo.Name())
			return
		}

		// genesis transactions must be single-message
		msgs := genStdTx.GetMsgs()
		if len(msgs) != 1 {
			err = errors.New("each genesis transaction must provide a single genesis message")
			return
		}

		// validate the validator address and funds against the accounts in the state
		msg := msgs[0].(stake.MsgCreateValidator)
		addr := string(sdk.AccAddress(msg.ValidatorAddr))
		acc, ok := addrMap[addr]
		if !ok {
			err = fmt.Errorf("account %v not in genesis.json: %+v", addr, addrMap)
			return
		}
		if acc.Coins.AmountOf(msg.Delegation.Denom).LT(msg.Delegation.Amount) {
			err = fmt.Errorf("insufficient fund for the delegation: %s < %s",
				acc.Coins.AmountOf(msg.Delegation.Denom), msg.Delegation.Amount)
		}

		// exclude itself from persistent peers
		if msg.Description.Moniker != moniker {
			addresses = append(addresses, nodeAddr)
		}
	}

	sort.Strings(addresses)
	persistentPeers = strings.Join(addresses, ",")

	return
}

func NewDefaultGenesisAccount(addr sdk.AccAddress) GenesisAccount {
	accAuth := auth.NewBaseAccountWithAddress(addr)
	accAuth.Coins = []sdk.Coin{
		{"fooToken", sdk.NewInt(1000)},
		{"steak", freeFermionsAcc},
	}
	return NewGenesisAccount(&accAuth)
}
