package keeper_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// TestMultiSendNewAccountsGetUniqueIDs verifies that a single MsgMultiSend
// transaction that creates multiple new recipient accounts assigns each one
// a distinct account ID (account number). This exercises the hash-based
// account ID generation which derives uniqueness from the recipient address.
func TestMultiSendNewAccountsGetUniqueIDs(t *testing.T) {
	// --- Setup real keepers (no mocks) ---
	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, banktypes.StoreKey)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}).Codec
	logger := log.NewTestLogger(t)

	cms := integration.CreateMultiStore(keys, logger)
	ctx := sdk.NewContext(cms.RootCacheMultiStore(), cmtproto.Header{
		Height:  1,
		AppHash: []byte("test-app-hash"),
	}, false, logger).
		WithTxIndex(0).
		WithMsgIndex(0)

	authority := authtypes.NewModuleAddress("gov")

	maccPerms := map[string][]string{
		minttypes.ModuleName: {authtypes.Minter},
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
		map[string]bool{},
		authority.String(),
		log.NewNopLogger(),
	)

	// Initialize module genesis (creates module accounts like mint).
	authModule := auth.NewAppModule(cdc, accountKeeper, nil, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	authModule.InitGenesis(ctx, cdc, authModule.DefaultGenesis(cdc))
	bankModule.InitGenesis(ctx, cdc, bankModule.DefaultGenesis(cdc))

	// --- Create a funded sender account ---
	senderPriv := secp256k1.GenPrivKey()
	senderAddr := sdk.AccAddress(senderPriv.PubKey().Address())

	fundAmount := sdk.NewCoins(sdk.NewInt64Coin("stake", 10_000))
	require.NoError(t, banktestutil.FundAccount(ctx, bankKeeper, senderAddr, fundAmount))

	// Verify the sender was created and funded.
	senderAcc := accountKeeper.GetAccount(ctx, senderAddr)
	require.NotNil(t, senderAcc)

	// --- Generate 5 new recipient addresses (accounts don't exist yet) ---
	numRecipients := 5
	recipientAddrs := make([]sdk.AccAddress, numRecipients)
	for i := range numRecipients {
		recipientAddrs[i] = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		// Sanity: recipient should not exist before multisend.
		require.False(t, accountKeeper.HasAccount(ctx, recipientAddrs[i]),
			"recipient %d should not exist before multisend", i)
	}

	// --- Build and execute MsgMultiSend ---
	sendPerRecipient := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	outputs := make([]banktypes.Output, numRecipients)
	for i, addr := range recipientAddrs {
		outputs[i] = banktypes.NewOutput(addr, sendPerRecipient)
	}
	totalSend := sdk.NewCoins(sdk.NewInt64Coin("stake", int64(100*numRecipients)))
	input := banktypes.NewInput(senderAddr, totalSend)

	msgServer := bankkeeper.NewMsgServerImpl(bankKeeper)
	_, err := msgServer.MultiSend(ctx, banktypes.NewMsgMultiSend(input, outputs))
	require.NoError(t, err)

	// --- Verify all recipients were created with unique account IDs ---
	seenIDs := make(map[uint64]sdk.AccAddress)
	for i, addr := range recipientAddrs {
		acc := accountKeeper.GetAccount(ctx, addr)
		require.NotNil(t, acc, "recipient %d account should exist after multisend", i)

		accNum := acc.GetAccountNumber()

		// Account ID should have the top bit set (hash-based ID, not legacy sequential).
		require.True(t, accNum >= (1<<63),
			"recipient %d: account number %d should have top bit set (>= 2^63)", i, accNum)

		// Account ID must be unique across all recipients in this transaction.
		if prevAddr, exists := seenIDs[accNum]; exists {
			addrCodec := addresscodec.NewBech32Codec(sdk.Bech32MainPrefix)
			prevStr, _ := addrCodec.BytesToString(prevAddr)
			curStr, _ := addrCodec.BytesToString(addr)
			t.Fatalf("duplicate account number %d: recipient %s collides with %s",
				accNum, curStr, prevStr)
		}
		seenIDs[accNum] = addr
	}

	// Also verify sender's balance was debited.
	senderBalance := bankKeeper.GetBalance(ctx, senderAddr, "stake")
	expectedRemaining := fundAmount.AmountOf("stake").SubRaw(int64(100 * numRecipients))
	require.True(t, senderBalance.Amount.Equal(expectedRemaining),
		"sender balance: got %s, want %s", senderBalance.Amount, expectedRemaining)
}
