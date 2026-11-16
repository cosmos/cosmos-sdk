package simapp

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
)

type appGasSimulationConn struct {
	app   *SimApp
	calls []sdk.GasInfo
}

func (c *appGasSimulationConn) Invoke(_ context.Context, _ string, req, reply any, _ ...grpc.CallOption) error {
	info, result, err := c.app.Simulate(req.(*txtypes.SimulateRequest).TxBytes)
	if err != nil {
		return err
	}
	c.calls = append(c.calls, info)
	*reply.(*txtypes.SimulateResponse) = txtypes.SimulateResponse{GasInfo: &info, Result: result}
	return nil
}

func (*appGasSimulationConn) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("unused")
}

func TestAutomaticGasIncludesFeePayment(t *testing.T) {
	for _, collectorFunded := range []bool{false, true} {
		for _, useFeegrant := range []bool{false, true} {
			for _, legacyEstimate := range []bool{true, false} {
				t.Run(fmt.Sprintf("collector-funded=%t/feegrant=%t/legacy-estimate=%t", collectorFunded, useFeegrant, legacyEstimate), func(t *testing.T) {
					const chainID = "gas-test"
					app := NewSimApp(log.NewNopLogger(), dbm.NewMemDB(), true,
						simtestutil.NewAppOptionsWithFlagHome(t.TempDir()), baseapp.SetChainID(chainID))
					config := app.TxConfig()
					kb := keyring.NewInMemory(app.AppCodec())
					record, _, err := kb.NewMnemonic("sender", keyring.English, hd.CreateHDPath(118, 0, 0).String(), "", hd.Secp256k1)
					require.NoError(t, err)
					pubkey, err := record.GetPubKey()
					require.NoError(t, err)
					sender := sdk.AccAddress(pubkey.Address())
					recipientKey := secp256k1.GenPrivKey().PubKey()
					recipient := sdk.AccAddress(recipientKey.Address())
					accounts := []authtypes.GenesisAccount{
						authtypes.NewBaseAccount(sender, pubkey, 0, 0),
						authtypes.NewBaseAccount(recipient, recipientKey, 1, 0),
					}
					validator := mock.NewPV()
					consensusKey, err := validator.GetPubKey()
					require.NoError(t, err)
					valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{cmttypes.NewValidator(consensusKey, 1)})
					genesis, err := simtestutil.GenesisStateWithValSet(app.AppCodec(), app.DefaultGenesis(), valSet, accounts,
						banktypes.Balance{Address: sender.String(), Coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 10000000))},
						banktypes.Balance{Address: recipient.String(), Coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 10000000))})
					require.NoError(t, err)
					state, err := json.Marshal(genesis)
					require.NoError(t, err)
					blockTime := time.Unix(100, 0).UTC()
					_, err = app.InitChain(&abci.RequestInitChain{
						ChainId: chainID, Time: blockTime, AppStateBytes: state, ConsensusParams: simtestutil.DefaultConsensusParams,
					})
					require.NoError(t, err)
					_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Time: blockTime, NextValidatorsHash: valSet.Hash()})
					require.NoError(t, err)
					ctx := app.NewContextLegacy(false, cmtproto.Header{Height: 1, Time: blockTime, ChainID: chainID})
					if collectorFunded {
						require.NoError(t, app.BankKeeper.SendCoinsFromAccountToModule(ctx, sender, authtypes.FeeCollectorName,
							sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))))
					}
					if useFeegrant {
						require.NoError(t, app.FeeGrantKeeper.GrantAllowance(ctx, recipient, sender,
							&feegrant.BasicAllowance{SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("stake", 1000000))}))
					}
					app.SimWriteState()
					_, err = app.Commit()
					require.NoError(t, err)
					ctx = app.NewContextLegacy(true, cmtproto.Header{Height: 1, Time: blockTime, ChainID: chainID})
					account := app.AccountKeeper.GetAccount(ctx, sender)
					balanceBefore := app.BankKeeper.GetBalance(ctx, sender, "stake")
					granterBalanceBefore := app.BankKeeper.GetBalance(ctx, recipient, "stake")
					collector := app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
					require.Equal(t, collectorFunded, app.BankKeeper.GetBalance(ctx, collector, "stake").IsPositive())
					f := tx.Factory{}.WithTxConfig(config).WithChainID(chainID).WithKeybase(kb).WithFromName("sender").
						WithAccountNumber(account.GetAccountNumber()).WithSequence(account.GetSequence()).
						WithSimulateAndExecute(true).WithGasPrices("0.025stake").WithGasAdjustment(1)
					if useFeegrant {
						f = f.WithFeeGranter(recipient)
					}
					msg := banktypes.NewMsgSend(sender, recipient, sdk.NewCoins(sdk.NewInt64Coin("stake", 1)))
					conn := &appGasSimulationConn{app: app}
					res, limit, err := tx.CalculateGas(conn, f, msg)
					require.NoError(t, err)
					require.Greater(t, conn.calls[1].GasUsed, conn.calls[0].GasUsed, "fee-bearing simulation includes bank payment")
					require.Equal(t, account.GetSequence(), app.AccountKeeper.GetAccount(ctx, sender).GetSequence())
					require.Equal(t, balanceBefore, app.BankKeeper.GetBalance(ctx, sender, "stake"))
					require.Equal(t, granterBalanceBefore, app.BankKeeper.GetBalance(ctx, recipient, "stake"))
					if useFeegrant {
						allowance, err := app.FeeGrantKeeper.GetAllowance(ctx, recipient, sender)
						require.NoError(t, err)
						require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("stake", 1000000)), allowance.(*feegrant.BasicAllowance).SpendLimit)
					}
					if legacyEstimate {
						// The old single-simulation estimator returned this zero-fee measurement.
						limit = conn.calls[0].GasUsed
					}
					built, err := f.WithGas(limit).BuildUnsignedTx(msg)
					require.NoError(t, err)
					require.NoError(t, tx.Sign(context.Background(), f.WithGas(limit), "sender", built, true))
					bz, err := config.TxEncoder()(built.GetTx())
					require.NoError(t, err)
					block, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
						Height: 2, Time: blockTime.Add(time.Second), NextValidatorsHash: valSet.Hash(), Txs: [][]byte{bz},
					})
					require.NoError(t, err)
					require.Len(t, block.TxResults, 1)
					result := block.TxResults[0]
					if legacyEstimate {
						require.EqualValues(t, 11, result.Code, "%+v", result)
					} else {
						require.Zero(t, result.Code, "%+v", result)
						require.LessOrEqual(t, uint64(result.GasUsed), limit)
					}
					t.Logf("bootstrap=%d measured=%d limit=%d execution=%d code=%d simulations=%d",
						conn.calls[0].GasUsed, res.GasInfo.GasUsed, limit, result.GasUsed, result.Code, len(conn.calls))
				})
			}
		}
	}
}
