package file

import (
	"io/ioutil"
	"os"
	"path"
	"sync"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// StreamingService is a concrete implementation of StreamingService that writes state changes out to files
type StreamingService struct {
	listeners  map[sdk.StoreKey][]types.WriteListener // the listeners that will be initialized with BaseApp
	srcChan    <-chan []byte                          // the channel that all of the WriteListeners write their data out to
	filePrefix string                                 // optional prefix for each of the generated files
	writeDir   string                                 // directory to write files into
	dstFile    *os.File                               // the current write output file
	marshaller codec.BinaryMarshaler                  // marshaller used for re-marshalling the ABCI messages to write them out to the destination files
	stateCache [][]byte                               // cache the protobuf binary encoded StoreKVPairs in the order they are received
}

// intermediateWriter is used so that we do not need to update the underlying io.Writer inside the StoreKVPairWriteListener
// everytime we begin writing to a new file
type intermediateWriter struct {
	outChan chan<- []byte
}

// NewIntermediateWriter create an instance of an intermediateWriter that sends to the provided channel
func NewIntermediateWriter(outChan chan<- []byte) *intermediateWriter {
	return &intermediateWriter{
		outChan: outChan,
	}
}

// Write satisfies io.Writer
func (iw *intermediateWriter) Write(b []byte) (int, error) {
	iw.outChan <- b
	return len(b), nil
}

// NewStreamingService creates a new StreamingService for the provided writeDir, (optional) filePrefix, and storeKeys
func NewStreamingService(writeDir, filePrefix string, storeKeys []sdk.StoreKey, m codec.BinaryMarshaler) (*StreamingService, error) {
	listenChan := make(chan []byte, 0)
	iw := NewIntermediateWriter(listenChan)
	listener := types.NewStoreKVPairWriteListener(iw, m)
	listeners := make(map[sdk.StoreKey][]types.WriteListener, len(storeKeys))
	// in this case, we are using the same listener for each Store
	for _, key := range storeKeys {
		listeners[key] = append(listeners[key], listener)
	}
	// check that the writeDir exists and is writeable so that we can catch the error here at initialization if it is not
	// we don't open a dstFile until we receive our first ABCI message
	if err := isDirWriteable(writeDir); err != nil {
		return nil, err
	}
	return &StreamingService{
		listeners:  listeners,
		srcChan:    listenChan,
		filePrefix: filePrefix,
		writeDir:   writeDir,
		marshaller: m,
		stateCache: make([][]byte, 0),
	}, nil
}

// Listeners returns the StreamingService's underlying WriteListeners, use for registering them with the BaseApp
func (fss *StreamingService) Listeners() map[sdk.StoreKey][]types.WriteListener {
	return fss.listeners
}

func (fss *StreamingService) ListenBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) {
	// NOTE: this could either be done synchronously or asynchronously
	// create a new file with the req info according to naming schema
	// write req to file
	// write all state changes cached for this stage to file
	// reset cache
	// write res to file
	// close file
}

func (fss *StreamingService) ListenEndBlock(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) {
	// NOTE: this could either be done synchronously or asynchronously
	// create a new file with the req info according to naming schema
	// write req to file
	// write all state changes cached for this stage to file
	// reset cache
	// write res to file
	// close file
}

func (fss *StreamingService) ListenDeliverTx(ctx sdk.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) {
	// NOTE: this could either be done synchronously or asynchronously
	// create a new file with the req info according to naming schema
	// NOTE: if the tx failed, handle accordingly
	// write req to file
	// write all state changes cached for this stage to file
	// reset cache
	// write res to file
	// close file
}

// Stream spins up a goroutine select loop which awaits length-prefixed binary encoded KV pairs and caches them in the order they were received
func (fss *StreamingService) Stream(wg *sync.WaitGroup, quitChan <-chan struct{}) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-quitChan:
				return
			case by := <-fss.srcChan:
				fss.stateCache = append(fss.stateCache, by)
			}
		}
	}()
}

// isDirWriteable checks if dir is writable by writing and removing a file
// to dir. It returns nil if dir is writable.
func isDirWriteable(dir string) error {
	f := path.Join(dir, ".touch")
	if err := ioutil.WriteFile(f, []byte(""), 0600); err != nil {
		return err
	}
	return os.Remove(f)
}
