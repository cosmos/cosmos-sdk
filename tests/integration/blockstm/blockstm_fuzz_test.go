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
	blockSTMMixedParticipantCount      = 4
	blockSTMMixedInitialValidatorCount = 2
	blockSTMMixedInitialBalance        = int64(1_000)
	blockSTMSelfDelegation             = int64(100)
	blockSTMMixedMaxOps                = 30
)

type blockSTMMixedApp struct {
	app           *baseapp.BaseApp
	bankKeeper    bankkeeper.BaseKeeper
	stakingKeeper *stakingkeeper.Keeper
	txConfig      client.TxConfig
}

type blockSTMMixedOpKind string

const (
	blockSTMMixedCreateValidator blockSTMMixedOpKind = "create-validator"
	blockSTMMixedSend            blockSTMMixedOpKind = "send"
	blockSTMMixedDelegate        blockSTMMixedOpKind = "delegate"
	blockSTMMixedUndelegate      blockSTMMixedOpKind = "undelegate"
	blockSTMMixedRedelegate      blockSTMMixedOpKind = "redelegate"
	blockSTMMixedCancelUnbonding blockSTMMixedOpKind = "cancel-unbonding"
)

type blockSTMMixedOp struct {
	kind         blockSTMMixedOpKind
	account      int
	to           int
	validator    int
	dstValidator int
	amount       int64
}

type blockSTMMixedWorld struct {
	balances              []int64
	validatorExists       []bool
	delegations           [][]int64
	unbondings            [][]int64
	receivingRedelegation [][]bool
}

type blockSTMMixedDelegationRef struct {
	delegator int
	validator int
}

func TestBlockSTM_MixedMessageDeterminism(t *testing.T) {
	participantAddrs := generateAddrs(blockSTMMixedParticipantCount)
	validatorPubKeys := generateValidatorPubKeys(blockSTMMixedParticipantCount)

	rapid.Check(t, func(rt *rapid.T) {
		regularApp := newBlockSTMMixedApp(t, dbm.NewMemDB(), false)
		blockSTMApp := newBlockSTMMixedApp(t, dbm.NewMemDB(), true)

		initChainAndBootstrapMixedApp(t, regularApp, participantAddrs, validatorPubKeys, blockSTMMixedInitialValidatorCount)
		initChainAndBootstrapMixedApp(t, blockSTMApp, participantAddrs, validatorPubKeys, blockSTMMixedInitialValidatorCount)

		require.Equal(t, regularApp.app.LastCommitID(), blockSTMApp.app.LastCommitID())

		ops := generateMixedOps(rt, newBlockSTMMixedWorld(
			len(participantAddrs),
			blockSTMMixedInitialValidatorCount,
			blockSTMMixedInitialBalance,
			blockSTMSelfDelegation,
		), blockSTMMixedMaxOps)
		execHeight := regularApp.app.LastBlockHeight() + 1
		txBytes := buildMixedTxs(t, regularApp.txConfig, participantAddrs, validatorPubKeys, execHeight, ops)

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

func newBlockSTMMixedApp(t *testing.T, db dbm.DB, enableBlockSTM bool) blockSTMMixedApp {
	t.Helper()

	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey)
	encCfg := moduletestutil.MakeTestEncodingConfig(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
	)
	cdc := encCfg.Codec

	bApp := baseapp.NewBaseApp(
		"blockstm-mixed-test",
		log.NewNopLogger(),
		db,
		encCfg.TxConfig.TxDecoder(),
		baseapp.SetChainID("blockstm-mixed-test"),
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

	return blockSTMMixedApp{
		app:           bApp,
		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,
		txConfig:      encCfg.TxConfig,
	}
}

func initChainAndBootstrapMixedApp(
	t *testing.T,
	testApp blockSTMMixedApp,
	participantAddrs []sdk.AccAddress,
	validatorPubKeys []cryptotypes.PubKey,
	initialValidatorCount int,
) {
	t.Helper()

	require.NoError(t, testApp.app.LoadLatestVersion())

	_, err := testApp.app.InitChain(&abci.RequestInitChain{ChainId: "blockstm-mixed-test"})
	require.NoError(t, err)

	_, err = testApp.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: testApp.app.LastBlockHeight() + 1})
	require.NoError(t, err)

	ctx := testApp.app.NewContext(false)

	for _, addr := range participantAddrs {
		require.NoError(t, banktestutil.FundAccount(
			ctx,
			testApp.bankKeeper,
			addr,
			sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, blockSTMMixedInitialBalance)),
		))
	}

	stakingHelper := stakingtestutil.NewHelper(t, ctx, testApp.stakingKeeper)
	for i := range initialValidatorCount {
		stakingHelper.CreateValidator(
			sdk.ValAddress(participantAddrs[i]),
			validatorPubKeys[i],
			math.NewInt(blockSTMSelfDelegation),
			true,
		)
	}

	_, err = testApp.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: testApp.app.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = testApp.app.Commit()
	require.NoError(t, err)
}

func generateMixedOps(rt *rapid.T, world blockSTMMixedWorld, maxOps int) []blockSTMMixedOp {
	opCount := rapid.IntRange(1, maxOps).Draw(rt, "op-count")
	ops := make([]blockSTMMixedOp, 0, opCount)

	for i := range opCount {
		kinds := availableMixedOpKinds(world)
		kind := rapid.SampledFrom(kinds).Draw(rt, opLabel(i, "kind"))

		switch kind {
		case blockSTMMixedCreateValidator:
			candidates := createValidatorCandidates(world)
			account := rapid.SampledFrom(candidates).Draw(rt, opLabel(i, "account"))
			amount := rapid.Int64Range(1, world.balances[account]).Draw(rt, opLabel(i, "amount"))

			world.balances[account] -= amount
			world.validatorExists[account] = true
			world.delegations[account][account] += amount

			ops = append(ops, blockSTMMixedOp{kind: kind, account: account, amount: amount})

		case blockSTMMixedSend:
			senders := accountsWithPositiveAmount(world.balances)
			account := rapid.SampledFrom(senders).Draw(rt, opLabel(i, "account"))

			toChoices := make([]int, 0, len(world.balances)-1)
			for candidate := range world.balances {
				if candidate != account {
					toChoices = append(toChoices, candidate)
				}
			}
			to := rapid.SampledFrom(toChoices).Draw(rt, opLabel(i, "to"))
			amount := rapid.Int64Range(1, world.balances[account]).Draw(rt, opLabel(i, "amount"))

			world.balances[account] -= amount
			world.balances[to] += amount

			ops = append(ops, blockSTMMixedOp{kind: kind, account: account, to: to, amount: amount})

		case blockSTMMixedDelegate:
			delegators := accountsWithPositiveAmount(world.balances)
			account := rapid.SampledFrom(delegators).Draw(rt, opLabel(i, "account"))
			validators := existingValidators(world.validatorExists)
			validator := rapid.SampledFrom(validators).Draw(rt, opLabel(i, "validator"))
			amount := rapid.Int64Range(1, world.balances[account]).Draw(rt, opLabel(i, "amount"))

			world.balances[account] -= amount
			world.delegations[account][validator] += amount

			ops = append(ops, blockSTMMixedOp{kind: kind, account: account, validator: validator, amount: amount})

		case blockSTMMixedUndelegate:
			delegations := positiveDelegations(world.delegations)
			ref := rapid.SampledFrom(delegations).Draw(rt, opLabel(i, "delegation"))
			amount := rapid.Int64Range(1, world.delegations[ref.delegator][ref.validator]).Draw(rt, opLabel(i, "amount"))

			world.delegations[ref.delegator][ref.validator] -= amount
			world.unbondings[ref.delegator][ref.validator] += amount

			ops = append(ops, blockSTMMixedOp{
				kind:      kind,
				account:   ref.delegator,
				validator: ref.validator,
				amount:    amount,
			})

		case blockSTMMixedRedelegate:
			sources := redelegationSources(world)
			ref := rapid.SampledFrom(sources).Draw(rt, opLabel(i, "delegation"))
			dstChoices := make([]int, 0, len(world.validatorExists)-1)
			for _, validator := range existingValidators(world.validatorExists) {
				if validator != ref.validator {
					dstChoices = append(dstChoices, validator)
				}
			}
			dstValidator := rapid.SampledFrom(dstChoices).Draw(rt, opLabel(i, "dst-validator"))
			amount := rapid.Int64Range(1, world.delegations[ref.delegator][ref.validator]).Draw(rt, opLabel(i, "amount"))

			world.delegations[ref.delegator][ref.validator] -= amount
			world.delegations[ref.delegator][dstValidator] += amount
			world.receivingRedelegation[ref.delegator][dstValidator] = true

			ops = append(ops, blockSTMMixedOp{
				kind:         kind,
				account:      ref.delegator,
				validator:    ref.validator,
				dstValidator: dstValidator,
				amount:       amount,
			})

		case blockSTMMixedCancelUnbonding:
			unbondings := positiveDelegations(world.unbondings)
			ref := rapid.SampledFrom(unbondings).Draw(rt, opLabel(i, "unbonding"))
			amount := rapid.Int64Range(1, world.unbondings[ref.delegator][ref.validator]).Draw(rt, opLabel(i, "amount"))

			world.unbondings[ref.delegator][ref.validator] -= amount
			world.delegations[ref.delegator][ref.validator] += amount

			ops = append(ops, blockSTMMixedOp{
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

func availableMixedOpKinds(world blockSTMMixedWorld) []blockSTMMixedOpKind {
	kinds := make([]blockSTMMixedOpKind, 0, 6)

	if len(createValidatorCandidates(world)) > 0 {
		kinds = append(kinds, blockSTMMixedCreateValidator)
	}

	if len(accountsWithPositiveAmount(world.balances)) > 0 {
		kinds = append(kinds, blockSTMMixedSend)
		if len(existingValidators(world.validatorExists)) > 0 {
			kinds = append(kinds, blockSTMMixedDelegate)
		}
	}

	if len(positiveDelegations(world.delegations)) > 0 {
		kinds = append(kinds, blockSTMMixedUndelegate)
	}

	if len(redelegationSources(world)) > 0 {
		kinds = append(kinds, blockSTMMixedRedelegate)
	}

	if len(positiveDelegations(world.unbondings)) > 0 {
		kinds = append(kinds, blockSTMMixedCancelUnbonding)
	}

	return kinds
}

func newBlockSTMMixedWorld(participantCount, initialValidatorCount int, initialBalance, selfDelegation int64) blockSTMMixedWorld {
	world := blockSTMMixedWorld{
		balances:              make([]int64, participantCount),
		validatorExists:       make([]bool, participantCount),
		delegations:           make([][]int64, participantCount),
		unbondings:            make([][]int64, participantCount),
		receivingRedelegation: make([][]bool, participantCount),
	}

	for i := range participantCount {
		world.balances[i] = initialBalance
		world.delegations[i] = make([]int64, participantCount)
		world.unbondings[i] = make([]int64, participantCount)
		world.receivingRedelegation[i] = make([]bool, participantCount)
	}

	for i := range initialValidatorCount {
		world.balances[i] -= selfDelegation
		world.validatorExists[i] = true
		world.delegations[i][i] = selfDelegation
	}

	return world
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

func createValidatorCandidates(world blockSTMMixedWorld) []int {
	candidates := make([]int, 0, len(world.balances))
	for i, balance := range world.balances {
		if !world.validatorExists[i] && balance > 0 {
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

func positiveDelegations(amounts [][]int64) []blockSTMMixedDelegationRef {
	refs := make([]blockSTMMixedDelegationRef, 0)
	for delegator, perValidator := range amounts {
		for validator, amount := range perValidator {
			if amount > 0 {
				refs = append(refs, blockSTMMixedDelegationRef{
					delegator: delegator,
					validator: validator,
				})
			}
		}
	}

	return refs
}

func redelegationSources(world blockSTMMixedWorld) []blockSTMMixedDelegationRef {
	refs := make([]blockSTMMixedDelegationRef, 0)
	validators := existingValidators(world.validatorExists)

	if len(validators) < 2 {
		return refs
	}

	for delegator, perValidator := range world.delegations {
		for validator, amount := range perValidator {
			if amount <= 0 || world.receivingRedelegation[delegator][validator] {
				continue
			}

			for _, candidate := range validators {
				if candidate != validator {
					refs = append(refs, blockSTMMixedDelegationRef{
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

func buildMixedTxs(
	t *testing.T,
	txConfig client.TxConfig,
	participantAddrs []sdk.AccAddress,
	validatorPubKeys []cryptotypes.PubKey,
	height int64,
	ops []blockSTMMixedOp,
) [][]byte {
	t.Helper()

	txBytes := make([][]byte, 0, len(ops))
	for i, op := range ops {
		var msg sdk.Msg

		switch op.kind {
		case blockSTMMixedCreateValidator:
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

		case blockSTMMixedSend:
			msg = banktypes.NewMsgSend(
				participantAddrs[op.account],
				participantAddrs[op.to],
				sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount)),
			)

		case blockSTMMixedDelegate:
			msg = stakingtypes.NewMsgDelegate(
				participantAddrs[op.account].String(),
				sdk.ValAddress(participantAddrs[op.validator]).String(),
				sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount),
			)

		case blockSTMMixedUndelegate:
			msg = stakingtypes.NewMsgUndelegate(
				participantAddrs[op.account].String(),
				sdk.ValAddress(participantAddrs[op.validator]).String(),
				sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount),
			)

		case blockSTMMixedRedelegate:
			msg = stakingtypes.NewMsgBeginRedelegate(
				participantAddrs[op.account].String(),
				sdk.ValAddress(participantAddrs[op.validator]).String(),
				sdk.ValAddress(participantAddrs[op.dstValidator]).String(),
				sdk.NewInt64Coin(sdk.DefaultBondDenom, op.amount),
			)

		case blockSTMMixedCancelUnbonding:
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
