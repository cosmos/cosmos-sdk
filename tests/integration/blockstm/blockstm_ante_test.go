package blockstm_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/baseapp/txnrunner"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	blockSTMAnteChainID       = "blockstm-ante-test"
	blockSTMAnteSignerCount   = 4
	blockSTMAnteInitialTokens = int64(1_000)
)

type blockSTMAnteAccount struct {
	priv cryptotypes.PrivKey
	addr sdk.AccAddress
}

type blockSTMAnteApp struct {
	app    *simapp.SimApp
	logBuf *bytes.Buffer
}

type blockSTMAnteTx struct {
	signer int
	seq    uint64
	msg    sdk.Msg
}

func TestBlockSTM_AnteHandlerAccountCreationDeterminism(t *testing.T) {
	accounts := newBlockSTMAnteAccounts(blockSTMAnteSignerCount)
	genesisState, valSet, _ := buildBlockSTMAnteGenesis(t, accounts)

	var logBuf bytes.Buffer
	regularApp := newBlockSTMAnteApp(t, genesisState, valSet, false, nil)
	blockSTMApp := newBlockSTMAnteApp(t, genesisState, valSet, true, &logBuf)

	txs := make([]blockSTMAnteTx, 0, len(accounts))
	for i, account := range accounts {
		recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		txs = append(txs, blockSTMAnteTx{
			signer: i,
			seq:    0,
			msg: banktypes.NewMsgSend(
				account.addr,
				recipient,
				sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)),
			),
		})
	}

	txBytes := buildAnteTxBytes(t, regularApp.app, accounts, txs)
	compareAnteExecution(t, regularApp.app, blockSTMApp.app, txBytes)

	logOutput := logBuf.String()
	panicCount := strings.Count(logOutput, "panic recovered in runTx")
	uniquenessViolationCount := strings.Count(logOutput, "uniqueness constraint violation")
	t.Logf("Ante account creation logs: %d recovered panics, %d uniqueness violations", panicCount, uniquenessViolationCount)

	require.Zero(t, panicCount, "expected no recovered panics during signed ante+execution account creation")
}

func TestBlockSTM_AnteHandlerSequenceAndStakingDeterminism(t *testing.T) {
	accounts := newBlockSTMAnteAccounts(3)
	genesisState, valSet, validatorAddr := buildBlockSTMAnteGenesis(t, accounts)

	regularApp := newBlockSTMAnteApp(t, genesisState, valSet, false, nil)
	blockSTMApp := newBlockSTMAnteApp(t, genesisState, valSet, true, nil)

	freshRecipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	txs := []blockSTMAnteTx{
		{
			signer: 0,
			seq:    0,
			msg: banktypes.NewMsgSend(
				accounts[0].addr,
				freshRecipient,
				sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 25)),
			),
		},
		{
			signer: 1,
			seq:    0,
			msg: stakingtypes.NewMsgDelegate(
				accounts[1].addr.String(),
				validatorAddr.String(),
				sdk.NewInt64Coin(sdk.DefaultBondDenom, 40),
			),
		},
		{
			signer: 0,
			seq:    1,
			msg: stakingtypes.NewMsgDelegate(
				accounts[0].addr.String(),
				validatorAddr.String(),
				sdk.NewInt64Coin(sdk.DefaultBondDenom, 15),
			),
		},
		{
			signer: 0,
			seq:    2,
			msg: banktypes.NewMsgSend(
				accounts[0].addr,
				accounts[2].addr,
				sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)),
			),
		},
		{
			signer: 1,
			seq:    1,
			msg: banktypes.NewMsgSend(
				accounts[1].addr,
				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
				sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 7)),
			),
		},
		{
			signer: 2,
			seq:    0,
			msg: stakingtypes.NewMsgDelegate(
				accounts[2].addr.String(),
				validatorAddr.String(),
				sdk.NewInt64Coin(sdk.DefaultBondDenom, 20),
			),
		},
	}

	txBytes := buildAnteTxBytes(t, regularApp.app, accounts, txs)
	compareAnteExecution(t, regularApp.app, blockSTMApp.app, txBytes)
}

func TestBlockSTM_AnteHandlerFailureDeterminism(t *testing.T) {
	accounts := newBlockSTMAnteAccounts(2)
	genesisState, valSet, validatorAddr := buildBlockSTMAnteGenesis(t, accounts)

	regularApp := newBlockSTMAnteApp(t, genesisState, valSet, false, nil)
	blockSTMApp := newBlockSTMAnteApp(t, genesisState, valSet, true, nil)

	txs := []blockSTMAnteTx{
		{
			signer: 0,
			seq:    0,
			msg: banktypes.NewMsgSend(
				accounts[0].addr,
				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
				sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 980)),
			),
		},
		{
			signer: 0,
			seq:    1,
			msg: stakingtypes.NewMsgDelegate(
				accounts[0].addr.String(),
				validatorAddr.String(),
				sdk.NewInt64Coin(sdk.DefaultBondDenom, 50),
			),
		},
	}

	txBytes := buildAnteTxBytes(t, regularApp.app, accounts, txs)
	regularRes, blockSTMRes := compareAnteExecution(t, regularApp.app, blockSTMApp.app, txBytes)

	require.Equal(t, uint32(0), regularRes.TxResults[0].Code)
	require.NotEqual(t, uint32(0), regularRes.TxResults[1].Code)
	require.Equal(t, regularRes.TxResults, blockSTMRes.TxResults)
}

func newBlockSTMAnteAccounts(count int) []blockSTMAnteAccount {
	accounts := make([]blockSTMAnteAccount, count)
	for i := range count {
		priv := secp256k1.GenPrivKey()
		accounts[i] = blockSTMAnteAccount{
			priv: priv,
			addr: sdk.AccAddress(priv.PubKey().Address()),
		}
	}

	return accounts
}

func buildBlockSTMAnteGenesis(
	t *testing.T,
	accounts []blockSTMAnteAccount,
) ([]byte, *cmttypes.ValidatorSet, sdk.ValAddress) {
	t.Helper()

	templateApp := simapp.NewSimApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		true,
		simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
		baseapp.SetChainID(blockSTMAnteChainID),
	)

	valSet, err := simtestutil.CreateRandomValidatorSet()
	require.NoError(t, err)

	genAccs := make([]authtypes.GenesisAccount, 0, len(accounts))
	balances := make([]banktypes.Balance, 0, len(accounts))
	for _, account := range accounts {
		genAccs = append(genAccs, authtypes.NewBaseAccount(account.addr, account.priv.PubKey(), 0, 0))
		balances = append(balances, banktypes.Balance{
			Address: account.addr.String(),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, blockSTMAnteInitialTokens)),
		})
	}

	genesisState, err := simtestutil.GenesisStateWithValSet(
		templateApp.AppCodec(),
		templateApp.DefaultGenesis(),
		valSet,
		genAccs,
		balances...,
	)
	require.NoError(t, err)

	stateBytes, err := json.Marshal(genesisState)
	require.NoError(t, err)

	return stateBytes, valSet, sdk.ValAddress(valSet.Validators[0].Address)
}

func newBlockSTMAnteApp(
	t *testing.T,
	genesisState []byte,
	valSet *cmttypes.ValidatorSet,
	enableBlockSTM bool,
	logBuf *bytes.Buffer,
) blockSTMAnteApp {
	t.Helper()

	logger := log.Logger(log.NewNopLogger())
	if logBuf != nil {
		logger = log.NewLogger(logBuf, log.OutputJSONOption())
	}

	app := simapp.NewSimApp(
		logger,
		dbm.NewMemDB(),
		true,
		simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
		baseapp.SetChainID(blockSTMAnteChainID),
	)

	if enableBlockSTM {
		app.SetBlockSTMTxRunner(txnrunner.NewSTMRunner(
			app.TxConfig().TxDecoder(),
			app.GetStoreKeys(),
			8,
			false,
			func(_ storetypes.MultiStore) string { return sdk.DefaultBondDenom },
		))
	}

	_, err := app.InitChain(&abci.RequestInitChain{
		ChainId:         blockSTMAnteChainID,
		ConsensusParams: simtestutil.DefaultConsensusParams,
		AppStateBytes:   genesisState,
	})
	require.NoError(t, err)

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Hash:               app.LastCommitID().Hash,
		NextValidatorsHash: valSet.Hash(),
	})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	return blockSTMAnteApp{
		app:    app,
		logBuf: logBuf,
	}
}

func buildAnteTxBytes(
	t *testing.T,
	app *simapp.SimApp,
	accounts []blockSTMAnteAccount,
	txs []blockSTMAnteTx,
) [][]byte {
	t.Helper()

	ctx := app.NewContext(true)
	accountNums := make([]uint64, len(accounts))
	for i, account := range accounts {
		acc := app.AccountKeeper.GetAccount(ctx, account.addr)
		require.NotNil(t, acc)
		accountNums[i] = acc.GetAccountNumber()
	}

	txBytes := make([][]byte, 0, len(txs))
	for _, tx := range txs {
		txBytes = append(txBytes, buildSignedAnteTxBytes(
			t,
			app.TxConfig(),
			tx.msg,
			blockSTMAnteChainID,
			accountNums[tx.signer],
			tx.seq,
			accounts[tx.signer].priv,
		))
	}

	return txBytes
}

func buildSignedAnteTxBytes(
	t *testing.T,
	txConfig client.TxConfig,
	msg sdk.Msg,
	chainID string,
	accountNumber uint64,
	sequence uint64,
	priv cryptotypes.PrivKey,
) []byte {
	t.Helper()

	signMode, err := authsign.APISignModeToInternal(txConfig.SignModeHandler().DefaultMode())
	require.NoError(t, err)

	sig := txsigning.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &txsigning.SingleSignatureData{
			SignMode: signMode,
		},
		Sequence: sequence,
	}

	txBuilder := txConfig.NewTxBuilder()
	require.NoError(t, txBuilder.SetMsgs(msg))
	txBuilder.SetFeeAmount(sdk.NewCoins())
	txBuilder.SetGasLimit(simtestutil.DefaultGenTxGas)
	require.NoError(t, txBuilder.SetSignatures(sig))

	signerData := authsign.SignerData{
		Address:       sdk.AccAddress(priv.PubKey().Address()).String(),
		ChainID:       chainID,
		AccountNumber: accountNumber,
		Sequence:      sequence,
		PubKey:        priv.PubKey(),
	}

	signBytes, err := authsign.GetSignBytesAdapter(
		context.Background(),
		txConfig.SignModeHandler(),
		signMode,
		signerData,
		txBuilder.GetTx(),
	)
	require.NoError(t, err)

	signature, err := priv.Sign(signBytes)
	require.NoError(t, err)

	sig.Data.(*txsigning.SingleSignatureData).Signature = signature
	require.NoError(t, txBuilder.SetSignatures(sig))

	bz, err := txConfig.TxEncoder()(txBuilder.GetTx())
	require.NoError(t, err)

	return bz
}

func compareAnteExecution(
	t *testing.T,
	regularApp *simapp.SimApp,
	blockSTMApp *simapp.SimApp,
	txBytes [][]byte,
) (*abci.ResponseFinalizeBlock, *abci.ResponseFinalizeBlock) {
	t.Helper()

	execHeight := regularApp.LastBlockHeight() + 1

	regularRes, err := regularApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: execHeight,
		Hash:   regularApp.LastCommitID().Hash,
		Txs:    txBytes,
	})
	require.NoError(t, err)

	blockSTMRes, err := blockSTMApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: execHeight,
		Hash:   blockSTMApp.LastCommitID().Hash,
		Txs:    txBytes,
	})
	require.NoError(t, err)

	require.Equal(t, regularRes.TxResults, blockSTMRes.TxResults)
	require.Equal(t, regularRes.AppHash, blockSTMRes.AppHash)

	_, err = regularApp.Commit()
	require.NoError(t, err)
	_, err = blockSTMApp.Commit()
	require.NoError(t, err)

	require.Equal(t, regularApp.LastCommitID(), blockSTMApp.LastCommitID())

	return regularRes, blockSTMRes
}
