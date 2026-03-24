package blockstm_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/math"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/baseapp/txnrunner"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	testChainID           = "blockstm-fuzz-test"
	accsCount             = 4
	initialValidatorCount = 2
	initialBalance        = int64(1_000)
	selfDelegation        = int64(100)
	maximumOperations     = 30
)

type account struct {
	priv cryptotypes.PrivKey
	addr sdk.AccAddress
}

type testApplication struct {
	app      *simapp.SimApp
	txConfig client.TxConfig
}

type operationKind string

const (
	opCreateValidator operationKind = "create-validator"
	opSendCoins       operationKind = "send"
	opDelegate        operationKind = "delegate"
	opUndelegate      operationKind = "undelegate"
	opRedelegate      operationKind = "redelegate"
	opCancelUnbonding operationKind = "cancel-unbonding"
)

type operation struct {
	kind         operationKind
	account      int
	to           int
	validator    int
	dstValidator int
	amount       int64
}

type state struct {
	balances              []int64
	validatorExists       []bool
	delegations           [][]int64
	unbondings            [][]int64
	receivingRedelegation [][]bool
}

type delegationRef struct {
	delegator int
	validator int
}

func TestBlockSTMDeterminism(t *testing.T) {
	accounts := generateAccounts(accsCount)
	accountAddrs := getAddrs(accounts)
	validatorPubKeys := generateValidatorPubKeys(accsCount)
	genesisState, valSet := buildGenesisState(t, accounts)

	rapid.Check(t, func(rt *rapid.T) {
		ops := generateOperations(rt, newState(
			len(accountAddrs),
			initialValidatorCount,
			initialBalance,
			selfDelegation,
		), maximumOperations)
		runOperations(t, genesisState, valSet, accounts, accountAddrs, validatorPubKeys, ops, 8)
	})
}

func runOperations(
	t *testing.T,
	genesisState []byte,
	valSet *cmttypes.ValidatorSet,
	accounts []account,
	accountAddrs []sdk.AccAddress,
	validatorPubKeys []cryptotypes.PubKey,
	ops []operation,
	blockSTMExecutors int,
) {
	t.Helper()

	regularApp := newTestApplication(t, dbm.NewMemDB(), genesisState, valSet, false, 0)
	blockSTMApp := newTestApplication(t, dbm.NewMemDB(), genesisState, valSet, true, blockSTMExecutors)

	initTestApplication(t, regularApp, accounts, accountAddrs, validatorPubKeys, initialValidatorCount)
	initTestApplication(t, blockSTMApp, accounts, accountAddrs, validatorPubKeys, initialValidatorCount)

	require.Equal(t, regularApp.app.LastCommitID(), blockSTMApp.app.LastCommitID())

	execHeight := regularApp.app.LastBlockHeight() + 1
	txBytes := buildTxs(t, regularApp.app, accounts, accountAddrs, validatorPubKeys, execHeight, ops)

	regularRes, err := regularApp.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: execHeight,
		Hash:   regularApp.app.LastCommitID().Hash,
		Txs:    txBytes,
	})
	require.NoError(t, err)

	blockSTMRes, err := blockSTMApp.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: execHeight,
		Hash:   blockSTMApp.app.LastCommitID().Hash,
		Txs:    txBytes,
	})
	require.NoError(t, err)

	require.Equal(t, regularRes.TxResults, blockSTMRes.TxResults)
	require.Equal(t, regularRes.AppHash, blockSTMRes.AppHash)

	_, err = regularApp.app.Commit()
	require.NoError(t, err)
	_, err = blockSTMApp.app.Commit()
	require.NoError(t, err)

	require.Equal(t, regularApp.app.LastCommitID(), blockSTMApp.app.LastCommitID())
}

func buildGenesisState(t *testing.T, accounts []account) ([]byte, *cmttypes.ValidatorSet) {
	t.Helper()

	templateApp := simapp.NewSimApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		true,
		simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
		baseapp.SetChainID(testChainID),
	)

	valSet, err := simtestutil.CreateRandomValidatorSet()
	require.NoError(t, err)

	genAccs := make([]authtypes.GenesisAccount, 0, len(accounts))
	balances := make([]banktypes.Balance, 0, len(accounts))
	for _, account := range accounts {
		genAccs = append(genAccs, authtypes.NewBaseAccount(account.addr, account.priv.PubKey(), 0, 0))
		balances = append(balances, banktypes.Balance{
			Address: account.addr.String(),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBalance)),
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

	return stateBytes, valSet
}

func newTestApplication(
	t *testing.T,
	db dbm.DB,
	genesisState []byte,
	valSet *cmttypes.ValidatorSet,
	enableBlockSTM bool,
	blockSTMExecutors int,
) testApplication {
	t.Helper()

	logger := log.NewNopLogger()

	app := simapp.NewSimApp(
		logger,
		db,
		true,
		simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
		baseapp.SetChainID(testChainID),
	)

	if enableBlockSTM {
		app.SetBlockSTMTxRunner(txnrunner.NewSTMRunner(
			app.TxConfig().TxDecoder(),
			app.GetStoreKeys(),
			blockSTMExecutors,
			false,
			func(_ storetypes.MultiStore) string { return sdk.DefaultBondDenom },
		))
	}

	_, err := app.InitChain(&abci.RequestInitChain{
		ChainId:         testChainID,
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

	return testApplication{
		app:      app,
		txConfig: app.TxConfig(),
	}
}

func initTestApplication(
	t *testing.T,
	testApp testApplication,
	accounts []account,
	accountAddrs []sdk.AccAddress,
	validatorPubKeys []cryptotypes.PubKey,
	initialValidatorCount int,
) {
	t.Helper()

	bootstrapOps := make([]operation, 0, initialValidatorCount)
	for i := range initialValidatorCount {
		bootstrapOps = append(bootstrapOps, operation{
			kind:    opCreateValidator,
			account: i,
			amount:  selfDelegation,
		})
	}

	bootstrapHeight := testApp.app.LastBlockHeight() + 1
	txBytes := buildTxs(t, testApp.app, accounts, accountAddrs, validatorPubKeys, bootstrapHeight, bootstrapOps)

	res, err := testApp.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: bootstrapHeight,
		Hash:   testApp.app.LastCommitID().Hash,
		Txs:    txBytes,
	})
	require.NoError(t, err)
	requireSuccessfulTxResults(t, res.TxResults)

	_, err = testApp.app.Commit()
	require.NoError(t, err)
}

func generateOperations(rt *rapid.T, s state, maxOps int) []operation {
	opCount := rapid.IntRange(1, maxOps).Draw(rt, "op-count")
	ops := make([]operation, 0, opCount)

	for i := range opCount {
		kinds := availableOps(s)
		kind := rapid.SampledFrom(kinds).Draw(rt, opLabel(i, "kind"))

		switch kind {
		case opCreateValidator:
			candidates := createValidatorCandidates(s)
			account := rapid.SampledFrom(candidates).Draw(rt, opLabel(i, "account"))
			amount := rapid.Int64Range(1, s.balances[account]).Draw(rt, opLabel(i, "amount"))

			s.balances[account] -= amount
			s.validatorExists[account] = true
			s.delegations[account][account] += amount

			ops = append(ops, operation{kind: kind, account: account, amount: amount})

		case opSendCoins:
			senders := accountsWithPositiveAmount(s.balances)
			account := rapid.SampledFrom(senders).Draw(rt, opLabel(i, "account"))

			toChoices := make([]int, 0, len(s.balances)-1)
			for candidate := range s.balances {
				if candidate != account {
					toChoices = append(toChoices, candidate)
				}
			}
			to := rapid.SampledFrom(toChoices).Draw(rt, opLabel(i, "to"))
			amount := rapid.Int64Range(1, s.balances[account]).Draw(rt, opLabel(i, "amount"))

			s.balances[account] -= amount
			s.balances[to] += amount

			ops = append(ops, operation{kind: kind, account: account, to: to, amount: amount})

		case opDelegate:
			delegators := accountsWithPositiveAmount(s.balances)
			account := rapid.SampledFrom(delegators).Draw(rt, opLabel(i, "account"))
			validators := existingValidators(s.validatorExists)
			validator := rapid.SampledFrom(validators).Draw(rt, opLabel(i, "validator"))
			amount := rapid.Int64Range(1, s.balances[account]).Draw(rt, opLabel(i, "amount"))

			s.balances[account] -= amount
			s.delegations[account][validator] += amount

			ops = append(ops, operation{kind: kind, account: account, validator: validator, amount: amount})

		case opUndelegate:
			delegations := positiveDelegations(s.delegations)
			ref := rapid.SampledFrom(delegations).Draw(rt, opLabel(i, "delegation"))
			amount := rapid.Int64Range(1, s.delegations[ref.delegator][ref.validator]).Draw(rt, opLabel(i, "amount"))

			s.delegations[ref.delegator][ref.validator] -= amount
			s.unbondings[ref.delegator][ref.validator] += amount

			ops = append(ops, operation{
				kind:      kind,
				account:   ref.delegator,
				validator: ref.validator,
				amount:    amount,
			})

		case opRedelegate:
			sources := redelegationSources(s)
			ref := rapid.SampledFrom(sources).Draw(rt, opLabel(i, "delegation"))
			dstChoices := make([]int, 0, len(s.validatorExists)-1)
			for _, validator := range existingValidators(s.validatorExists) {
				if validator != ref.validator {
					dstChoices = append(dstChoices, validator)
				}
			}
			dstValidator := rapid.SampledFrom(dstChoices).Draw(rt, opLabel(i, "dst-validator"))
			amount := rapid.Int64Range(1, s.delegations[ref.delegator][ref.validator]).Draw(rt, opLabel(i, "amount"))

			s.delegations[ref.delegator][ref.validator] -= amount
			s.delegations[ref.delegator][dstValidator] += amount
			s.receivingRedelegation[ref.delegator][dstValidator] = true

			ops = append(ops, operation{
				kind:         kind,
				account:      ref.delegator,
				validator:    ref.validator,
				dstValidator: dstValidator,
				amount:       amount,
			})

		case opCancelUnbonding:
			unbondings := positiveDelegations(s.unbondings)
			ref := rapid.SampledFrom(unbondings).Draw(rt, opLabel(i, "unbonding"))
			amount := rapid.Int64Range(1, s.unbondings[ref.delegator][ref.validator]).Draw(rt, opLabel(i, "amount"))

			s.unbondings[ref.delegator][ref.validator] -= amount
			s.delegations[ref.delegator][ref.validator] += amount

			ops = append(ops, operation{
				kind:      kind,
				account:   ref.delegator,
				validator: ref.validator,
				amount:    amount,
			})

		default:
			rt.Fatalf("unsupported operation kind %q", kind)
		}
	}

	return ops
}

func availableOps(s state) []operationKind {
	kinds := make([]operationKind, 0, 6)

	if len(createValidatorCandidates(s)) > 0 {
		kinds = append(kinds, opCreateValidator)
	}

	if len(accountsWithPositiveAmount(s.balances)) > 0 {
		kinds = append(kinds, opSendCoins)
		if len(existingValidators(s.validatorExists)) > 0 {
			kinds = append(kinds, opDelegate)
		}
	}

	if len(positiveDelegations(s.delegations)) > 0 {
		kinds = append(kinds, opUndelegate)
	}

	if len(redelegationSources(s)) > 0 {
		kinds = append(kinds, opRedelegate)
	}

	if len(positiveDelegations(s.unbondings)) > 0 {
		kinds = append(kinds, opCancelUnbonding)
	}

	return kinds
}

func newState(numAccounts, initialValidatorCount int, initialBalance, selfDelegation int64) state {
	s := state{
		balances:              make([]int64, numAccounts),
		validatorExists:       make([]bool, numAccounts),
		delegations:           make([][]int64, numAccounts),
		unbondings:            make([][]int64, numAccounts),
		receivingRedelegation: make([][]bool, numAccounts),
	}

	for i := range numAccounts {
		s.balances[i] = initialBalance
		s.delegations[i] = make([]int64, numAccounts)
		s.unbondings[i] = make([]int64, numAccounts)
		s.receivingRedelegation[i] = make([]bool, numAccounts)
	}

	for i := range initialValidatorCount {
		s.balances[i] -= selfDelegation
		s.validatorExists[i] = true
		s.delegations[i][i] = selfDelegation
	}

	return s
}

func accountsWithPositiveAmount(amounts []int64) []int {
	indices := make([]int, 0, len(amounts))
	for i, amount := range amounts {
		if amount > 0 {
			indices = append(indices, i)
		}
	}

	return indices
}

func createValidatorCandidates(s state) []int {
	candidates := make([]int, 0, len(s.balances))
	for i, balance := range s.balances {
		if !s.validatorExists[i] && balance > 0 {
			candidates = append(candidates, i)
		}
	}

	return candidates
}

func existingValidators(validatorExists []bool) []int {
	validators := make([]int, 0, len(validatorExists))
	for i, exists := range validatorExists {
		if exists {
			validators = append(validators, i)
		}
	}

	return validators
}

func positiveDelegations(amounts [][]int64) []delegationRef {
	refs := make([]delegationRef, 0)
	for delegator, perValidator := range amounts {
		for validator, amount := range perValidator {
			if amount > 0 {
				refs = append(refs, delegationRef{
					delegator: delegator,
					validator: validator,
				})
			}
		}
	}

	return refs
}

func redelegationSources(s state) []delegationRef {
	refs := make([]delegationRef, 0)
	validators := existingValidators(s.validatorExists)

	if len(validators) < 2 {
		return refs
	}

	for delegator, perValidator := range s.delegations {
		for validator, amount := range perValidator {
			if amount <= 0 || s.receivingRedelegation[delegator][validator] {
				continue
			}

			for _, candidate := range validators {
				if candidate != validator {
					refs = append(refs, delegationRef{
						delegator: delegator,
						validator: validator,
					})
					break
				}
			}
		}
	}

	return refs
}

func buildTxs(
	t *testing.T,
	app *simapp.SimApp,
	accounts []account,
	accountAddrs []sdk.AccAddress,
	validatorPubKeys []cryptotypes.PubKey,
	height int64,
	ops []operation,
) [][]byte {
	t.Helper()

	ctx := app.NewContext(true)
	accountNumbers := make([]uint64, len(accounts))
	nextSequences := make([]uint64, len(accounts))
	for i, acc := range accounts {
		acc := app.AccountKeeper.GetAccount(ctx, acc.addr)
		require.NotNil(t, acc)
		accountNumbers[i] = acc.GetAccountNumber()
		nextSequences[i] = acc.GetSequence()
	}

	txBytes := make([][]byte, 0, len(ops))
	for i, op := range ops {
		var msg sdk.Msg

		switch op.kind {
		case opCreateValidator:
			description := stakingtypes.Description{Moniker: fmt.Sprintf("validator-%d", op.account)}
			createValidatorMsg, err := stakingtypes.NewMsgCreateValidator(
				sdk.ValAddress(accountAddrs[op.account]).String(),
				validatorPubKeys[op.account],
				sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount),
				description,
				stakingtestutil.ZeroCommission(),
				math.OneInt(),
			)
			require.NoError(t, err)
			msg = createValidatorMsg

		case opSendCoins:
			msg = banktypes.NewMsgSend(
				accountAddrs[op.account],
				accountAddrs[op.to],
				sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount)),
			)

		case opDelegate:
			msg = stakingtypes.NewMsgDelegate(
				accountAddrs[op.account].String(),
				sdk.ValAddress(accountAddrs[op.validator]).String(),
				sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount),
			)

		case opUndelegate:
			msg = stakingtypes.NewMsgUndelegate(
				accountAddrs[op.account].String(),
				sdk.ValAddress(accountAddrs[op.validator]).String(),
				sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount),
			)

		case opRedelegate:
			msg = stakingtypes.NewMsgBeginRedelegate(
				accountAddrs[op.account].String(),
				sdk.ValAddress(accountAddrs[op.validator]).String(),
				sdk.ValAddress(accountAddrs[op.dstValidator]).String(),
				sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount),
			)

		case opCancelUnbonding:
			msg = stakingtypes.NewMsgCancelUnbondingDelegation(
				accountAddrs[op.account].String(),
				sdk.ValAddress(accountAddrs[op.validator]).String(),
				height,
				sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount),
			)

		default:
			t.Fatalf("unsupported operation at index %d: %q", i, op.kind)
		}

		signer := op.account
		txBytes = append(txBytes, buildSignedTx(
			t,
			app.TxConfig(),
			msg,
			testChainID,
			accountNumbers[signer],
			nextSequences[signer],
			accounts[signer].priv,
		))
		nextSequences[signer]++
	}

	return txBytes
}

func buildSignedTx(
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

func opLabel(index int, suffix string) string {
	return fmt.Sprintf("op-%d-%s", index, suffix)
}

func generateAccounts(count int) []account {
	accounts := make([]account, count)
	for i := range count {
		priv := secp256k1.GenPrivKey()
		accounts[i] = account{
			priv: priv,
			addr: sdk.AccAddress(priv.PubKey().Address()),
		}
	}

	return accounts
}

func getAddrs(accounts []account) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, len(accounts))
	for i, acc := range accounts {
		addrs[i] = acc.addr
	}

	return addrs
}

func generateValidatorPubKeys(count int) []cryptotypes.PubKey {
	pubKeys := make([]cryptotypes.PubKey, count)
	for i := range count {
		pubKeys[i] = ed25519.GenPrivKey().PubKey()
	}

	return pubKeys
}
