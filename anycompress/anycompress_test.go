package anycompress

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"reflect"
	"sort"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

	"github.com/google/go-cmp/cmp"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
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
