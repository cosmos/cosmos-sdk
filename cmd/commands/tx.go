package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/types"
	crypto "github.com/tendermint/go-crypto"

	cmn "github.com/tendermint/go-common"
	client "github.com/tendermint/go-rpc/client"
	"github.com/tendermint/go-wire"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

//commands
var (
	TxCmd = &cobra.Command{
		Use:   "tx",
		Short: "Create, sign, and broadcast a transaction",
	}

	SendTxCmd = &cobra.Command{
		Use:   "send",
		Short: "A SendTx transaction, for sending tokens around",
		Run:   sendTxCmd,
	}

	AppTxCmd = &cobra.Command{
		Use:   "app",
		Short: "An AppTx transaction, for sending raw data to plugins",
		Run:   appTxCmd,
	}
)

//flags
var (
	txNodeFlag  string
	toFlag      string
	amountFlag  string
	fromFlag    string
	seqFlag     int
	gasFlag     int
	feeFlag     string
	dataFlag    string
	nameFlag    string
	chainIDFlag string
)

func init() {

	// register flags
	cmdTxFlags := []Flag2Register{
		{&txNodeFlag, "node", "tcp://localhost:46657", "Tendermint RPC address"},
		{&chainIDFlag, "chain_id", "test_chain_id", "ID of the chain for replay protection"},
		{&fromFlag, "from", "key.json", "Path to a private key to sign the transaction"},
		{&amountFlag, "amount", "", "Coins to send in transaction of the format <amt><coin>,<amt2><coin2>,... (eg: 1btc,2gold,5silver},"},
		{&gasFlag, "gas", 0, "The amount of gas for the transaction"},
		{&feeFlag, "fee", "", "Coins for the transaction fee of the format <amt><coin>"},
		{&seqFlag, "sequence", -1, "Sequence number for the account (-1 to autocalculate},"},
	}

	sendTxFlags := []Flag2Register{
		{&toFlag, "to", "", "Destination address for the transaction"},
	}

	appTxFlags := []Flag2Register{
		{&nameFlag, "name", "", "Plugin to send the transaction to"},
		{&dataFlag, "data", "", "Data to send with the transaction"},
	}

	RegisterPersistentFlags(TxCmd, cmdTxFlags)
	RegisterFlags(SendTxCmd, sendTxFlags)
	RegisterFlags(AppTxCmd, appTxFlags)

	//register commands
	TxCmd.AddCommand(SendTxCmd, AppTxCmd)
}

func sendTxCmd(cmd *cobra.Command, args []string) {

	// convert destination address to bytes
	to, err := hex.DecodeString(StripHex(toFlag))
	if err != nil {
		cmn.Exit(fmt.Sprintf("To address is invalid hex: %+v\n", err))
	}

	// load the priv key
	privKey := LoadKey(fromFlag)

	// get the sequence number for the tx
	sequence, err := getSeq(privKey.Address[:])
	if err != nil {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}

	//parse the fee and amounts into coin types
	feeCoin, err := ParseCoin(feeFlag)
	if err != nil {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}

	amountCoins, err := ParseCoins(amountFlag)
	if err != nil {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}

	// craft the tx
	input := types.NewTxInput(privKey.PubKey, amountCoins, sequence)
	output := newOutput(to, amountCoins)
	tx := &types.SendTx{
		Gas:     int64(gasFlag),
		Fee:     feeCoin,
		Inputs:  []types.TxInput{input},
		Outputs: []types.TxOutput{output},
	}

	// sign that puppy
	signBytes := tx.SignBytes(chainIDFlag)
	tx.Inputs[0].Signature = crypto.SignatureS{privKey.Sign(signBytes)}

	fmt.Println("Signed SendTx:")
	fmt.Println(string(wire.JSONBytes(tx)))

	// broadcast the transaction to tendermint
	data, log, err := broadcastTx(tx)
	if err != nil {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}
	fmt.Printf("Response: %X ; %s\n", data, log)
}

func appTxCmd(cmd *cobra.Command, args []string) {
	// convert data to bytes
	data := []byte(dataFlag)
	if isHex(dataFlag) {
		data, _ = hex.DecodeString(dataFlag)
	}
	name := nameFlag
	AppTx(name, data)
}

func AppTx(name string, data []byte) {

	privKey := LoadKey(fromFlag)

	sequence, err := getSeq(privKey.Address[:])
	if err != nil {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}

	//parse the fee and amounts into coin types
	feeCoin, err := ParseCoin(feeFlag)
	if err != nil {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}

	amountCoins, err := ParseCoins(amountFlag)
	if err != nil {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}

	input := types.NewTxInput(privKey.PubKey, amountCoins, sequence)
	tx := &types.AppTx{
		Gas:   int64(gasFlag),
		Fee:   feeCoin,
		Name:  name,
		Input: input,
		Data:  data,
	}

	tx.Input.Signature = crypto.SignatureS{privKey.Sign(tx.SignBytes(chainIDFlag))}

	fmt.Println("Signed AppTx:")
	fmt.Println(string(wire.JSONBytes(tx)))

	data, log, err := broadcastTx(tx)
	if err != nil {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}
	fmt.Printf("Response: %X ; %s\n", data, log)
}

// broadcast the transaction to tendermint
func broadcastTx(tx types.Tx) ([]byte, string, error) {

	tmResult := new(ctypes.TMResult)
	clientURI := client.NewClientURI(txNodeFlag)

	// Don't you hate having to do this?
	// How many times have I lost an hour over this trick?!
	txBytes := []byte(wire.BinaryBytes(struct {
		types.Tx `json:"unwrap"`
	}{tx}))

	_, err := clientURI.Call("broadcast_tx_commit", map[string]interface{}{"tx": txBytes}, tmResult)
	if err != nil {
		return nil, "", errors.New(cmn.Fmt("Error on broadcast tx: %v", err))
	}

	res := (*tmResult).(*ctypes.ResultBroadcastTxCommit)

	// if it fails check, we don't even get a delivertx back!
	if !res.CheckTx.Code.IsOK() {
		r := res.CheckTx
		return nil, "", errors.New(cmn.Fmt("BroadcastTxCommit got non-zero exit code: %v. %X; %s", r.Code, r.Data, r.Log))
	}

	if !res.DeliverTx.Code.IsOK() {
		r := res.DeliverTx
		return nil, "", errors.New(cmn.Fmt("BroadcastTxCommit got non-zero exit code: %v. %X; %s", r.Code, r.Data, r.Log))
	}

	return res.DeliverTx.Data, res.DeliverTx.Log, nil
}

// if the sequence flag is set, return it;
// else, fetch the account by querying the app and return the sequence number
func getSeq(address []byte) (int, error) {

	if seqFlag >= 0 {
		return seqFlag, nil
	}

	acc, err := getAcc(txNodeFlag, address)
	if err != nil {
		return 0, err
	}
	return acc.Sequence + 1, nil
}

func newOutput(to []byte, amount types.Coins) types.TxOutput {
	return types.TxOutput{
		Address: to,
		Coins:   amount,
	}
}
