package cmd

import (
	"math/rand"
	"os"
	"time"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/tools/speedtest"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func NewBankSpeedTest() *cobra.Command {
	cmd := &cobra.Command{
		Use: "speedtest",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.MkdirTemp("", "bankspeedtest-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(dir)
			db, err := dbm.NewDB("app", dbm.PebbleDBBackend, dir)
			if err != nil {
				return err
			}
			defer db.Close()
			chainID := "foo"
			app := simapp.NewSimApp(log.NewNopLogger(), db, nil, true, simtestutil.NewAppOptionsWithFlagHome(dir), baseapp.SetChainID(chainID))
			gen := generator{
				app:      app,
				accounts: make([]accountInfo, 0),
			}
			speedCmd := speedtest.NewCmd(gen.createAccount, gen.generateTx, app, app.AppCodec(), app.DefaultGenesis(), chainID)
			speedCmd.SetArgs(args)
			return speedCmd.Execute()
		},
	}
	cmd.DisableFlagParsing = true
	return cmd
}

type generator struct {
	app      *simapp.SimApp
	accounts []accountInfo
}

type accountInfo struct {
	privKey cryptotypes.PrivKey
	address sdk.AccAddress
	accNum  uint64
	seqNum  uint64
}

func (g *generator) createAccount() (*authtypes.BaseAccount, sdk.Coins) {
	privKey := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(privKey.PubKey().Address())
	accNum := len(g.accounts)
	baseAcc := authtypes.NewBaseAccount(addr, privKey.PubKey(), uint64(accNum), 0)

	g.accounts = append(g.accounts, accountInfo{
		privKey: privKey,
		address: addr,
		accNum:  uint64(accNum),
		seqNum:  0,
	})

	return baseAcc, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000))
}

func (g *generator) generateTx() []byte {
	senderIdx := r.Intn(len(g.accounts))
	recipientIdx := (senderIdx + 1 + r.Intn(len(g.accounts)-1)) % len(g.accounts)
	sender := g.accounts[senderIdx]
	recipient := g.accounts[recipientIdx]
	sendAmount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1))
	msg := banktypes.NewMsgSend(sender.address, recipient.address, sendAmount)
	txConfig := g.app.TxConfig()
	// Build and sign transaction
	tx, err := simtestutil.GenSignedMockTx(
		r,
		txConfig,
		[]sdk.Msg{msg},
		sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)),
		simtestutil.DefaultGenTxGas,
		g.app.ChainID(),
		[]uint64{sender.accNum},
		[]uint64{sender.seqNum},
		sender.privKey,
	)
	if err != nil {
		panic(err)
	}
	txBytes, err := txConfig.TxEncoder()(tx)
	if err != nil {
		panic(err)
	}
	g.accounts[senderIdx].seqNum++
	return txBytes
}
