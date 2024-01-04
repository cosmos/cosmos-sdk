//go:build app_v1

package accounts

import (
	"context"
	"testing"

	"cosmossdk.io/simapp"
	"cosmossdk.io/x/accounts"
	rotationv1 "cosmossdk.io/x/accounts/testing/rotation/v1"
	accountsv1 "cosmossdk.io/x/accounts/v1"
	"cosmossdk.io/x/bank/testutil"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/nft"
	stakingtypes "cosmossdk.io/x/staking/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

var (
	privKey     = secp256k1.GenPrivKey()
	accCreator  = []byte("creator")
	bundlerAddr = secp256k1.GenPrivKey().PubKey().Address()
	aliceAddr   = secp256k1.GenPrivKey().PubKey().Address()
)

func TestAccountAbstraction(t *testing.T) {
	app := setupApp(t)
	ak := app.AccountsKeeper
	ctx := sdk.NewContext(app.CommitMultiStore(), false, app.Logger())

	_, aaAddr, err := ak.Init(ctx, "aa_minimal", accCreator, &rotationv1.MsgInit{
		PubKeyBytes: privKey.PubKey().Bytes(),
	})
	require.NoError(t, err)

	_, aaFullAddr, err := ak.Init(ctx, "aa_full", accCreator, &rotationv1.MsgInit{
		PubKeyBytes: privKey.PubKey().Bytes(),
	})
	require.NoError(t, err)

	aaAddrStr, err := app.AuthKeeper.AddressCodec().BytesToString(aaAddr)
	require.NoError(t, err)

	aaFullAddrStr, err := app.AuthKeeper.AddressCodec().BytesToString(aaFullAddr)
	require.NoError(t, err)

	// let's give aa some coins.
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, aaAddr, sdk.NewCoins(sdk.NewInt64Coin("stake", 100000000000))))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, aaFullAddr, sdk.NewCoins(sdk.NewInt64Coin("stake", 100000000000))))

	bundlerAddrStr, err := app.AuthKeeper.AddressCodec().BytesToString(bundlerAddr)
	require.NoError(t, err)

	aliceAddrStr, err := app.AuthKeeper.AddressCodec().BytesToString(aliceAddr)
	require.NoError(t, err)

	t.Run("ok - pay bundler and exec not implemented", func(t *testing.T) {
		// we simulate executing an user operation in an abstracted account
		// which only implements the authentication.
		resp := ak.ExecuteUserOperation(ctx, bundlerAddrStr, &accountsv1.UserOperation{
			Sender:                 aaAddrStr,
			AuthenticationMethod:   "secp256k1",
			AuthenticationData:     []byte("signature"),
			AuthenticationGasLimit: 10000,
			BundlerPaymentMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: aaAddrStr,
				ToAddress:   bundlerAddrStr,
				Amount:      coins(t, "1stake"), // the sender is the AA, so it has the coins and wants to pay the bundler for the gas
			}),
			BundlerPaymentGasLimit: 50000,
			ExecutionMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: aaAddrStr,
				ToAddress:   aliceAddrStr,
				Amount:      coins(t, "2000stake"), // as the real action the sender wants to send coins to alice
			}),
			ExecutionGasLimit: 36000,
		})
		require.Empty(t, resp.Error) // no error
		require.Len(t, resp.BundlerPaymentResponses, 1)
		require.Len(t, resp.ExecutionResponses, 1)
		require.NotZero(t, resp.ExecutionGasUsed)
		require.NotZero(t, resp.BundlerPaymentGasUsed)
		require.NotZero(t, resp.AuthenticationGasUsed)
		// assert there were state changes
		balanceIs(t, ctx, app, bundlerAddr.Bytes(), "1stake")  // pay bundler state change
		balanceIs(t, ctx, app, aliceAddr.Bytes(), "2000stake") // execute messages state change.
	})
	t.Run("pay bundle impersonation", func(t *testing.T) {
		// we simulate the execution of an abstracted account
		// which only implements authentication and tries in the pay
		// bundler messages the account tries to impersonate another one.
		resp := ak.ExecuteUserOperation(ctx, bundlerAddrStr, &accountsv1.UserOperation{
			Sender:                 aaAddrStr,
			AuthenticationMethod:   "secp256k1",
			AuthenticationData:     []byte("signature"),
			AuthenticationGasLimit: 10000,
			BundlerPaymentMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: bundlerAddrStr, // abstracted account tries to send money from bundler to itself.
				ToAddress:   aaAddrStr,
				Amount:      coins(t, "1stake"),
			}),
			BundlerPaymentGasLimit: 50000,
			ExecutionMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: aaAddrStr,
				ToAddress:   aliceAddrStr,
				Amount:      coins(t, "2000stake"), // as the real action the sender wants to send coins to alice
			}),
			ExecutionGasLimit: 36000,
		})
		require.Contains(t, resp.Error, accounts.ErrUnauthorized.Error()) // error is unauthorized
		require.Empty(t, resp.BundlerPaymentResponses)                    // no bundler payment responses, since the atomic exec failed
		require.Empty(t, resp.ExecutionResponses)                         // no execution responses, since the atomic exec failed
		require.Zero(t, resp.ExecutionGasUsed)                            // no execution gas used, since the atomic exec failed
		require.NotZero(t, resp.BundlerPaymentGasUsed)                    // bundler payment gas used, even if the atomic exec failed
	})
	t.Run("exec message impersonation", func(t *testing.T) {
		// we simulate a case in which the abstracted account tries to impersonate
		// someone else in the execution of messages.
		resp := ak.ExecuteUserOperation(ctx, bundlerAddrStr, &accountsv1.UserOperation{
			Sender:                 aaAddrStr,
			AuthenticationMethod:   "secp256k1",
			AuthenticationData:     []byte("signature"),
			AuthenticationGasLimit: 10000,
			BundlerPaymentMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: aaAddrStr,
				ToAddress:   bundlerAddrStr,
				Amount:      coins(t, "1stake"),
			}),
			BundlerPaymentGasLimit: 50000,
			ExecutionMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: aliceAddrStr, // abstracted account attempts to send money from alice to itself
				ToAddress:   aaAddrStr,
				Amount:      coins(t, "2000stake"),
			}),
			ExecutionGasLimit: 36000,
		})
		require.Contains(t, resp.Error, accounts.ErrUnauthorized.Error()) // error is unauthorized
		require.NotEmpty(t, resp.BundlerPaymentResponses)                 // bundler payment responses, since the bundler payment succeeded
		require.Empty(t, resp.ExecutionResponses)                         // no execution responses, since the atomic exec failed
		require.NotZero(t, resp.ExecutionGasUsed)                         // execution gas used, even if the atomic exec failed
		require.NotZero(t, resp.BundlerPaymentGasUsed)                    // bundler payment gas used, even if the atomic exec failed
	})
	t.Run("auth failure", func(t *testing.T) {
		// if auth fails nothing more should be attempted, the authentication
		// should have spent gas and the error should be returned.
		// we simulate a case in which the abstracted account tries to impersonate
		// someone else in the execution of messages.
		resp := ak.ExecuteUserOperation(ctx, bundlerAddrStr, &accountsv1.UserOperation{
			Sender:                 aaAddrStr,
			AuthenticationMethod:   "invalid",
			AuthenticationData:     []byte("signature"),
			AuthenticationGasLimit: 10000,
			BundlerPaymentMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: aaAddrStr,
				ToAddress:   bundlerAddrStr,
				Amount:      coins(t, "1stake"),
			}),
			BundlerPaymentGasLimit: 50000,
			ExecutionMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: aliceAddrStr, // abstracted account attempts to send money from alice to itself
				ToAddress:   aaAddrStr,
				Amount:      coins(t, "2000stake"),
			}),
			ExecutionGasLimit: 36000,
		})
		require.Contains(t, resp.Error, accounts.ErrAuthentication.Error()) // error is authentication
		require.Empty(t, resp.BundlerPaymentResponses)                      // no bundler payment responses, since the atomic exec failed
		require.Empty(t, resp.ExecutionResponses)                           // no execution responses, since the atomic exec failed
		require.Zero(t, resp.ExecutionGasUsed)                              // no execution gas used, since the atomic exec failed
		require.Zero(t, resp.BundlerPaymentGasUsed)                         // no bundler payment gas used, since the atomic exec failed
		require.NotZero(t, resp.AuthenticationGasUsed)                      // authentication gas used, even if the atomic exec failed
	})
	t.Run("pay bundle failure", func(t *testing.T) {
		// pay bundler fails, nothing more should be attempted, the authentication
		// succeeded. We expect gas used in auth and pay bundler step.
		// we simulate a case in which the abstracted account tries to impersonate
		// someone else in the execution of messages.
		resp := ak.ExecuteUserOperation(ctx, bundlerAddrStr, &accountsv1.UserOperation{
			Sender:                 aaAddrStr,
			AuthenticationMethod:   "secp256k1",
			AuthenticationData:     []byte("signature"),
			AuthenticationGasLimit: 10000,
			BundlerPaymentMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: aaAddrStr,
				ToAddress:   bundlerAddrStr,
				Amount:      coins(t, "1atom"), // abstracted account does not have enough money to pay the bundler, since it does not hold atom
			}),
			BundlerPaymentGasLimit: 50000,
			ExecutionMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: aliceAddrStr, // abstracted account attempts to send money from alice to itself
				ToAddress:   aaAddrStr,
				Amount:      coins(t, "2000stake"),
			}),
			ExecutionGasLimit: 36000,
		})
		require.Contains(t, resp.Error, accounts.ErrBundlerPayment.Error()) // error is bundler payment
		require.Empty(t, resp.BundlerPaymentResponses)                      // no bundler payment responses, since the atomic exec failed
		require.Empty(t, resp.ExecutionResponses)                           // no execution responses, since the atomic exec failed
		require.Zero(t, resp.ExecutionGasUsed)                              // no execution gas used, since the atomic exec failed
		require.NotZero(t, resp.BundlerPaymentGasUsed)                      // bundler payment gas used, even if the atomic exec failed
		require.NotZero(t, resp.AuthenticationGasUsed)                      // authentication gas used, even if the atomic exec failed
	})
	t.Run("exec message failure", func(t *testing.T) {
		// execution message fails, nothing more should be attempted, the authentication
		// and pay bundler succeeded. We expect gas used in auth, pay bundler and
		// execution step.
		resp := ak.ExecuteUserOperation(ctx, bundlerAddrStr, &accountsv1.UserOperation{
			Sender:                 aaAddrStr,
			AuthenticationMethod:   "secp256k1",
			AuthenticationData:     []byte("signature"),
			AuthenticationGasLimit: 10000,
			BundlerPaymentMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: aaAddrStr,
				ToAddress:   bundlerAddrStr,
				Amount:      coins(t, "1stake"),
			}),
			BundlerPaymentGasLimit: 50000,
			ExecutionMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: aaAddrStr,
				ToAddress:   aliceAddrStr,
				Amount:      coins(t, "2000atom"), // abstracted account does not have enough money to pay alice, since it does not hold atom
			}),
			ExecutionGasLimit: 36000,
		})
		require.Contains(t, resp.Error, accounts.ErrExecution.Error()) // error is execution
		require.Len(t, resp.BundlerPaymentResponses, 1)                // bundler payment response, since the pay bundler succeeded
		require.Empty(t, resp.ExecutionResponses)                      // no execution responses, since the atomic exec failed
		require.NotZero(t, resp.ExecutionGasUsed)                      // execution gas used, even if the atomic exec failed
		require.NotZero(t, resp.BundlerPaymentGasUsed)                 // bundler payment gas used, even if the atomic exec failed
		require.NotZero(t, resp.AuthenticationGasUsed)                 // authentication gas used, even if the atomic exec failed
	})

	t.Run("implements bundler payment - fail ", func(t *testing.T) {
		// we assert that if an aa implements the bundler payment interface, then
		// that is called.
		resp := ak.ExecuteUserOperation(ctx, bundlerAddrStr, &accountsv1.UserOperation{
			Sender:                 aaFullAddrStr,
			AuthenticationMethod:   "secp256k1",
			AuthenticationData:     []byte("signature"),
			AuthenticationGasLimit: 10000,
			BundlerPaymentMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: aaFullAddrStr,
				ToAddress:   bundlerAddrStr,
				Amount:      coins(t, "1stake"), // we expect this to fail since the account is implement in such a way not to allow bank sends.
			}),
			BundlerPaymentGasLimit: 50000,
			ExecutionMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: aaFullAddrStr,
				ToAddress:   aliceAddrStr,
				Amount:      coins(t, "2000stake"),
			}),
			ExecutionGasLimit: 36000,
		})
		// in order to assert the call we expect an error to be returned.
		require.Contains(t, resp.Error, accounts.ErrBundlerPayment.Error())               // error is bundler payment
		require.Contains(t, resp.Error, "this account does not allow bank send messages") // error is bundler payment
	})

	t.Run("implements execution - fail", func(t *testing.T) {
		resp := ak.ExecuteUserOperation(ctx, bundlerAddrStr, &accountsv1.UserOperation{
			Sender:                 aaFullAddrStr,
			AuthenticationMethod:   "secp256k1",
			AuthenticationData:     []byte("signature"),
			AuthenticationGasLimit: 10000,
			BundlerPaymentMessages: nil,
			BundlerPaymentGasLimit: 50000,
			ExecutionMessages: intoAny(t, &stakingtypes.MsgDelegate{
				DelegatorAddress: aaFullAddrStr,
				ValidatorAddress: "some-validator",
				Amount:           coins(t, "2000stake")[0],
			}),
			ExecutionGasLimit: 36000,
		})
		// in order to assert the call we expect an error to be returned.
		require.Contains(t, resp.Error, accounts.ErrExecution.Error()) // error is in execution
		require.Contains(t, resp.Error, "this account does not allow delegation messages")
	})

	t.Run("implements bundler payment and execution - success", func(t *testing.T) {
		// we simulate the abstracted account pays the bundler using an NFT.
		require.NoError(t, app.NFTKeeper.SaveClass(ctx, nft.Class{
			Id: "omega-rare",
		}))
		require.NoError(t, app.NFTKeeper.Mint(ctx, nft.NFT{
			ClassId: "omega-rare",
			Id:      "the-most-rare",
		}, aaFullAddr))

		resp := ak.ExecuteUserOperation(ctx, bundlerAddrStr, &accountsv1.UserOperation{
			Sender:                 aaFullAddrStr,
			AuthenticationMethod:   "secp256k1",
			AuthenticationData:     []byte("signature"),
			AuthenticationGasLimit: 10000,
			BundlerPaymentMessages: intoAny(t, &nft.MsgSend{
				ClassId:  "omega-rare",
				Id:       "the-most-rare",
				Sender:   aaFullAddrStr,
				Receiver: bundlerAddrStr,
			}),
			BundlerPaymentGasLimit: 50000,
			ExecutionMessages: intoAny(t, &banktypes.MsgSend{
				FromAddress: aaFullAddrStr,
				ToAddress:   aliceAddrStr,
				Amount:      coins(t, "2000stake"),
			}),
			ExecutionGasLimit: 36000,
		})
		require.Empty(t, resp.Error) // no error
	})
}

func intoAny(t *testing.T, msgs ...gogoproto.Message) (anys []*codectypes.Any) {
	t.Helper()
	for _, msg := range msgs {
		any, err := codectypes.NewAnyWithValue(msg)
		require.NoError(t, err)
		anys = append(anys, any)
	}
	return
}

func coins(t *testing.T, s string) sdk.Coins {
	t.Helper()
	coins, err := sdk.ParseCoinsNormalized(s)
	require.NoError(t, err)
	return coins
}

func balanceIs(t *testing.T, ctx context.Context, app *simapp.SimApp, addr sdk.AccAddress, s string) {
	t.Helper()
	balance := app.BankKeeper.GetAllBalances(ctx, addr)
	require.Equal(t, s, balance.String())
}
