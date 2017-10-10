package stack

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/iavl"
	db "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

func makeState() state.SimpleDB {
	// return state.NewMemKVStore()

	return state.NewBonsai(iavl.NewVersionedTree(0, db.NewMemDB()))

	// tree with persistence....
	// tmpDir, err := ioutil.TempDir("", "state-tests")
	// if err != nil {
	// 	panic(err)
	// }
	// db := dbm.NewDB("test-get-dbs", dbm.LevelDBBackendStr, tmpDir)
	// persist := iavl.NewIAVLTree(500, db)
	// return state.NewBonsai(persist)
}

func TestCheckpointer(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	good := writerHand{name: "foo", key: []byte{1, 2}, value: []byte("bar")}
	bad := FailHandler{Err: errors.New("no go")}

	app := New(
		Checkpoint{OnCheck: true},
		writerMid{name: "bing", key: []byte{1, 2}, value: []byte("bang")},
		Checkpoint{OnDeliver: true},
	).Use(
		NewDispatcher(
			WrapHandler(good),
			WrapHandler(bad),
		))

	sdk.TxMapper.RegisterImplementation(RawTx{}, good.Name(), byte(80))

	mid := state.Model{
		Key:   []byte{'b', 'i', 'n', 'g', 0, 1, 2},
		Value: []byte("bang"),
	}
	end := state.Model{
		Key:   []byte{'f', 'o', 'o', 0, 1, 2},
		Value: []byte("bar"),
	}

	cases := []struct {
		// tx to send down the line
		tx sdk.Tx
		// expect no error?
		valid bool
		// models to check afterwards
		toGetCheck []state.Model
		// models to check afterwards
		toGetDeliver []state.Model
	}{
		// everything writen on success
		{
			tx:           NewRawTx([]byte{45, 67}),
			valid:        true,
			toGetCheck:   []state.Model{mid, end},
			toGetDeliver: []state.Model{mid, end},
		},
		// mostly reverted on failure
		{
			tx:           NewFailTx(),
			valid:        false,
			toGetCheck:   []state.Model{},
			toGetDeliver: []state.Model{mid},
		},
	}

	for i, tc := range cases {
		ctx := NewContext("foo", 100, log.NewNopLogger())

		store := makeState()
		_, err := app.CheckTx(ctx, store, tc.tx)
		if tc.valid {
			require.Nil(err, "%+v", err)
		} else {
			require.NotNil(err)
		}
		for _, m := range tc.toGetCheck {
			val := store.Get(m.Key)
			assert.EqualValues(m.Value, val, "%d: %#v", i, m)
		}

		store = makeState()
		_, err = app.DeliverTx(ctx, store, tc.tx)
		if tc.valid {
			require.Nil(err, "%+v", err)
		} else {
			require.NotNil(err)
		}
		for _, m := range tc.toGetDeliver {
			val := store.Get(m.Key)
			assert.EqualValues(m.Value, val, "%d: %#v", i, m)
		}

	}
}
