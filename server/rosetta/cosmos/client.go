package cosmos

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/go-amino"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

const (
	OptionAddress = "address"
	OptionGas     = "gas"
	OperationFee  = "fee"
	OptionMemo    = "memo"
)
const (
	// Metadata Keys
	ChainIDKey       = "chain_id"
	SequenceKey      = "sequence"
	AccountNumberKey = "account_number"
	GasKey           = "gas"
)

// sdk operation identifiers
var (
	opDelegate          = "delegate"
	opFee               = "fee"
	opBankSend          = "send"
	supportedOperations = []string{opDelegate, opFee, opBankSend}
)

type Client struct {
	tm        rpcclient.Client
	lcd       string
	cdc       *amino.Codec
	txDecoder sdk.TxDecoder
	txEncoder sdk.TxEncoder
}

func (d Client) SupportedOperations() []string {
	return supportedOperations
}

func (d Client) NodeVersion() string {
	return "0.37.12"
}

func NewDataClient(tmEndpoint string, lcdEndpoint string, cdc *amino.Codec) (Client, error) {
	tmClient := rpcclient.NewHTTP(tmEndpoint, "/websocket")
	// test it works
	_, err := tmClient.Health()
	if err != nil {
		return Client{}, err
	}
	dc := Client{
		tm:        tmClient,
		lcd:       lcdEndpoint,
		cdc:       cdc,
		txDecoder: auth.DefaultTxDecoder(cdc),
		txEncoder: auth.DefaultTxEncoder(cdc),
	}
	return dc, nil
}
