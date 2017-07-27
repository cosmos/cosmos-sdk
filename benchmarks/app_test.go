package app

import (
	"fmt"
	"io/ioutil"
	"testing"

	wire "github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/fee"
	"github.com/tendermint/basecoin/modules/nonce"
	"github.com/tendermint/basecoin/modules/roles"
	"github.com/tendermint/basecoin/stack"
)

type BenchApp struct {
	App      *app.Basecoin
	Accounts []*coin.AccountWithKey
	ChainID  string
}

// DefaultHandler - placeholder to just handle sendtx
func DefaultHandler(feeDenom string) basecoin.Handler {
	// use the default stack
	c := coin.NewHandler()
	r := roles.NewHandler()
	d := stack.NewDispatcher(
		c,
		stack.WrapHandler(r),
	)
	return stack.New(
		base.Logger{},
		stack.Recovery{},
		auth.Signatures{},
		base.Chain{},
		nonce.ReplayCheck{},
		roles.NewMiddleware(),
		fee.NewSimpleFeeMiddleware(coin.Coin{feeDenom, 0}, fee.Bank),
	).Use(d)
}

func NewBenchApp(h basecoin.Handler, chainID string, n int,
	persist bool) BenchApp {

	logger := log.NewNopLogger()
	// logger := log.NewFilter(log.NewTMLogger(os.Stdout), log.AllowError())
	// logger = log.NewTracingLogger(logger)

	// TODO: disk writing
	var store *app.Store
	var err error

	if persist {
		tmpDir, _ := ioutil.TempDir("", "bc-app-benchmark")
		store, err = app.NewStore(tmpDir, 500, logger)
	} else {
		store, err = app.NewStore("", 0, logger)
	}
	if err != nil {
		panic(err)
	}

	app := app.NewBasecoin(
		h,
		store,
		logger.With("module", "app"),
	)
	res := app.SetOption("base/chain_id", chainID)
	if res != "Success" {
		panic("cannot set chain")
	}

	// make keys
	money := coin.Coins{{"mycoin", 1234567890}}
	accts := make([]*coin.AccountWithKey, n)
	for i := 0; i < n; i++ {
		accts[i] = coin.NewAccountWithKey(money)
		res := app.SetOption("coin/account", accts[i].MakeOption())
		if res != "Success" {
			panic("can't set account")
		}
	}

	return BenchApp{
		App:      app,
		Accounts: accts,
		ChainID:  chainID,
	}
}

// make a random tx...
func (b BenchApp) makeTx(useFee bool) []byte {
	n := len(b.Accounts)
	sender := b.Accounts[cmn.RandInt()%n]
	recipient := b.Accounts[cmn.RandInt()%n]
	amount := coin.Coins{{"mycoin", 123}}
	tx := coin.NewSendOneTx(sender.Actor(), recipient.Actor(), amount)
	if useFee {
		toll := coin.Coin{"mycoin", 2}
		tx = fee.NewFee(tx, toll, sender.Actor())
	}
	sequence := sender.NextSequence()
	tx = nonce.NewTx(sequence, []basecoin.Actor{sender.Actor()}, tx)
	tx = base.NewChainTx(b.ChainID, 0, tx)
	stx := auth.NewMulti(tx)
	auth.Sign(stx, sender.Key)
	res := wire.BinaryBytes(stx.Wrap())
	return res
}

func BenchmarkMakeTx(b *testing.B) {
	h := DefaultHandler("mycoin")
	app := NewBenchApp(h, "bench-chain", 10, false)
	b.ResetTimer()
	for i := 1; i <= b.N; i++ {
		txBytes := app.makeTx(true)
		if len(txBytes) < 2 {
			panic("cannot commit")
		}
	}
}

func benchmarkTransfers(b *testing.B, app BenchApp, blockSize int, useFee bool) {
	// prepare txs
	txs := make([][]byte, b.N)
	for i := 1; i <= b.N; i++ {
		txBytes := app.makeTx(useFee)
		if len(txBytes) < 2 {
			panic("cannot make bytes")
		}
		txs[i-1] = txBytes
	}

	b.ResetTimer()

	for i := 1; i <= b.N; i++ {
		res := app.App.DeliverTx(txs[i-1])
		if res.IsErr() {
			panic(res.Error())
		}
		if i%blockSize == 0 {
			res := app.App.Commit()
			if res.IsErr() {
				panic("cannot commit")
			}
		}
	}
}

func BenchmarkSimpleTransfer(b *testing.B) {
	benchmarks := []struct {
		accounts  int
		blockSize int
		useFee    bool
		toDisk    bool
	}{
		{100, 10, false, false},
		{100, 10, true, false},
		{100, 200, false, false},
		{100, 200, true, false},
		{10000, 10, false, false},
		{10000, 10, true, false},
		{10000, 200, false, false},
		{10000, 200, true, false},
		{100, 10, false, true},
		{100, 10, true, true},
		{100, 200, false, true},
		{100, 200, true, true},
		{10000, 10, false, true},
		{10000, 10, true, true},
		{10000, 200, false, true},
		{10000, 200, true, true},
	}

	for _, bb := range benchmarks {
		prefix := fmt.Sprintf("%d-%d", bb.accounts, bb.blockSize)
		if bb.useFee {
			prefix += "-fee"
		} else {
			prefix += "-nofee"
		}
		if bb.toDisk {
			prefix += "-persist"
		} else {
			prefix += "-memdb"
		}

		h := DefaultHandler("mycoin")
		app := NewBenchApp(h, "bench-chain", bb.accounts, bb.toDisk)
		b.Run(prefix, func(sub *testing.B) {
			benchmarkTransfers(sub, app, bb.blockSize, bb.useFee)
		})
	}
}
