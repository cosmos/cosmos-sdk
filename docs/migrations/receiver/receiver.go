package main

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	kr "github.com/cosmos/cosmos-sdk/crypto/keyring"
	types2 "github.com/cosmos/cosmos-sdk/docs/migrations/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"log"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	//kr "github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

func main() {
	log.SetOutput(os.Stdout)
	http.HandleFunc("/", Send)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Send(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s - %s\n", r.Method, r.URL.Path)

	encodingConfig := simapp.MakeTestEncodingConfig()

	keyring := kr.NewInMemory()

	cfg := client.Context{}.
		WithJSONMarshaler(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(".example").
		WithKeyring(keyring).
		WithSkipConfirmation(true)

	var bsb types2.BankSendBody

	infoaccount, err := cfg.Keyring.NewAccount("s", types2.Mnemonic, "", "", hd.Secp256k1)
	if err != nil {
		http.Error(w, errors.Wrap(err, "keyring.NewAccount").Error(), 500)
		return
	}
	//senderAddr := info.GetAddress()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, errors.Wrap(err, "send:ioutil.ReadAll").Error(), 500)
		return
	}

	if err := json.Unmarshal(body, &bsb); err != nil {
		http.Error(w, errors.Wrap(err, "send:json.Unmarshal").Error(), 500)
		return
	}

	info, err := cfg.Keyring.KeyByAddress(infoaccount.GetAddress())
	if err != nil {
		http.Error(w, errors.Wrap(err, "sdk.keyring.KeyByAddress").Error(), 500)
		return
	}

	bsb.Denom = "atom"
	coins := types.NewCoins(types.NewInt64Coin("atom", bsb.Amount))

	txfNoKeybase := tx.Factory{}.
		WithTxConfig(encodingConfig.TxConfig).
		//WithAccountNumber(bsb.AccountNumber).
		//WithSequence(bsb.Sequence).
		WithFees(fmt.Sprintf("%d%s", bsb.Fee, coins.GetDenomByIndex(0))).
		WithMemo(bsb.Memo).
		WithGas(bsb.Gas).
		WithGasAdjustment(bsb.GasAdjustment).
		WithChainID(bsb.ChainID)

	txfDirect := txfNoKeybase.
		WithKeybase(cfg.Keyring).
		WithSignMode(signingtypes.SignMode_SIGN_MODE_DIRECT)

	//txfAmino := txfDirect.
	//	WithSignMode(signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)

	msg := banktypes.NewMsgSend(info.GetAddress(), bsb.Receiver, coins)
	//pubKey := info.GetPubKey()

	//txBuilder := cfg.TxConfig.NewTxBuilder()

	txb, err := tx.BuildUnsignedTx(txfDirect, msg)
	if err != nil {
		http.Error(w, errors.Wrap(err, "sdk.tx.BuildUnsignedTx").Error(), 500)
		return
	}

	txb.SetMemo(bsb.Memo)

	fmt.Println("info:>", info)
	fmt.Println("info:>", infoaccount)

	fmt.Println("sender:>", info.GetAddress())
	fmt.Println("sender:>", infoaccount.GetAddress())

	if err = tx.Sign(TxFactory(cfg), "s", txb, true); err != nil {
		http.Error(w, errors.Wrap(err, "sdk.tx.Sign").Error(), 500)
		return
	}

	stx := txb.GetTx()
	log.Printf("stx: %+v\n", stx)
	fmt.Printf("stx: %+v\n", stx)
	sigs, err := stx.GetSignaturesV2()
	if err != nil {
		http.Error(w, errors.Wrap(err, "sdk.stx.GetSignaturesV2").Error(), 500)
		return
	}

	for _, sig := range sigs {
		fmt.Printf("sig: %+v\n", sig)
	}
}

// TxFactory returns a factory for sending transactions
func TxFactory(ctx client.Context) tx.Factory {
	return tx.Factory{}.
		WithTxConfig(ctx.TxConfig).
		WithAccountRetriever(ctx.AccountRetriever).
		WithKeybase(ctx.Keyring).
		WithChainID("chain1").
		WithSimulateAndExecute(true).
		WithGasAdjustment(1.5).
		WithGasPrices("5000uatom").
		WithSignMode(signingtypes.SignMode_SIGN_MODE_DIRECT)
}
