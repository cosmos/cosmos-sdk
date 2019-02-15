package app

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

var (
	pk1   = ed25519.GenPrivKey().PubKey()
	pk2   = ed25519.GenPrivKey().PubKey()
	pk3   = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.ValAddress(pk1.Address())
	addr2 = sdk.ValAddress(pk2.Address())
	addr3 = sdk.ValAddress(pk3.Address())

	emptyAddr   sdk.ValAddress
	emptyPubkey crypto.PubKey
)

func makeGenesisState(t *testing.T, genTxs []auth.StdTx) GenesisState {
	// start with the default staking genesis state
	appState := NewDefaultGenesisState()
	stakingData := appState.StakingData
	genAccs := make([]GenesisAccount, len(genTxs))

	for i, genTx := range genTxs {
		msgs := genTx.GetMsgs()
		require.Equal(t, 1, len(msgs))
		msg := msgs[0].(staking.MsgCreateValidator)

		acc := auth.NewBaseAccountWithAddress(sdk.AccAddress(msg.ValidatorAddr))
		acc.Coins = sdk.Coins{sdk.NewInt64Coin(defaultBondDenom, 150)}
		genAccs[i] = NewGenesisAccount(&acc)
		stakingData.Pool.NotBondedTokens = stakingData.Pool.NotBondedTokens.Add(sdk.NewInt(150)) // increase the supply
	}

	// create the final app state
	appState.Accounts = genAccs
	return appState
}

func TestToAccount(t *testing.T) {
	priv := ed25519.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())
	authAcc := auth.NewBaseAccountWithAddress(addr)
	authAcc.SetCoins(sdk.Coins{sdk.NewInt64Coin(defaultBondDenom, 150)})
	genAcc := NewGenesisAccount(&authAcc)
	acc := genAcc.ToAccount()
	require.IsType(t, &auth.BaseAccount{}, acc)
	require.Equal(t, &authAcc, acc.(*auth.BaseAccount))

	vacc := auth.NewContinuousVestingAccount(
		&authAcc, time.Now().Unix(), time.Now().Add(24*time.Hour).Unix(),
	)
	genAcc = NewGenesisAccountI(vacc)
	acc = genAcc.ToAccount()
	require.IsType(t, &auth.ContinuousVestingAccount{}, acc)
	require.Equal(t, vacc, acc.(*auth.ContinuousVestingAccount))
}

func TestGaiaAppGenTx(t *testing.T) {
	cdc := MakeCodec()
	_ = cdc

	//TODO test that key overwrite flags work / no overwrites if set off
	//TODO test validator created has provided pubkey
	//TODO test the account created has the correct pubkey
}

func TestGaiaAppGenState(t *testing.T) {
	cdc := MakeCodec()
	_ = cdc
	var genDoc tmtypes.GenesisDoc

	// test unmarshalling error
	_, err := GaiaAppGenState(cdc, genDoc, []json.RawMessage{})
	require.Error(t, err)

	appState := makeGenesisState(t, []auth.StdTx{})
	genDoc.AppState, err = json.Marshal(appState)
	require.NoError(t, err)

	// test validation error
	_, err = GaiaAppGenState(cdc, genDoc, []json.RawMessage{})
	require.Error(t, err)

	// TODO test must provide at least genesis transaction
	// TODO test with both one and two genesis transactions:
	// TODO        correct: genesis account created, canididates created, pool token variance
}

func makeMsg(name string, pk crypto.PubKey) auth.StdTx {
	desc := staking.NewDescription(name, "", "", "")
	comm := staking.CommissionMsg{}
	msg := staking.NewMsgCreateValidator(sdk.ValAddress(pk.Address()), pk, sdk.NewInt64Coin(defaultBondDenom,
		50), desc, comm, sdk.OneInt())
	return auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, nil, "")
}

func TestGaiaGenesisValidation(t *testing.T) {
	genTxs := []auth.StdTx{makeMsg("test-0", pk1), makeMsg("test-1", pk2)}
	dupGenTxs := []auth.StdTx{makeMsg("test-0", pk1), makeMsg("test-1", pk1)}

	// require duplicate accounts fails validation
	genesisState := makeGenesisState(t, dupGenTxs)
	err := GaiaValidateGenesisState(genesisState)
	require.Error(t, err)

	// require invalid vesting account fails validation (invalid end time)
	genesisState = makeGenesisState(t, genTxs)
	genesisState.Accounts[0].OriginalVesting = genesisState.Accounts[0].Coins
	err = GaiaValidateGenesisState(genesisState)
	require.Error(t, err)
	genesisState.Accounts[0].StartTime = 1548888000
	genesisState.Accounts[0].EndTime = 1548775410
	err = GaiaValidateGenesisState(genesisState)
	require.Error(t, err)

	// require bonded + jailed validator fails validation
	genesisState = makeGenesisState(t, genTxs)
	val1 := staking.NewValidator(addr1, pk1, staking.NewDescription("test #2", "", "", ""))
	val1.Jailed = true
	val1.Status = sdk.Bonded
	genesisState.StakingData.Validators = append(genesisState.StakingData.Validators, val1)
	err = GaiaValidateGenesisState(genesisState)
	require.Error(t, err)

	// require duplicate validator fails validation
	val1.Jailed = false
	genesisState = makeGenesisState(t, genTxs)
	val2 := staking.NewValidator(addr1, pk1, staking.NewDescription("test #3", "", "", ""))
	genesisState.StakingData.Validators = append(genesisState.StakingData.Validators, val1)
	genesisState.StakingData.Validators = append(genesisState.StakingData.Validators, val2)
	err = GaiaValidateGenesisState(genesisState)
	require.Error(t, err)
}

func TestNewDefaultGenesisAccount(t *testing.T) {
	addr := secp256k1.GenPrivKeySecp256k1([]byte("")).PubKey().Address()
	acc := NewDefaultGenesisAccount(sdk.AccAddress(addr))
	require.Equal(t, sdk.NewInt(1000), acc.Coins.AmountOf("footoken"))
	require.Equal(t, sdk.TokensFromTendermintPower(150), acc.Coins.AmountOf(defaultBondDenom))
}

func TestGenesisStateSanitize(t *testing.T) {
	genesisState := makeGenesisState(t, nil)
	require.Nil(t, GaiaValidateGenesisState(genesisState))

	addr1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	authAcc1 := auth.NewBaseAccountWithAddress(addr1)
	authAcc1.SetCoins(sdk.Coins{
		sdk.NewInt64Coin("bcoin", 150),
		sdk.NewInt64Coin("acoin", 150),
	})
	authAcc1.SetAccountNumber(1)
	genAcc1 := NewGenesisAccount(&authAcc1)

	addr2 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	authAcc2 := auth.NewBaseAccountWithAddress(addr2)
	authAcc2.SetCoins(sdk.Coins{
		sdk.NewInt64Coin("acoin", 150),
		sdk.NewInt64Coin("bcoin", 150),
	})
	genAcc2 := NewGenesisAccount(&authAcc2)

	genesisState.Accounts = []GenesisAccount{genAcc1, genAcc2}
	require.True(t, genesisState.Accounts[0].AccountNumber > genesisState.Accounts[1].AccountNumber)
	require.Equal(t, genesisState.Accounts[0].Coins[0].Denom, "bcoin")
	require.Equal(t, genesisState.Accounts[0].Coins[1].Denom, "acoin")
	require.Equal(t, genesisState.Accounts[1].Address, addr2)
	genesisState.Sanitize()
	require.False(t, genesisState.Accounts[0].AccountNumber > genesisState.Accounts[1].AccountNumber)
	require.Equal(t, genesisState.Accounts[1].Address, addr1)
	require.Equal(t, genesisState.Accounts[1].Coins[0].Denom, "acoin")
	require.Equal(t, genesisState.Accounts[1].Coins[1].Denom, "bcoin")
}
