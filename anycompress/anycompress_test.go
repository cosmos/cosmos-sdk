package anycompress

import (
	"bytes"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/google/go-cmp/cmp"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	dbm "github.com/tendermint/tm-db"
)

func TestCompressDB(t *testing.T) {
	db, err := New("inmem", dbm.MemDBBackend, "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Insert a regular value to test pass-through behavior.
	if err := db.Set([]byte("foo"), []byte("bar")); err != nil {
		t.Fatalf("Failed to set a regular value: %v", err)
	}
	got, err := db.Get([]byte("foo"))
	if err != nil {
		t.Fatalf("Unexpected error for an innocent value:%v", err)
	}
	if g, w := got, []byte("bar"); !bytes.Equal(g, w) {
		t.Fatalf("Mismatch on retrieved values\nGot:  %q\nWant: %q", g, w)
	}

	// Now insert a types.Any.
	any := new(types.Any)
	any.Pack(&testdata.Cat{
		Moniker: "Garfield",
		Lives:   9,
	})

	anyAsValue, err := proto.Marshal(any)
	if err != nil {
		t.Fatalf("Unexpectedly failed to proto marshal any: %v", err)
	}

	if err := db.Set([]byte("foo"), anyAsValue); err != nil {
		t.Fatalf("Failed to save any: %v", err)
	}

	gotAnyBlob, err := db.Get([]byte("foo"))
	if err != nil {
		t.Fatalf("Unexpectedly failed to retrieve foo: %v", err)
	}
	if g, w := gotAnyBlob, anyAsValue; !bytes.Equal(g, w) {
		t.Fatalf("any was not transparently gotten\nGot:  % x\nWant: % x", g, w)
	}

	// Let's assert fully that our value is as it would have been inserted.
	checkAny := new(types.Any)
	if err := proto.Unmarshal(gotAnyBlob, checkAny); err != nil {
		t.Fatalf("Failed to unmarshal types.Any: %v", err)
	}
	if diff := cmp.Diff(any, checkAny); diff != "" {
		t.Fatalf("Mismatch after roundtrip deserialization: got - want +\n%s", diff)
	}

	if err := db.DeleteSync([]byte("foo")); err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	// Check again.
	got, err = db.Get([]byte("foo"))
	if err != nil {
		t.Fatalf("Unexpected error for an innocent value:%v", err)
	}
	if got != nil {
		t.Fatalf("Got %q yet expected nil back", got)
	}
}
