/*
anycompress presents a layer atop tendermint/tm-db that serves to compress *types.Any
transparently. When a value is issued, it is scanned for plausibility of being of the
form of types.Any.TypeURL. If it doesn't match that format, .Set and .Get proceed
transparently as though the high level database were being used.
Otherwise, we replace typeURLs by their respective indices from the typeURLListing registry,
encoded as varint. The replaced value will be prefixed with 2 bytes containing "\\\xfe"; it serves as our unique
marker to perform a sleight of hand replacement when we need to retrieve values.

For cost savings calculation:
* we'll ALWAYS use at least 2 bytes for "\\\xfe"
* for the typeURL.index, in most cases where the types registry has say <64 types, the varint equivalent is 1 byte
If the number of unique registered types is 100 million, the varint equivalent is 4 bytes.

In short, in the worst case of 100 million unique types, the amount of bytes used by this scheme is a maximum of 6
bytes, of which the most nonsensical typeURL is "/A.B" aka 4 bytes, but for the smallest sensible package names e.g "/foo.Biz"
is 8 bytes, and our compression will still save at least 2 bytes. Typical typeURLs might look like "/cosmos.bank.Input" of 18 bytes.
As you can see we always provide big savings!
*/
package anycompress

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	dbm "github.com/tendermint/tm-db"
)

const minTypeURLLen = len("/a.T")

// compressDB transparently compresses types.Any with the backing of an underlying tendermint/tm-db database.
type compressDB struct {
	dbm.DB

	// trie stores the prefixes to match the longest prefixes to disambiguate
	// between typeURLs and the binary data in which they live.
	trie *trie

	typesRegistry map[string]int
	// indexToTypeURL's key is typed as int64 because, insertions are performed exactly
	// once, but retrievals will occur very frequently and decode a varint whose type is int64.
	indexToTypeURL map[int64][]byte
}

// New mimicks the signature that tm-db.New presents and allows the database to be used transparently.
// typeURLListing MUST maintain a deterministic ordering of typeURLs
// Note: whenever we have a global gRPC based typesRegistry, perhaps pass it in here.
func New(name string, backend dbm.BackendType, dir string, typeURLListing []string) (_ dbm.DB, err error) {
	baseDB, err := dbm.NewDB(name, backend, dir)
	if err != nil {
		return nil, err
	}

	cdb := &compressDB{
		typesRegistry:  make(map[string]int),
		trie:           newTrie(),
		DB:             baseDB,
		indexToTypeURL: make(map[int64][]byte),
	}

	for i, typeURL := range typeURLListing {
		cdb.typesRegistry[typeURL] = i
		bTypeURL := []byte(typeURL)
		cdb.trie.set(bTypeURL, bTypeURL)
		cdb.indexToTypeURL[int64(i)] = bTypeURL
	}

	return cdb, nil
}

var _ dbm.DB = (*compressDB)(nil)

var serializeMagic = []byte("\\\xfe")

func (cdb *compressDB) SetSync(key, value []byte) error {
	indices, err := cdb.potentialIndicesForAny(value)
	if errors.Is(err, errNoMatch) {
		return cdb.DB.SetSync(key, value)
	}
	if err != nil {
		return err
	}

	replaceBuf := make([]byte, binary.MaxVarintLen64)
	for _, typeURL := range indices {
		registryIndex, ok := cdb.typesRegistry[string(typeURL)]
		if !ok {
			return fmt.Errorf("no registry index for typeURL %q", typeURL)
		}
		n := binary.PutVarint(replaceBuf, int64(registryIndex))
		replace := make([]byte, len(serializeMagic)+n)
		ni := copy(replace, serializeMagic)
		copy(replace[ni:], replaceBuf[:n])
		value = bytes.ReplaceAll(value, typeURL, replace)
	}
	return cdb.DB.SetSync(key, value)
}

func (cdb *compressDB) Set(key, value []byte) error {
	return cdb.SetSync(key, value)
}

type unfurlingIterator struct {
	dbm.Iterator
	cdb *compressDB
}

var _ dbm.Iterator = (*unfurlingIterator)(nil)

func (ufi *unfurlingIterator) Value() (value []byte) {
	compressed := ufi.Iterator.Value()
	uncompressed, err := ufi.cdb.unfurlOrReturnValue(compressed)
	if err != nil {
		panic(err)
	}
	return uncompressed
}

func (cdb *compressDB) unfurlOrReturnValue(compressed []byte) ([]byte, error) {
	if !bytes.Contains(compressed, serializeMagic) {
		return compressed, nil
	}

	unfurled := make([]byte, 0, len(compressed))
	for len(compressed) > 0 {
		// Find and replace all occurences of: serializedMagic + varint(typeURL index).
		index := bytes.Index(compressed, serializeMagic)

		// No more occurences available, so bail out.
		if index < 0 {
			break
		}

		// Otherwise, decode the registryIndex and then retrieve its associated typeURL.
		registryIndex, n := binary.Varint(compressed[index+len(serializeMagic):])
		if n <= 0 {
			return nil, fmt.Errorf("failed to varint parse value at index: %d", index)
		}
		typeURL, ok := cdb.indexToTypeURL[registryIndex]
		if !ok {
			return nil, fmt.Errorf("could not find a corresponding typeURL for registry index: %d", registryIndex)
		}

		unfurled = append(unfurled, compressed[:index]...)
		unfurled = append(unfurled, typeURL...)
		compressed = compressed[index+len(serializeMagic)+n:]
	}
	if len(compressed) > 0 {
		unfurled = append(unfurled, compressed...)
	}
	return unfurled, nil
}

func (cdb *compressDB) Get(key []byte) ([]byte, error) {
	got, err := cdb.DB.Get(key)
	if err != nil {
		return nil, err
	}
	return cdb.unfurlOrReturnValue(got)
}

func (cdb *compressDB) Iterator(start, end []byte) (dbm.Iterator, error) {
	ri, err := cdb.DB.Iterator(start, end)
	if err != nil {
		return ri, err
	}
	return &unfurlingIterator{Iterator: ri, cdb: cdb}, nil
}

func (cdb *compressDB) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	ri, err := cdb.DB.ReverseIterator(start, end)
	if err != nil {
		return ri, err
	}
	return &unfurlingIterator{Iterator: ri, cdb: cdb}, nil
}

func (cdb *compressDB) potentialIndicesForAny(b []byte) (indices [][]byte, err error) {
	defer func() {
		if len(indices) > 0 && errors.Is(err, errNoMatch) {
			err = nil
		}
	}()
	for i := 0; i < len(b); {
		index := bytes.IndexByte(b[i:], '/')
		if index < 0 {
			return
		}

		// Otherwise plausibly could be a match.
		typeURL, index, rerr := cdb.findClosestTypeURL(b[index:])
		if rerr != nil {
			err = rerr
			return
		}
		if index == -1 || len(typeURL) < minTypeURLLen {
			i++
			continue
		}

		indices = append(indices, typeURL)
		i = index + len(typeURL)
	}
	return
}

func (cdb *compressDB) findClosestTypeURL(b []byte) ([]byte, int, error) {
	if len(b) == 0 {
		return nil, -1, errNoMatch
	}

	index := -1
	for index = 0; index < len(b); index++ {
		if trieIndex(b[index]) == -1 {
			break
		}
	}

	// Now check if the typesRegistry has this type.
	typeURLIndex, ok := cdb.typesRegistry[string(b[:index])]
	if ok {
		return b[:index], typeURLIndex, nil
	}

	// Now to confirm if this was a false positive, we've got to find the closest prefix.
	longestPrefix, i := cdb.trie.longestPrefix(b[:index])
	if longestPrefix == nil {
		return nil, -1, errNoMatch
	}
	if longestPrefix.value == nil {
		return nil, -1, errNoMatch
	}
	return longestPrefix.value, i, nil
}
