package commands

import (
	"fmt"

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
	RegisterPersistentFlags(TxCmd, cmdTxFlags)
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

	out := wire.BinaryBytes(tx)
	fmt.Println("Signed AppTx:")
	fmt.Printf("%X\n", out)

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
