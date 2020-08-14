package anycompress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gofuzz"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/proto/crypto/keys"
	tmtypes "github.com/tendermint/tendermint/proto/types"
	"github.com/tendermint/tendermint/proto/version"
	dbm "github.com/tendermint/tm-db"
)

func mustAny(msg proto.Message) (*types.Any, []byte) {
	any := new(types.Any)
	any.Pack(msg)
	anyAsValue, err := proto.Marshal(any)
	if err != nil {
		panic(err)
	}
	return any, anyAsValue
}

func TestCompressDB(t *testing.T) {
	anyGarfield, _ := mustAny(&testdata.Cat{
		Moniker: "Garfield",
		Lives:   9,
	})
	any, anyAsValue := mustAny(anyGarfield)

	typesRegistry := []string{anyGarfield.TypeUrl, any.TypeUrl}
	db, err := New("inmem", dbm.MemDBBackend, "", typesRegistry)
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

	if err := db.Set([]byte("foo"), anyAsValue); err != nil {
		t.Fatalf("Failed to save any: %v", err)
	}

	gotAnyBlob, err := db.Get([]byte("foo"))
	if err != nil {
		t.Fatalf("Unexpectedly failed to retrieve foo: %v", err)
	}
	if g, w := gotAnyBlob, anyAsValue; !bytes.Equal(g, w) {
		t.Fatalf("any was not transparently gotten\nGot:  %s\nWant: %s", g, w)
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

type descriptorIface interface {
	Descriptor() ([]byte, []int)
}

func seedRegistry(msg descriptorIface) (typeURLs []string) {
	gzippedPb, _ := msg.Descriptor()
	gzr, err := gzip.NewReader(bytes.NewReader(gzippedPb))
	if err != nil {
		panic(err)
	}
	protoBlob, err := ioutil.ReadAll(gzr)
	if err != nil {
		panic(err)
	}

	fdesc := new(descriptor.FileDescriptorProto)
	if err := proto.Unmarshal(protoBlob, fdesc); err != nil {
		panic(err)
	}
	typeURLs = make([]string, 0, len(fdesc.MessageType))
	for _, msgDesc := range fdesc.MessageType {
		protoFullName := fdesc.GetPackage() + "." + msgDesc.GetName()
		typeURLs = append(typeURLs, "/"+protoFullName)
	}
	return typeURLs
}

var pb0 proto.Message
var protoMessage = reflect.TypeOf(&pb0).Elem()
var gzipMapping = make(map[string]bool)

func traverseAllRegistries(typ reflect.Type, memoize map[string]bool) (typeURLs []string) {
	if !typ.Implements(protoMessage) {
		switch typ.Kind() {
		case reflect.Slice, reflect.Array:
			return traverseAllRegistries(typ.Elem(), memoize)

		default:
			ptrType := reflect.New(typ).Type()
			if ptrType.Implements(protoMessage) {
				return traverseAllRegistries(ptrType, memoize)
			}

			for typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
				if typ.Implements(protoMessage) {
					println("CAUGHT A DEFERENCE")
					return traverseAllRegistries(typ, memoize)
				}
			}
			// Nothing else that we can do here.
			return nil
		}
	}

	rv := reflect.New(typ).Elem()
	pbDesc := rv.Interface().(descriptorIface)
	typeURLs = seedRegistry(pbDesc)

	for _, typeURL := range typeURLs {
		protoFullName := typeURL[1:]
		if _, ok := memoize[typeURL]; ok {
			continue
		}
		rt := proto.MessageType(protoFullName)
		memoize[typeURL] = true
		typeURLs = append(typeURLs, traverseAllRegistries(rt, memoize)...)
	}

	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	switch typ.Kind() {
	case reflect.Struct:
		// Now traverse all the members of the element of type.
		for i, n := 0, typ.NumField(); i < n; i++ {
			sf := typ.Field(i)
			typeURLs = append(typeURLs, traverseAllRegistries(sf.Type, memoize)...)
		}
	}

	return
}

func retrieveTypeURLs(msgs ...proto.Message) (typeURLs []string) {
	memoize := make(map[string]bool)
	for _, msg := range msgs {
		rt := reflect.ValueOf(msg).Type()
		_ = traverseAllRegistries(rt, memoize)
	}
	typeURLs = make([]string, 0, len(memoize))
	for typeURL := range memoize {
		typeURLs = append(typeURLs, typeURL)
	}
	sort.Strings(typeURLs)
	return
}

func genRandBlocks(n int) (blocks []*tmtypes.Block) {
	fz := fuzz.New().NilChance(0).Funcs(
		func(h *tmtypes.Header, c fuzz.Continue) {
			c.Fuzz(&h.Version)
			c.Fuzz(&h.Height)
			c.Fuzz(&h.Time)
			c.Fuzz(&h.LastBlockID)
			c.Fuzz(&h.LastCommitHash)
			c.Fuzz(&h.EvidenceHash)
			c.Fuzz(&h.ProposerAddress)
			c.Fuzz(&h.NextValidatorsHash)
			c.Fuzz(&h.ConsensusHash)
			c.Fuzz(&h.AppHash)
			c.Fuzz(&h.ChainID)
		}, func(vc *version.Consensus, c fuzz.Continue) {
			c.Fuzz(&vc.App)
			c.Fuzz(&vc.Block)
		}, func(ct *tmtypes.Commit, c fuzz.Continue) {
			c.Fuzz(&ct.Hash)
		}, func(ct *tmtypes.Data, c fuzz.Continue) {
			c.Fuzz(&ct.Hash)
		}, func(ph *tmtypes.PartSetHeader, c fuzz.Continue) {
			c.Fuzz(&ph.Total)
			c.Fuzz(&ph.Hash)
		}, func(ph *tmtypes.Evidence, c fuzz.Continue) {
			switch c.Intn(11) {
			case 0:
				ev := &tmtypes.Evidence_DuplicateVoteEvidence{}
				c.Fuzz(ev)
				ph.Sum = ev
			case 1:
				ev := &tmtypes.Evidence_ConflictingHeadersEvidence{}
				c.Fuzz(ev)
				ph.Sum = ev
			case 2:
				ev := &tmtypes.Evidence_LunaticValidatorEvidence{}
				c.Fuzz(ev)
				ph.Sum = ev
			case 3:
				ev := &tmtypes.Evidence_PotentialAmnesiaEvidence{}
				c.Fuzz(ev)
				ph.Sum = ev
			case 4:
				ev := &tmtypes.Evidence_MockEvidence{}
				c.Fuzz(ev)
				ph.Sum = ev
			case 5:
				ev := &tmtypes.Evidence_MockRandomEvidence{}
				c.Fuzz(ev)
				ph.Sum = ev
			}
		}, func(ev *tmtypes.DuplicateVoteEvidence, c fuzz.Continue) {
		}, func(pk *keys.PublicKey, c fuzz.Continue) {
			key := ed25519.GenPrivKey()
			pk.Sum = &keys.PublicKey_Ed25519{Ed25519: key.PubKey().Bytes()}
		},
	)

	for i := 0; i < n; i++ {
		block := new(tmtypes.Block)
		fz.Fuzz(block)
		blocks = append(blocks, block)
	}
	return
}

var blocks = genRandBlocks(5000000)

var registry = retrieveTypeURLs(
	new(tmtypes.Block),
	new(tmtypes.DuplicateVoteEvidence), new(tmtypes.ValidatorSet),
	new(tmtypes.ConsensusParams), new(tmtypes.BlockParams),
	new(tmtypes.EvidenceParams), new(tmtypes.ValidatorParams),
	new(tmtypes.EventDataRoundState),
	new(tmtypes.PartSetHeader),
	new(tmtypes.Part), new(tmtypes.BlockID),
	new(tmtypes.Header), new(tmtypes.Data),
	new(tmtypes.Vote), new(tmtypes.Commit),
	new(tmtypes.CommitSig), new(tmtypes.Proposal),
	new(tmtypes.SignedHeader), new(tmtypes.BlockMeta),
)

func TestCompresseDBVsGoLevelDB(t *testing.T) {
	anyCompDB, golevelDB := testCompresseDBVsNativeDB(t, dbm.GoLevelDBBackend)
	defer anyCompDB.Close()
	defer golevelDB.Close()
}

func TestCompresseDBVsBoltDB(t *testing.T) {
	anyCompDB, boltDB := testCompresseDBVsNativeDB(t, dbm.BoltDBBackend)
	defer anyCompDB.Close()
	defer boltDB.Close()
}

func TestCompresseDBVsRocksDB(t *testing.T) {
	anyCompDB, rocksDB := testCompresseDBVsNativeDB(t, dbm.RocksDBBackend)
	defer anyCompDB.Close()
	defer rocksDB.Close()
}

func testCompresseDBVsNativeDB(t *testing.T, backend dbm.BackendType) (_, _ dbm.DB) {
	if testing.Short() {
		t.Skip("This test runs for a long time")
	}

	t.Parallel()

	dir := "./dbdir-" + string(backend)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	anyCompDB, err := New("compress-db", backend, dir, registry)
	if err != nil {
		t.Fatal(err)
	}
	nativeDB, err := dbm.NewDB("native-db", backend, dir)
	if err != nil {
		t.Fatal(err)
	}

	type kv struct {
		key   []byte
		value []byte
	}
	acCh := make(chan *kv, len(blocks))
	nCh := make(chan *kv, len(blocks))

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for kvi := range acCh {
			anyCompDB.SetSync(kvi.key, kvi.value)
		}
	}()
	go func() {
		defer wg.Done()
		for kvi := range nCh {
			nativeDB.SetSync(kvi.key, kvi.value)
		}
	}()

	for i, block := range blocks {
		key := []byte(fmt.Sprintf("%d", i))
		_, anyBlob := mustAny(block)

		kvi := &kv{key, anyBlob}
		acCh <- kvi
		nCh <- kvi
	}
	close(acCh)
	close(nCh)
	wg.Wait()

	return anyCompDB, nativeDB
}
