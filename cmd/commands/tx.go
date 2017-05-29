package commands

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/types"

	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/tendermint/rpc/client"
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
		RunE:  sendTxCmd,
	}

	AppTxCmd = &cobra.Command{
		Use:   "app",
		Short: "An AppTx transaction, for sending raw data to plugins",
		RunE:  appTxCmd,
	}
)

var (
	//persistent flags
	txNodeFlag  string
	amountFlag  string
	fromFlag    string
	seqFlag     int
	gasFlag     int
	feeFlag     string
	chainIDFlag string

	//non-persistent flags
	toFlag   string
	dataFlag string
	nameFlag string
)

func init() {

	// register flags
	cmdTxFlags := []Flag2Register{
		{&txNodeFlag, "node", "tcp://localhost:46657", "Tendermint RPC address"},
		{&chainIDFlag, "chain_id", "test_chain_id", "ID of the chain for replay protection"},
		{&fromFlag, "from", "key.json", "Path to a private key to sign the transaction"},
		{&amountFlag, "amount", "", "Coins to send in transaction of the format <amt><coin>,<amt2><coin2>,... (eg: 1btc,2gold,5silver)"},
		{&gasFlag, "gas", 0, "The amount of gas for the transaction"},
		{&feeFlag, "fee", "0coin", "Coins for the transaction fee of the format <amt><coin>"},
		{&seqFlag, "sequence", -1, "Sequence number for the account (-1 to autocalculate)"},
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

func sendTxCmd(cmd *cobra.Command, args []string) error {

	var toHex string
	var chainPrefix string
	spl := strings.Split(toFlag, "/")
	switch len(spl) {
	case 1:
		toHex = spl[0]
	case 2:
		chainPrefix = spl[0]
		toHex = spl[1]
	default:
		return errors.Errorf("To address has too many slashes")
	}

	// convert destination address to bytes
	to, err := hex.DecodeString(StripHex(toHex))
	if err != nil {
		return errors.Errorf("To address is invalid hex: %v\n", err)
	}

	if chainPrefix != "" {
		to = []byte(chainPrefix + "/" + string(to))
	}

	// load the priv key
	privKey, err := LoadKey(fromFlag)
	if err != nil {
		return err
	}

	// get the sequence number for the tx
	sequence, err := getSeq(privKey.Address[:])
	if err != nil {
		return err
	}

	//parse the fee and amounts into coin types
	feeCoin, err := types.ParseCoin(feeFlag)
	if err != nil {
		return err
	}
	amountCoins, err := types.ParseCoins(amountFlag)
	if err != nil {
		return err
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
	tx.Inputs[0].Signature = privKey.Sign(signBytes)

	fmt.Println("Signed SendTx:")
	fmt.Println(string(wire.JSONBytes(tx)))

	// broadcast the transaction to tendermint
	data, log, err := broadcastTx(tx)
	if err != nil {
		return err
	}
	fmt.Printf("Response: %X ; %s\n", data, log)
	return nil
}

func appTxCmd(cmd *cobra.Command, args []string) error {
	// convert data to bytes
	data := []byte(dataFlag)
	if isHex(dataFlag) {
		data, _ = hex.DecodeString(dataFlag)
	}
	name := nameFlag
	return AppTx(name, data)
}

func AppTx(name string, data []byte) error {

	privKey, err := LoadKey(fromFlag)
	if err != nil {
		return err
	}

	sequence, err := getSeq(privKey.Address[:])
	if err != nil {
		return err
	}

	//parse the fee and amounts into coin types
	feeCoin, err := types.ParseCoin(feeFlag)
	if err != nil {
		return err
	}

	amountCoins, err := types.ParseCoins(amountFlag)
	if err != nil {
		return err
	}

	input := types.NewTxInput(privKey.PubKey, amountCoins, sequence)
	tx := &types.AppTx{
		Gas:   int64(gasFlag),
		Fee:   feeCoin,
		Name:  name,
		Input: input,
		Data:  data,
	}

	tx.Input.Signature = privKey.Sign(tx.SignBytes(chainIDFlag))

	fmt.Println("Signed AppTx:")
	fmt.Println(string(wire.JSONBytes(tx)))

	data, log, err := broadcastTx(tx)
	if err != nil {
		return err
	}
	fmt.Printf("Response: %X ; %s\n", data, log)
	return nil
}

// broadcast the transaction to tendermint
func broadcastTx(tx types.Tx) ([]byte, string, error) {
	httpClient := client.NewHTTP(txNodeFlag, "/websocket")
	// Don't you hate having to do this?
	// How many times have I lost an hour over this trick?!
	txBytes := []byte(wire.BinaryBytes(struct {
		types.Tx `json:"unwrap"`
	}{tx}))
	res, err := httpClient.BroadcastTxCommit(txBytes)
	if err != nil {
		return nil, "", errors.Errorf("Error on broadcast tx: %v", err)
	}

	// if it fails check, we don't even get a delivertx back!
	if !res.CheckTx.Code.IsOK() {
		r := res.CheckTx
		return nil, "", errors.Errorf("BroadcastTxCommit got non-zero exit code: %v. %X; %s", r.Code, r.Data, r.Log)
	}

	if !res.DeliverTx.Code.IsOK() {
		r := res.DeliverTx
		return nil, "", errors.Errorf("BroadcastTxCommit got non-zero exit code: %v. %X; %s", r.Code, r.Data, r.Log)
	}

	return res.DeliverTx.Data, res.DeliverTx.Log, nil
}

// if the sequence flag is set, return it;
// else, fetch the account by querying the app and return the sequence number
func getSeq(address []byte) (int, error) {
	if seqFlag >= 0 {
		return seqFlag, nil
	}

	httpClient := client.NewHTTP(txNodeFlag, "/websocket")
	acc, err := getAccWithClient(httpClient, address)
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
