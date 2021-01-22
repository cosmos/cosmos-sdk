package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/pprof"
	"sort"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

	"github.com/google/gofuzz"

	"github.com/cosmos/cosmos-sdk/anycompress"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	keys "github.com/tendermint/tendermint/proto/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/proto/tendermint/version"
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
	runtime.GC()
	return
}

const iters = 40

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

type kv struct {
	key   []byte
	value []byte
}

func main() {
	nBlocks := flag.Int("n", 100, "the number of blocks to create")
	nType := flag.String("type", "rocksdb", "The kind of database to use, options are: -goleveldb, -rocksdb, -boltdb")
	flag.Parse()

	start := time.Now()
	blocks := genRandBlocks(*nBlocks)

	var kvps []*kv
	for iter := 0; iter < iters; iter++ {
		for i, block := range blocks {
			values := map[string]proto.Message{
				"block":       block,
				"header":      &block.Header,
				"evidence":    &block.Header,
				"last_commit": block.LastCommit,
			}

			for kind, msg := range values {
				key := []byte(fmt.Sprintf("%s-%d-%d", kind, iter, i))
				_, anyBlob := mustAny(msg)
				kvps = append(kvps, &kv{key, anyBlob})
			}
		}
	}

	runtime.GC()

	println("time to generate ", len(kvps), "key-value pairs:", time.Since(start).String())

	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-shutdownCtx.Done():
				return
			case <-time.After(20 * time.Second):
			}

			filename := fmt.Sprintf("cpu-%d", time.Now().Unix())
			cpuF, err := os.Create(filename)
			if err != nil {
				panic(err)
			}
			if err := pprof.StartCPUProfile(cpuF); err != nil {
				panic(err)
			}
			<-time.After(30 * time.Second)
			pprof.StopCPUProfile()
			cpuF.Close()
			println("Wrote CPUProfile to disk", filename)
		}
	}()

	go func() {
		for {
			select {
			case <-shutdownCtx.Done():
				return
			case <-time.After(45 * time.Second):
			}

			filename := fmt.Sprintf("mem-%d", time.Now().Unix())
			memF, err := os.Create(filename)
			if err != nil {
				panic(err)
			}
			runtime.GC()
			if err := pprof.WriteHeapProfile(memF); err != nil {
				panic(err)
			}
			memF.Close()
			println("Wrote MemoryProfile to disk", filename)
		}
	}()

	switch *nType {
	case "rocksdb":
		CompresseDBVsRocksDB(kvps)
	case "goleveldb":
		CompresseDBVsGoLevelDB(kvps)
	case "boltdb":
		CompresseDBVsBoltDB(kvps)
	default:
		log.Fatalf("Unknown database type: %q", *nType)
	}
}

func CompresseDBVsGoLevelDB(kvps []*kv) {
	_, _ = compresseDBVsNativeDB(dbm.GoLevelDBBackend, kvps)
}

func CompresseDBVsBoltDB(kvps []*kv) {
	_, _ = compresseDBVsNativeDB(dbm.BoltDBBackend, kvps)
}

func CompresseDBVsRocksDB(kvps []*kv) {
	_, _ = compresseDBVsNativeDB(dbm.RocksDBBackend, kvps)
}

func compresseDBVsNativeDB(backend dbm.BackendType, kvps []*kv) (_, _ dbm.DB) {
	startTime := time.Now()

	dir := "./dbdir-" + string(backend)
	os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	anyCompDB, err := anycompress.New("compress-db", backend, dir, registry)
	if err != nil {
		panic(err)
	}
	nativeDB, err := dbm.NewDB("native-db", backend, dir)
	if err != nil {
		panic(err)
	}

        n := 200

	var wg sync.WaitGroup
	wg.Add(2)
	var compressSize int64
	var compressTime time.Duration
	go func() {
		defer func() {
			wg.Done()
			compressTime = time.Since(startTime)
		}()
		for i := 0; i < n; i++ {
			for _, kvi := range kvps {
				anyCompDB.Set(kvi.key, kvi.value)
			}
		}
		anyCompDB.SetSync([]byte("a"), []byte("b"))
		var err error
		anyCompDB.Close()
		compressSize, err = walkAndAddFileSizes(filepath.Join(dir, "compress-db") + "*")
		if err != nil {
			panic(err)
		}
	}()

	var nativeSize int64
	var nativeTime time.Duration
	go func() {
		defer func() {
			wg.Done()
			nativeTime = time.Since(startTime)
		}()
		for i := 0; i < n; i++ {
			for _, kvi := range kvps {
				nativeDB.Set(kvi.key, kvi.value)
			}
		}
		nativeDB.SetSync([]byte("xa"), []byte("xb"))
		var err error
		nativeDB.Close()
		nativeSize, err = walkAndAddFileSizes(filepath.Join(dir, "native-db") + "*")
		if err != nil {
			panic(fmt.Sprintf("nativesize computation failed: %v", err))
		}
	}()

	wg.Wait()

	savings := 100 * float64(nativeSize-compressSize) / float64(nativeSize)
	fmt.Printf("%s savings: %.3f%%\nOriginal:      %s (%dB) - %s\nAnyCompressed: %s (%dB) - %s\nTimeSpent: %s\n\n",
		backend, savings, mostReadable(nativeSize), nativeSize, nativeTime, mostReadable(compressSize), compressSize, compressTime,
		time.Since(startTime))

	return anyCompDB, nativeDB
}

func walkAndAddFileSizes(dirPattern string) (total int64, _ error) {
	matches, err := filepath.Glob(dirPattern)
	if err != nil {
		return 0, err
	}
	if len(matches) == 0 {
		return 0, fmt.Errorf("no matches for %q", dirPattern)
	}
	if err := filepath.Walk(matches[0], func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			total += fi.Size()
		}
		return nil
	}); err != nil {
		return total, err
	}
	return total, nil
}

func mostReadable(size int64) string {
	if size < 1<<10 {
		return fmt.Sprintf("%dB", size)
	}
	if size < 1<<20 {
		return fmt.Sprintf("%3fKiB", float64(size)/float64(1<<10))
	}
	if size < 1<<30 {
		return fmt.Sprintf("%.3fMiB", float64(size)/float64(1<<20))
	}
	if size < 1<<40 {
		return fmt.Sprintf("%.3fGiB", float64(size)/float64(1<<30))
	}
	if size < 1<<50 {
		return fmt.Sprintf("%3.fTiB", float64(size)/float64(1<<40))
	}
	return fmt.Sprintf("%.3fPiB", float64(size)/float64(1<<50))
}
