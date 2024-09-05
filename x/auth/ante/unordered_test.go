package ante_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/header"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/auth/ante"
	"cosmossdk.io/x/auth/ante/unorderedtx"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

const gasConsumed = uint64(25)

func TestUnorderedTxDecorator_OrderedTx(t *testing.T) {
	txm := unorderedtx.NewManager(t.TempDir())
	defer func() {
		require.NoError(t, txm.Close())
	}()

	txm.Start()

	suite := SetupTestSuite(t, false)

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, false, time.Time{})
	ctx := sdk.Context{}.WithTxBytes(txBz)

	_, err := chain(ctx, tx, false)
	require.NoError(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_NoTTL(t *testing.T) {
	txm := unorderedtx.NewManager(t.TempDir())
	defer func() {
		require.NoError(t, txm.Close())
	}()

	txm.Start()

	suite := SetupTestSuite(t, false)

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, true, time.Time{})
	ctx := sdk.Context{}.WithTxBytes(txBz)

	_, err := chain(ctx, tx, false)
	require.Error(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_InvalidTTL(t *testing.T) {
	txm := unorderedtx.NewManager(t.TempDir())
	defer func() {
		require.NoError(t, txm.Close())
	}()

	txm.Start()

	suite := SetupTestSuite(t, false)

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, true, time.Now().Add(unorderedtx.DefaultMaxTimeoutDuration+time.Second))
	ctx := sdk.Context{}.WithTxBytes(txBz).WithHeaderInfo(header.Info{Time: time.Now()})
	_, err := chain(ctx, tx, false)
	require.Error(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_AlreadyExists(t *testing.T) {
	txm := unorderedtx.NewManager(t.TempDir())
	defer func() {
		require.NoError(t, txm.Close())
	}()

	txm.Start()

	suite := SetupTestSuite(t, false)

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, true, time.Now().Add(time.Minute))
	ctx := sdk.Context{}.WithTxBytes(txBz).WithHeaderInfo(header.Info{Time: time.Now()}).WithGasMeter(storetypes.NewGasMeter(gasConsumed))

	bz := [32]byte{}
	copy(bz[:], txBz[:32])
	txm.Add(bz, time.Now().Add(time.Minute))

	_, err := chain(ctx, tx, false)
	require.Error(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_ValidCheckTx(t *testing.T) {
	txm := unorderedtx.NewManager(t.TempDir())
	defer func() {
		require.NoError(t, txm.Close())
	}()

	txm.Start()

	suite := SetupTestSuite(t, false)

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, true, time.Now().Add(time.Minute))
	ctx := sdk.Context{}.WithTxBytes(txBz).WithHeaderInfo(header.Info{Time: time.Now()}).WithExecMode(sdk.ExecModeCheck).WithGasMeter(storetypes.NewGasMeter(gasConsumed))

	_, err := chain(ctx, tx, false)
	require.NoError(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_ValidDeliverTx(t *testing.T) {
	txm := unorderedtx.NewManager(t.TempDir())
	defer func() {
		require.NoError(t, txm.Close())
	}()

	txm.Start()

	suite := SetupTestSuite(t, false)

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, true, time.Now().Add(time.Minute))
	ctx := sdk.Context{}.WithTxBytes(txBz).WithHeaderInfo(header.Info{Time: time.Now()}).WithExecMode(sdk.ExecModeFinalize).WithGasMeter(storetypes.NewGasMeter(gasConsumed))

	_, err := chain(ctx, tx, false)
	require.NoError(t, err)

	bz := [32]byte{}
	copy(bz[:], txBz[:32])

	require.True(t, txm.Contains(bz))
}

func genUnorderedTx(t testing.TB, unordered bool, timestamp time.Time) (sdk.Tx, []byte) {
	t.Helper()

	s := SetupTestSuite(t, true)
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	require.NoError(t, s.txBuilder.SetMsgs(msg))

	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)
	s.txBuilder.SetUnordered(unordered)
	s.txBuilder.SetTimeoutTimestamp(timestamp)

	privKeys, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(s.ctx, privKeys, accNums, accSeqs, s.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	txBz, err := ante.TxIdentifier(uint64(timestamp.Unix()), tx)

	require.NoError(t, err)

	return tx, txBz[:]
}

// Benchmark function for genUnorderedTx
func BenchmarkGenUnorderedTxOld(b *testing.B) {
	// tx, _ := genUnorderedTx(b, true, time.Now().Add(time.Minute))
	// b.ResetTimer()
	// for _, iterations := range []int{1000, 10000, 100000, 100000} {
	// 	b.Run("Iterations_"+strconv.Itoa(iterations), func(b *testing.B) {
	// 		for i := 0; i < iterations; i++ {
	// createAllocations(12000000)
	// 			_, err := ante.TxIdentifier(uint64(time.Now().Unix()+int64(i)), tx)
	// 			if err != nil {
	// 				b.Fatal(err)
	// 			}
	// 		}
	// 	})
	// }
}

func BenchmarkGenUnorderedTx(b *testing.B) {
	tx, _ := genUnorderedTx(b, true, time.Now().Add(time.Minute))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ante.TxIdentifier(uint64(time.Now().Unix()+int64(i)), tx)

	}
}

// goos: darwin
// goarch: arm64
// pkg: cosmossdk.io/x/auth/ante
// cpu: Apple M2
// BenchmarkGenUnorderedTx-8   	 5807298	       205.9 ns/op	      88 B/op	       5 allocs/op
// BenchmarkGenUnorderedTx-8   	 5373726	       207.8 ns/op	      88 B/op	       5 allocs/op
// BenchmarkGenUnorderedTx-8   	 5718591	       204.9 ns/op	      88 B/op	       5 allocs/op
// PASS
// ok  	cosmossdk.io/x/auth/ante	2.066s

// goos: darwin
// goarch: arm64
// pkg: cosmossdk.io/x/auth/ante
// cpu: Apple M2
// BenchmarkGenUnorderedTx-8   	 5298655	       225.0 ns/op	     248 B/op	       7 allocs/op
// BenchmarkGenUnorderedTx-8   	 5279116	       226.7 ns/op	     248 B/op	       7 allocs/op
// BenchmarkGenUnorderedTx-8   	 5361398	       221.6 ns/op	     248 B/op	       7 allocs/op
// PASS
// ok  	cosmossdk.io/x/auth/ante	2.230s
