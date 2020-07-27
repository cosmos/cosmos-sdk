/*
anycompress presents a layer atop tendermint/tm-db that serves to compress *types.Any
transparently. When a value is issued, it is scanned for plausibility of being a protobuf
serialized types.Any. If it isn't a protobuf serialized types.Any, .Set and .Get proceed
transparently as though the high level database were being used.
Otherwise, we extract the typeURL, hash it with a 4 byte FNV32 hash, and append the
hash as well as our magic identifier "\xfe\xff" to the value and store to that to the
underlying database. When a .Get is issued, the underlying database is firstly checked
for presence of the unique 8 bytes suffix signature and if it is, we then extract the
hashed FNV32 hash and on the fly extract the appropriate TypeUrl and then return the
protobuf serialized equivalent as if the types.Any had been stored that way. This conserves
*/
package anycompress

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash"
	"hash/fnv"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"google.golang.org/protobuf/encoding/protowire"

	"github.com/cosmos/cosmos-sdk/codec/types"
	dbm "github.com/tendermint/tm-db"
)

// TODO: Examine the plausibility of "\xfe\xff" naturally existing in files.
// UNIX's EOF is "\xff", so our use of "\xfe\xff" might be super rare too.
var magicSuffixForValue = []byte("\xfe\xff")

// compressDB transparently compresses types.Any with the backing of an underlying tendermint/tm-db database.
type compressDB struct {
	dbm.DB

	mu sync.RWMutex

	startTime    time.Time
	path         string
	index        map[string]uint32
	reverseIndex map[uint32]string

	fnvHash                 hash.Hash32
	cancelBackgroundBacking func()
}

// New mimicks the signature that tm-db.New presents and allows the database to be used transparently.
func New(name string, backend dbm.BackendType, dir string) (_ dbm.DB, err error) {
	baseDB := dbm.NewDB(name, backend, dir)

	var cdb *compressDB
	backingFilepath := filepath.Join(dir, ".compressanydb")
	if _, cerr := os.Stat(backingFilepath); cerr == nil {
		// TODO: Perhaps log to the caller that we are re-using and existent path.
		cdb, err = loadedFromFilepath(backingFilepath)
	} else {
		cdb = freshFromRAM(backingFilepath)
	}

	if err != nil {
		return nil, err
	}

	cdb.DB = baseDB
	cdb.fnvHash = fnv.New32()

	// Otherwise, this is the first time we'll be using this path.
	// TODO: Perhaps log to the caller that we are using a fresh path.
	ctx, cancel := context.WithCancel(context.Background())
	cdb.cancelBackgroundBacking = cancel

	// Start the background backing in the background.
	go cdb.flushPeriodically(ctx)

	return cdb, nil
}

type journal struct {
	WriteTimeUnix int64             `json:"w"`
	Index         map[string]uint32 `json:"i"`
}

func freshFromRAM(path string) *compressDB {
	return &compressDB{
		fnvHash:      fnv.New32(),
		path:         path,
		index:        make(map[string]uint32),
		reverseIndex: make(map[uint32]string),
	}
}

func loadedFromFilepath(path string) (*compressDB, error) {
	blob, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	jrnl := new(journal)
	if err := json.Unmarshal(blob, jrnl); err != nil {
		return nil, err
	}

	// Otherwise, we've retrieved our state from disk and let's construct the indices.
	index := jrnl.Index
	if index == nil {
		index = make(map[string]uint32)
	}
	reverseIndex := make(map[uint32]string)
	for url, hash := range index {
		reverseIndex[hash] = url
	}

	startTimeUnix := jrnl.WriteTimeUnix
	if startTimeUnix <= 0 {
		startTimeUnix = time.Now().Unix()
	}
	cdb := &compressDB{
		startTime:    time.Unix(startTimeUnix, 0),
		path:         path,
		reverseIndex: reverseIndex,
	}

	return cdb, nil
}

var _ dbm.DB = (*compressDB)(nil)

func (cdb *compressDB) Get(key []byte) ([]byte, error) {
	value, err := cdb.DB.Get(key)
	if err != nil {
		return nil, err
	}

	if len(value) <= len(magicSuffixForValue)+4 || !bytes.HasSuffix(value, magicSuffixForValue) {
		// Definitely wasn't saved with our magic header.
		return value, nil
	}

	// Otherwise, we now need to retrieve the appropriate TypeURL, as well as
	// value and serialize it as the proto for a *types.Any.
	valueWithSignature := value[:len(magicSuffixForValue)]

	// Retrieve the signature of the types.Any corresponding URL.
	typeURLSignature := binary.BigEndian.Uint32(value[len(valueWithSignature)-4:])
	value = valueWithSignature[:len(value)-4]

	cdb.mu.Lock()
	typeURL, ok := cdb.reverseIndex[typeURLSignature]
	cdb.mu.Unlock()
	if !ok {
		panic(fmt.Sprintf("Unexpectedly could not find the typeURL with fnv hash signature: %d", typeURLSignature))
	}

	// TODO: We can perhaps inline this proto.Marshal if we deem it consumes more resources than necessary.
	any := &types.Any{
		TypeUrl: typeURL,
		Value:   value,
	}
	return proto.Marshal(any)
}

func (cdb *compressDB) DeleteSync(key []byte) error {
	rawValue, err := cdb.DB.Get(key)
	if err != nil {
		return err
	}
	if err := cdb.DB.DeleteSync(key); err != nil {
		return err
	}

	if len(rawValue) <= len(magicSuffixForValue)+4 || !bytes.HasSuffix(rawValue, magicSuffixForValue) {
		// Definitely wasn't saved with our magic header.
		return nil
	}

	// Otherwise, we also need to delete from our caches.
	typeURLSignature := binary.BigEndian.Uint32(rawValue[len(rawValue)-4:][:4])

	cdb.mu.Lock()
	typeURL := cdb.reverseIndex[typeURLSignature]
	delete(cdb.reverseIndex, typeURLSignature)
	delete(cdb.index, typeURL)
	cdb.mu.Unlock()

	return nil
}

func (cdb *compressDB) Set(key, value []byte) error {
	return cdb.SetSync(key, value)
}

func (cdb *compressDB) SetSync(key, value []byte) error {
	typeURL, value, plausiblyAny := isAnyAsPb(value)
	if !plausiblyAny {
		// Pass through.
		return cdb.DB.SetSync(key, value)
	}

	// Otherwise it is, and we've got to hash the typeURL.
	cdb.mu.Lock()
	cdb.fnvHash.Write([]byte(typeURL))
	blobHashSignature := cdb.fnvHash.Sum(nil)
	indexAsUint32 := cdb.fnvHash.Sum32()
	cdb.fnvHash.Reset()
	cdb.index[string(typeURL)] = indexAsUint32
	cdb.reverseIndex[indexAsUint32] = string(typeURL)
	cdb.mu.Unlock()

	preparedValue := make([]byte, len(value)+len(blobHashSignature)+len(magicSuffixForValue))
	n := copy(preparedValue, value)
	n += copy(preparedValue[n:], blobHashSignature)
	copy(preparedValue[n:], magicSuffixForValue)
	return cdb.DB.Set(key, preparedValue)
}

func isAnyAsPb(blob []byte) (typeURL, value []byte, isPlausiblyAny bool) {
	defer func() {
		if !isPlausiblyAny {
			value = blob
		}
	}()
	// We assume that types.Any as a proto message is stored in the blob.
	typeURLTagNum, typeURLWireType, nTypeURL := protowire.ConsumeField(blob)
	if nTypeURL < 0 {
		return
	}
	if typeURLTagNum != 1 {
		return
	}
	if typeURLWireType != protowire.Type(descriptor.FieldDescriptorProto_TYPE_STRING) {
		return
	}
	if nTypeURL >= len(blob)+2 {
		return
	}
	typeURL = blob[2:nTypeURL]
	if !bytes.HasPrefix(typeURL, []byte("/")) {
		return
	}

	// Now let's check the plausibility of the next value being the Value.
	valueTagNum, valueWireType, nValue := protowire.ConsumeField(blob[nTypeURL+2:])
	if nValue < 0 {
		return
	}
	if valueTagNum != 2 {
		return
	}
	if valueWireType != protowire.Type(descriptor.FieldDescriptorProto_TYPE_BYTES) {
		return
	}
	// Plausibly is a types.Any.
	value = blob[nTypeURL+2+2:]
	return typeURL, value, true
}

func (cdb *compressDB) Close() error {
	if fn := cdb.cancelBackgroundBacking; fn != nil {
		fn()
	}
	return cdb.DB.Close()
}

func (cdb *compressDB) getFlushInterval() time.Duration {
	return 2 * time.Minute
}

// flushPeriodically periodically writes known types.URL hashes to disk.
// This ensures that later on, we can have a mapping of FNV32(typeURL) -> signature
// for lookup and compression, so that typeURLs don't consume so much space.
func (cdb *compressDB) flushPeriodically(ctx context.Context) error {
	shouldExit := false
	for shouldExit == false {
		select {
		case <-ctx.Done():
			// Our job is done, return now.
			shouldExit = true

		case <-time.After(cdb.getFlushInterval()):
		}

		if len(cdb.index) == 0 {
			// Nothing to serialize.
			continue
		}

		// Now serialize the entire index to disk.
		blob, err := json.Marshal(cdb.index)
		if err != nil {
			// TODO: Figure out if we should log this, or panic?
			continue
		}
		f, err := os.Create(cdb.path)
		if err != nil {
			// TODO: Figure out if we should log this, or panic?
			return err
		}
		if _, err := f.Write(blob); err != nil {
			// TODO: Figure out if we should log this, or panic?
			continue
		}
	}

	return ctx.Err()
}
