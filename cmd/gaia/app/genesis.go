package app

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/stake"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/spf13/pflag"
)

// DefaultKeyPass contains the default key password for genesis transactions
const DefaultKeyPass = "12345678"

var (
	// bonded tokens given to genesis validators/accounts
	freeFermionVal  = int64(100)
	freeFermionsAcc = sdk.NewInt(50)
)

// State to Unmarshal
type GenesisState struct {
	Accounts     []GenesisAccount      `json:"accounts"`
	StakeData    stake.GenesisState    `json:"stake"`
	DistrData    distr.GenesisState    `json:"distr"`
	GovData      gov.GenesisState      `json:"gov"`
	SlashingData slashing.GenesisState `json:"slashing"`
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
	fsAppGenState := pflag.NewFlagSet("", pflag.ContinueOnError)
	fsAppGenTx := pflag.NewFlagSet("", pflag.ContinueOnError)
	fsAppGenTx.String(server.FlagName, "", "validator moniker, required")
	fsAppGenTx.String(server.FlagClientHome, DefaultCLIHome,
		"home directory for the client, used for key generation")
	fsAppGenTx.Bool(server.FlagOWK, false, "overwrite the accounts created")

	return server.AppInit{
		FlagsAppGenState: fsAppGenState,
		FlagsAppGenTx:    fsAppGenTx,
		AppGenState:      GaiaAppGenStateJSON,
		AppGenTx:         GaiaAppGenTx,
	}
}

// simple genesis tx
type GaiaGenTx struct {
	Name    string         `json:"name"`
	Address sdk.AccAddress `json:"address"`
	PubKey  string         `json:"pub_key"`
}

// GaiaAppGenTx generates a Gaia genesis transaction.
func GaiaAppGenTx(
	cdc *codec.Codec, pk crypto.PubKey, genTxConfig config.GenTx,
) (appGenTx auth.StdTx, cliPrint json.RawMessage, err error) {
	if genTxConfig.Name == "" {
		err = errors.New("Must specify --name (validator moniker)")
		return
	}

	buf := client.BufferStdin()
	prompt := fmt.Sprintf("Password for account '%s' (default %s):", genTxConfig.Name, DefaultKeyPass)

	keyPass, err := client.GetPassword(prompt, buf)
	if err != nil && keyPass != "" {
		// An error was returned that either failed to read the password from
		// STDIN or the given password is not empty but failed to meet minimum
		// length requirements.
		return
	}

	if keyPass == "" {
		keyPass = DefaultKeyPass
	}

	addr, secret, err := server.GenerateSaveCoinKey(
		genTxConfig.CliRoot,
		genTxConfig.Name,
		keyPass,
		genTxConfig.Overwrite,
	)
	if err != nil {
		return
	}

	mm := map[string]string{"secret": secret}
	bz, err := cdc.MarshalJSON(mm)
	if err != nil {
		return
	}

	desc := stake.NewDescription(genTxConfig.Name, "", "", "")
	comm := stakeTypes.CommissionMsg{}
	msg := stake.NewMsgCreateValidator(sdk.ValAddress(addr), pk, sdk.NewInt64Coin("steak", 50), desc, comm)
	appGenTx = auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, nil, "")

	cliPrint = json.RawMessage(bz)

	return
}

// Create the core parameters for genesis initialization for gaia
// note that the pubkey input is this machines pubkey
func GaiaAppGenState(cdc *codec.Codec, appGenTxs []json.RawMessage) (genesisState GenesisState, err error) {
	if len(appGenTxs) == 0 {
		err = errors.New("must provide at least genesis transaction")
		return
	}

	// start with the default staking genesis state
	stakeData := stake.DefaultGenesisState()
	slashingData := slashing.DefaultGenesisState()

	// get genesis flag account information
	genaccs := make([]GenesisAccount, len(appGenTxs))

	for i, jsonRawMessage := range appGenTxs {
		var genTx auth.StdTx
		err = cdc.UnmarshalJSON(jsonRawMessage, &genTx)
		if err != nil {
			return
		}
		msgs := genTx.GetMsgs()
		if len(msgs) == 0 {
			err = errors.New("must provide at least genesis message")
			return
		}
		msg := msgs[0].(stake.MsgCreateValidator)

		// create the genesis account, give'm few steaks and a buncha token with there name
		genaccs[i] = genesisAccountFromMsgCreateValidator(msg)
		stakeData.Pool.LooseTokens = stakeData.Pool.LooseTokens.Add(sdk.NewDecFromInt(freeFermionsAcc)) // increase the supply

		// add the validator
		if len(msg.Description.Moniker) > 0 {
			stakeData = addValidatorToStakeData(msg, stakeData)
		}

	}

	// create the final app state
	genesisState = GenesisState{
		Accounts:     genaccs,
		StakeData:    stakeData,
		DistrData:    distr.DefaultGenesisState(),
		GovData:      gov.DefaultGenesisState(),
		SlashingData: slashingData,
	}

	return
}

func addValidatorToStakeData(msg stake.MsgCreateValidator, stakeData stake.GenesisState) stake.GenesisState {
	validator := stake.NewValidator(
		sdk.ValAddress(msg.ValidatorAddr), msg.PubKey, msg.Description,
	)

	stakeData.Pool.LooseTokens = stakeData.Pool.LooseTokens.Add(sdk.NewDec(freeFermionVal)) // increase the supply

	// add some new shares to the validator
	var issuedDelShares sdk.Dec
	validator, stakeData.Pool, issuedDelShares = validator.AddTokensFromDel(stakeData.Pool, sdk.NewInt(freeFermionVal))
	stakeData.Validators = append(stakeData.Validators, validator)

	// create the self-delegation from the issuedDelShares
	delegation := stake.Delegation{
		DelegatorAddr: sdk.AccAddress(validator.OperatorAddr),
		ValidatorAddr: validator.OperatorAddr,
		Shares:        issuedDelShares,
		Height:        0,
	}

	stakeData.Bonds = append(stakeData.Bonds, delegation)
	return stakeData
}

func genesisAccountFromMsgCreateValidator(msg stake.MsgCreateValidator) GenesisAccount {
	accAuth := auth.NewBaseAccountWithAddress(sdk.AccAddress(msg.ValidatorAddr))
	accAuth.Coins = []sdk.Coin{msg.Delegation}
	return NewGenesisAccount(&accAuth)
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
	err = stake.ValidateGenesis(genesisState.StakeData)
	if err != nil {
		return
	}
	return
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
func GaiaAppGenStateJSON(cdc *codec.Codec, appGenTxs []auth.StdTx) (appState json.RawMessage, err error) {
	// create the final app state
	genesisState, err := GaiaAppGenState(cdc, appGenTxs)
	if err != nil {
		return nil, err
	}
	appState, err = codec.MarshalJSONIndent(cdc, genesisState)
	return
}
