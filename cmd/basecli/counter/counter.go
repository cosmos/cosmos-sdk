package counter

import (
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
	lightclient "github.com/tendermint/light-client"
	"github.com/tendermint/light-client/commands"
	"github.com/tendermint/light-client/commands/txs"

	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"
	"github.com/tendermint/basecoin/plugins/counter"
	btypes "github.com/tendermint/basecoin/types"
)

type CounterPresenter struct{}

func (_ CounterPresenter) MakeKey(str string) ([]byte, error) {
	key := counter.New().StateKey()
	return key, nil
}

func (_ CounterPresenter) ParseData(raw []byte) (interface{}, error) {
	var cp counter.CounterPluginState
	err := wire.ReadBinaryBytes(raw, &cp)
	return cp, err
}

/**** build out the tx ****/

var (
	_ txs.ReaderMaker      = CounterTxMaker{}
	_ lightclient.TxReader = CounterTxReader{}
)

type CounterTxMaker struct{}

func (m CounterTxMaker) MakeReader() (lightclient.TxReader, error) {
	chainID := viper.GetString(commands.ChainFlag)
	return CounterTxReader{bcmd.AppTxReader{ChainID: chainID}}, nil
}

// define flags

type CounterFlags struct {
	bcmd.AppFlags `mapstructure:",squash"`
	Valid         bool
	CountFee      string
}

func (m CounterTxMaker) Flags() (*flag.FlagSet, interface{}) {
	fs, app := bcmd.AppFlagSet()
	fs.String("countfee", "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	fs.Bool("valid", false, "Is count valid?")
	return fs, &CounterFlags{AppFlags: app}
}

// parse flags

type CounterTxReader struct {
	App bcmd.AppTxReader
}

func (t CounterTxReader) ReadTxJSON(data []byte, pk crypto.PubKey) (interface{}, error) {
	// TODO: something.  maybe?
	return t.App.ReadTxJSON(data, pk)
}

func (t CounterTxReader) ReadTxFlags(flags interface{}, pk crypto.PubKey) (interface{}, error) {
	data := flags.(*CounterFlags)
	countFee, err := btypes.ParseCoins(data.CountFee)
	if err != nil {
		return nil, err
	}

	ctx := counter.CounterTx{
		Valid: viper.GetBool("valid"),
		Fee:   countFee,
	}
	txBytes := wire.BinaryBytes(ctx)

	return t.App.ReadTxFlags(&data.AppFlags, counter.New().Name(), txBytes, pk)
}
