package simapp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// SimAppChainID hardcoded chainID for simulation
const (
	DefaultGenTxGas = 1000000
	SimAppChainID   = "simulation-app"
)

// DefaultConsensusParams defines the default Tendermint consensus params used in
// SimApp testing.
var DefaultConsensusParams = &abci.ConsensusParams{
	Block: &abci.BlockParams{
		MaxBytes: 200000,
		MaxGas:   2000000,
	},
	Evidence: &abci.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  1814400,
	},
	Validator: &abci.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
			tmtypes.ABCIPubKeyTypeSecp256k1,
		},
	},
}

// Setup initializes a new SimApp. A Nop logger is set in SimApp.
func Setup(isCheckTx bool) *SimApp {
	db := dbm.NewMemDB()
	app := NewSimApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5)
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		genesisState := NewDefaultGenesisState()
		stateBytes, err := codec.MarshalJSONIndent(app.Codec(), genesisState)
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		app.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}

// SetupWithGenesisValSet initializes a new SimApp with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit (10^6) in the default token of the simapp from first genesis
// account. A Nop logger is set in SimApp.
func SetupWithGenesisValSet(t *testing.T, valSet *tmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *SimApp {
	db := dbm.NewMemDB()
	app := NewSimApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5)

	genesisState := NewDefaultGenesisState()

	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.Codec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.NewInt(1000000)

	for _, val := range valSet.Validators {
		validator := stakingtypes.Validator{
			OperatorAddress:   val.Address.Bytes(),
			ConsensusPubkey:   sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, val.PubKey),
			Jailed:            false,
			Status:            sdk.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdk.OneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			MinSelfDelegation: sdk.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress(), val.Address.Bytes(), sdk.OneDec()))

	}

	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.Codec().MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens and delegated tokens to total supply
		totalSupply = totalSupply.Add(b.Coins.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))...)
	}

	// update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply)
	genesisState[banktypes.ModuleName] = app.Codec().MustMarshalJSON(bankGenesis)

	stateBytes, err := codec.MarshalJSONIndent(app.Codec(), genesisState)
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	// commit genesis changes
	app.Commit()
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	return app
}

// SetupWithGenesisAccounts initializes a new SimApp with the provided genesis
// accounts and possible balances.
func SetupWithGenesisAccounts(genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *SimApp {
	db := dbm.NewMemDB()
	app := NewSimApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0)

	// initialize the chain with the passed in genesis accounts
	genesisState := NewDefaultGenesisState()

	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.Codec().MustMarshalJSON(authGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		totalSupply = totalSupply.Add(b.Coins...)
	}

	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply)
	genesisState[banktypes.ModuleName] = app.Codec().MustMarshalJSON(bankGenesis)

	stateBytes, err := codec.MarshalJSONIndent(app.Codec(), genesisState)
	if err != nil {
		panic(err)
	}

	app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	app.Commit()
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: app.LastBlockHeight() + 1}})

	return app
}

type GenerateAccountStrategy func(int) []sdk.AccAddress

// createRandomAccounts is a strategy used by addTestAddrs() in order to generated addresses in random order.
func createRandomAccounts(accNum int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, accNum)
	for i := 0; i < accNum; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}

// createIncrementalAccounts is a strategy used by addTestAddrs() in order to generated addresses in ascending order.
func createIncrementalAccounts(accNum int) []sdk.AccAddress {
	var addresses []sdk.AccAddress
	var buffer bytes.Buffer

	// start at 100 so we can make up to 999 test addresses with valid test addresses
	for i := 100; i < (accNum + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6") //base address string

		buffer.WriteString(numString) //adding on final two digits to make addresses unique
		res, _ := sdk.AccAddressFromHex(buffer.String())
		bech := res.String()
		addr, _ := TestAddr(buffer.String(), bech)

		addresses = append(addresses, addr)
		buffer.Reset()
	}

	return addresses
}

// AddTestAddrsFromPubKeys adds the addresses into the SimApp providing only the public keys.
func AddTestAddrsFromPubKeys(app *SimApp, ctx sdk.Context, pubKeys []crypto.PubKey, accAmt sdk.Int) {
	initCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), accAmt))

	setTotalSupply(app, ctx, accAmt, len(pubKeys))

	// fill all the addresses with some coins, set the loose pool tokens simultaneously
	for _, pubKey := range pubKeys {
		saveAccount(app, ctx, sdk.AccAddress(pubKey.Address()), initCoins)
	}
}

// setTotalSupply provides the total supply based on accAmt * totalAccounts.
func setTotalSupply(app *SimApp, ctx sdk.Context, accAmt sdk.Int, totalAccounts int) {
	totalSupply := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), accAmt.MulRaw(int64(totalAccounts))))
	prevSupply := app.BankKeeper.GetSupply(ctx)
	app.BankKeeper.SetSupply(ctx, banktypes.NewSupply(prevSupply.GetTotal().Add(totalSupply...)))
}

// AddTestAddrs constructs and returns accNum amount of accounts with an
// initial balance of accAmt in random order
func AddTestAddrs(app *SimApp, ctx sdk.Context, accNum int, accAmt sdk.Int) []sdk.AccAddress {
	return addTestAddrs(app, ctx, accNum, accAmt, createRandomAccounts)
}

// AddTestAddrs constructs and returns accNum amount of accounts with an
// initial balance of accAmt in random order
func AddTestAddrsIncremental(app *SimApp, ctx sdk.Context, accNum int, accAmt sdk.Int) []sdk.AccAddress {
	return addTestAddrs(app, ctx, accNum, accAmt, createIncrementalAccounts)
}

func addTestAddrs(app *SimApp, ctx sdk.Context, accNum int, accAmt sdk.Int, strategy GenerateAccountStrategy) []sdk.AccAddress {
	testAddrs := strategy(accNum)

	initCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), accAmt))
	setTotalSupply(app, ctx, accAmt, accNum)

	// fill all the addresses with some coins, set the loose pool tokens simultaneously
	for _, addr := range testAddrs {
		saveAccount(app, ctx, addr, initCoins)
	}

	return testAddrs
}

// saveAccount saves the provided account into the simapp with balance based on initCoins.
func saveAccount(app *SimApp, ctx sdk.Context, addr sdk.AccAddress, initCoins sdk.Coins) {
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acc)
	_, err := app.BankKeeper.AddCoins(ctx, addr, initCoins)
	if err != nil {
		panic(err)
	}
}

// ConvertAddrsToValAddrs converts the provided addresses to ValAddress.
func ConvertAddrsToValAddrs(addrs []sdk.AccAddress) []sdk.ValAddress {
	valAddrs := make([]sdk.ValAddress, len(addrs))

	for i, addr := range addrs {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	return valAddrs
}

func TestAddr(addr string, bech string) (sdk.AccAddress, error) {
	res, err := sdk.AccAddressFromHex(addr)
	if err != nil {
		return nil, err
	}
	bechexpected := res.String()
	if bech != bechexpected {
		return nil, fmt.Errorf("bech encoding doesn't match reference")
	}

	bechres, err := sdk.AccAddressFromBech32(bech)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(bechres, res) {
		return nil, err
	}

	return res, nil
}

// CheckBalance checks the balance of an account.
func CheckBalance(t *testing.T, app *SimApp, addr sdk.AccAddress, balances sdk.Coins) {
	ctxCheck := app.BaseApp.NewContext(true, abci.Header{})
	require.True(t, balances.IsEqual(app.BankKeeper.GetAllBalances(ctxCheck, addr)))
}

// GenTx generates a signed mock transaction.
func GenTx(gen client.TxGenerator, msgs []sdk.Msg, feeAmt sdk.Coins, gas uint64, chainID string, accnums []uint64, seq []uint64, priv ...crypto.PrivKey) (sdk.Tx, error) {
	sigs := make([]signing.SignatureV2, len(priv))

	// create a random length memo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))

	signMode := gen.SignModeHandler().DefaultMode()

	for i, p := range priv {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
		}
	}

	tx := gen.NewTxBuilder()
	err := tx.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = tx.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	tx.SetMemo(memo)
	tx.SetFeeAmount(feeAmt)
	tx.SetGasLimit(gas)
	for i, p := range priv {
		// use a empty chainID for ease of testing
		signerData := authsign.SignerData{
			ChainID:         chainID,
			AccountNumber:   accnums[i],
			AccountSequence: seq[i],
		}
		signBytes, err := gen.SignModeHandler().GetSignBytes(signMode, signerData, tx.GetTx())
		if err != nil {
			panic(err)
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			panic(err)
		}
		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
		err = tx.SetSignatures(sigs...)
		if err != nil {
			panic(err)
		}
	}

	return tx.GetTx(), nil
}

// SignCheckDeliver checks a generated signed transaction and simulates a
// block commitment with the given transaction. A test assertion is made using
// the parameter 'expPass' against the result. A corresponding result is
// returned.
func SignCheckDeliver(
	t *testing.T, txGen client.TxGenerator, app *bam.BaseApp, header abci.Header, msgs []sdk.Msg,
	accNums, seq []uint64, expSimPass, expPass bool, priv ...crypto.PrivKey,
) (sdk.GasInfo, *sdk.Result, error) {

	tx, err := GenTx(
		txGen,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
		DefaultGenTxGas,
		"",
		accNums,
		seq,
		priv...,
	)
	require.NoError(t, err)
	txBytes, err := txGen.TxEncoder()(tx)
	require.Nil(t, err)

	// Must simulate now as CheckTx doesn't run Msgs anymore
	_, res, err := app.Simulate(txBytes, tx)

	if expSimPass {
		require.NoError(t, err)
		require.NotNil(t, res)
	} else {
		require.Error(t, err)
		require.Nil(t, res)
	}

	// Simulate a sending a transaction and committing a block
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	gInfo, res, err := app.Deliver(tx)

	if expPass {
		require.NoError(t, err)
		require.NotNil(t, res)
	} else {
		require.Error(t, err)
		require.Nil(t, res)
	}

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	return gInfo, res, err
}

// GenSequenceOfTxs generates a set of signed transactions of messages, such
// that they differ only by having the sequence numbers incremented between
// every transaction.
func GenSequenceOfTxs(txGen client.TxGenerator, msgs []sdk.Msg, accNums []uint64, initSeqNums []uint64, numToGenerate int, priv ...crypto.PrivKey) ([]sdk.Tx, error) {
	txs := make([]sdk.Tx, numToGenerate)
	var err error
	for i := 0; i < numToGenerate; i++ {
		txs[i], err = GenTx(
			txGen,
			msgs,
			sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
			DefaultGenTxGas,
			"",
			accNums,
			initSeqNums,
			priv...,
		)
		if err != nil {
			break
		}
		incrementAllSequenceNumbers(initSeqNums)
	}

	return txs, err
}

func incrementAllSequenceNumbers(initSeqNums []uint64) {
	for i := 0; i < len(initSeqNums); i++ {
		initSeqNums[i]++
	}
}

// CreateTestPubKeys returns a total of numPubKeys public keys in ascending order.
func CreateTestPubKeys(numPubKeys int) []crypto.PubKey {
	var publicKeys []crypto.PubKey
	var buffer bytes.Buffer

	// start at 10 to avoid changing 1 to 01, 2 to 02, etc
	for i := 100; i < (numPubKeys + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AF") // base pubkey string
		buffer.WriteString(numString)                                                       // adding on final two digits to make pubkeys unique
		publicKeys = append(publicKeys, NewPubKeyFromHex(buffer.String()))
		buffer.Reset()
	}

	return publicKeys
}

// NewPubKeyFromHex returns a PubKey from a hex string.
func NewPubKeyFromHex(pk string) (res crypto.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	var pkEd ed25519.PubKeyEd25519
	copy(pkEd[:], pkBytes)
	return pkEd
}
