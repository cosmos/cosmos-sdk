package blockstm_test

import (
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/baseapp/txnrunner"
	"github.com/cosmos/cosmos-sdk/client"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	accountsCount         = 4
	initialValidatorCount = 2
	initialBalance        = int64(1_000)
	selfDelegation        = int64(100)
	maximumOperations     = 30
)

type testApplication struct {
	app           *baseapp.BaseApp
	bankKeeper    bankkeeper.BaseKeeper
	stakingKeeper *stakingkeeper.Keeper
	txConfig      client.TxConfig
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

func TestBlockSTM_MixedMessageDeterminism(t *testing.T) {
	participantAddrs := generateAddrs(accountsCount)
	validatorPubKeys := generateValidatorPubKeys(accountsCount)

	rapid.Check(t, func(rt *rapid.T) {
		regularApp := newTestApplication(t, dbm.NewMemDB(), false)
		blockSTMApp := newTestApplication(t, dbm.NewMemDB(), true)

		initTestApplication(t, regularApp, participantAddrs, validatorPubKeys, initialValidatorCount)
		initTestApplication(t, blockSTMApp, participantAddrs, validatorPubKeys, initialValidatorCount)

		require.Equal(t, regularApp.app.LastCommitID(), blockSTMApp.app.LastCommitID())

		ops := generateOperations(rt, newState(
			len(participantAddrs),
			initialValidatorCount,
			initialBalance,
			selfDelegation,
		), maximumOperations)
		execHeight := regularApp.app.LastBlockHeight() + 1
		txBytes := buildTxs(t, regularApp.txConfig, participantAddrs, validatorPubKeys, execHeight, ops)

		regularRes, err := regularApp.app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: execHeight,
			Txs:    txBytes,
		})
		require.NoError(t, err)

		blockSTMRes, err := blockSTMApp.app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: execHeight,
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
	})
}

func newTestApplication(t *testing.T, db dbm.DB, enableBlockSTM bool) testApplication {
	t.Helper()

	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey)
	encCfg := moduletestutil.MakeTestEncodingConfig(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
	)
	cdc := encCfg.Codec

	bApp := baseapp.NewBaseApp(
		"test",
		log.NewNopLogger(),
		db,
		encCfg.TxConfig.TxDecoder(),
		baseapp.SetChainID("test"),
	)
	bApp.MountKVStores(keys)
	bApp.SetInterfaceRegistry(encCfg.InterfaceRegistry)

	authority := authtypes.NewModuleAddress("gov")

	maccPerms := map[string][]string{
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.ModuleName:        {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		map[string]bool{accountKeeper.GetAuthority(): false},
		authority.String(),
		log.NewNopLogger(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		authority.String(),
		addresscodec.NewBech32Codec(sdk.Bech32PrefixValAddr),
		addresscodec.NewBech32Codec(sdk.Bech32PrefixConsAddr),
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, nil)

	bApp.SetInitChainer(func(ctx sdk.Context, _ *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
		authModule.InitGenesis(ctx, cdc, authModule.DefaultGenesis(cdc))
		bankModule.InitGenesis(ctx, cdc, bankModule.DefaultGenesis(cdc))
		stakingModule.InitGenesis(ctx, cdc, stakingModule.DefaultGenesis(cdc))
		return &abci.ResponseInitChain{}, nil
	})

	bApp.SetBeginBlocker(func(ctx sdk.Context) (sdk.BeginBlock, error) {
		return sdk.BeginBlock{}, stakingModule.BeginBlock(ctx)
	})

	bApp.SetEndBlocker(func(ctx sdk.Context) (sdk.EndBlock, error) {
		if err := bankModule.EndBlock(ctx); err != nil {
			return sdk.EndBlock{}, err
		}

		validatorUpdates, err := stakingModule.EndBlock(ctx)
		if err != nil {
			return sdk.EndBlock{}, err
		}

		return sdk.EndBlock{ValidatorUpdates: validatorUpdates}, nil
	})

	banktypes.RegisterMsgServer(bApp.MsgServiceRouter(), bankkeeper.NewMsgServerImpl(bankKeeper))
	stakingtypes.RegisterMsgServer(bApp.MsgServiceRouter(), stakingkeeper.NewMsgServerImpl(stakingKeeper))

	if enableBlockSTM {
		bApp.SetBlockSTMTxRunner(txnrunner.NewSTMRunner(
			encCfg.TxConfig.TxDecoder(),
			[]storetypes.StoreKey{
				keys[authtypes.StoreKey],
				keys[banktypes.StoreKey],
				keys[stakingtypes.StoreKey],
			},
			8,
			false,
			func(_ storetypes.MultiStore) string { return sdk.DefaultBondDenom },
		))
	}

	return testApplication{
		app:           bApp,
		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,
		txConfig:      encCfg.TxConfig,
	}
}

func initTestApplication(
	t *testing.T,
	testApp testApplication,
	participantAddrs []sdk.AccAddress,
	validatorPubKeys []cryptotypes.PubKey,
	initialValidatorCount int,
) {
	t.Helper()

	require.NoError(t, testApp.app.LoadLatestVersion())

	_, err := testApp.app.InitChain(&abci.RequestInitChain{ChainId: "test"})
	require.NoError(t, err)

	_, err = testApp.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: testApp.app.LastBlockHeight() + 1})
	require.NoError(t, err)

	ctx := testApp.app.NewContext(false)

	for _, addr := range participantAddrs {
		require.NoError(t, banktestutil.FundAccount(
			ctx,
			testApp.bankKeeper,
			addr,
			sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, initialBalance)),
		))
	}

	stakingHelper := stakingtestutil.NewHelper(t, ctx, testApp.stakingKeeper)
	for i := range initialValidatorCount {
		stakingHelper.CreateValidator(
			sdk.ValAddress(participantAddrs[i]),
			validatorPubKeys[i],
			math.NewInt(selfDelegation),
			true,
		)
	}

	_, err = testApp.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: testApp.app.LastBlockHeight() + 1})
	require.NoError(t, err)
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

func newState(participantCount, initialValidatorCount int, initialBalance, selfDelegation int64) state {
	s := state{
		balances:              make([]int64, participantCount),
		validatorExists:       make([]bool, participantCount),
		delegations:           make([][]int64, participantCount),
		unbondings:            make([][]int64, participantCount),
		receivingRedelegation: make([][]bool, participantCount),
	}

	for i := range participantCount {
		s.balances[i] = initialBalance
		s.delegations[i] = make([]int64, participantCount)
		s.unbondings[i] = make([]int64, participantCount)
		s.receivingRedelegation[i] = make([]bool, participantCount)
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
	txConfig client.TxConfig,
	participantAddrs []sdk.AccAddress,
	validatorPubKeys []cryptotypes.PubKey,
	height int64,
	ops []operation,
) [][]byte {
	t.Helper()

	txBytes := make([][]byte, 0, len(ops))
	for i, op := range ops {
		var msg sdk.Msg

		switch op.kind {
		case opCreateValidator:
			description := stakingtypes.Description{Moniker: fmt.Sprintf("validator-%d", op.account)}
			createValidatorMsg, err := stakingtypes.NewMsgCreateValidator(
				sdk.ValAddress(participantAddrs[op.account]).String(),
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
				participantAddrs[op.account],
				participantAddrs[op.to],
				sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount)),
			)

		case opDelegate:
			msg = stakingtypes.NewMsgDelegate(
				participantAddrs[op.account].String(),
				sdk.ValAddress(participantAddrs[op.validator]).String(),
				sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount),
			)

		case opUndelegate:
			msg = stakingtypes.NewMsgUndelegate(
				participantAddrs[op.account].String(),
				sdk.ValAddress(participantAddrs[op.validator]).String(),
				sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount),
			)

		case opRedelegate:
			msg = stakingtypes.NewMsgBeginRedelegate(
				participantAddrs[op.account].String(),
				sdk.ValAddress(participantAddrs[op.validator]).String(),
				sdk.ValAddress(participantAddrs[op.dstValidator]).String(),
				sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount),
			)

		case opCancelUnbonding:
			msg = stakingtypes.NewMsgCancelUnbondingDelegation(
				participantAddrs[op.account].String(),
				sdk.ValAddress(participantAddrs[op.validator]).String(),
				height,
				sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount),
			)

		default:
			t.Fatalf("unsupported operation at index %d: %q", i, op.kind)
		}

		txBuilder := txConfig.NewTxBuilder()
		require.NoError(t, txBuilder.SetMsgs(msg))

		bz, err := txConfig.TxEncoder()(txBuilder.GetTx())
		require.NoError(t, err)

		txBytes = append(txBytes, bz)
	}

	return txBytes
}

func opLabel(index int, suffix string) string {
	return fmt.Sprintf("op-%d-%s", index, suffix)
}

func generateValidatorPubKeys(count int) []cryptotypes.PubKey {
	pubKeys := make([]cryptotypes.PubKey, count)
	for i := range count {
		pubKeys[i] = ed25519.GenPrivKey().PubKey()
	}

	return pubKeys
}
