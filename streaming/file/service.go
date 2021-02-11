package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/*
The naming schema and data format for the files this service writes out to is as such:

After every `BeginBlock` request a new file is created with the name `block-{N}-begin`, where N is the block number. All
subsequent state changes are written out to this file until the first `DeliverTx` request is received. At the head of these files,
the length-prefixed protobuf encoded `BeginBlock` request is written, and the response is written at the tail.

After every `DeliverTx` request a new file is created with the name `block-{N}-tx-{M}` where N is the block number and M
is the tx number in the block (i.e. 0, 1, 2...). All subsequent state changes are written out to this file until the next
`DeliverTx` request is received or an `EndBlock` request is received. At the head of these files, the length-prefixed protobuf
encoded `DeliverTx` request is written, and the response is written at the tail.

After every `EndBlock` request a new file is created with the name `block-{N}-end`, where N is the block number. All
subsequent state changes are written out to this file until the next `BeginBlock` request is received. At the head of these files,
the length-prefixed protobuf encoded `EndBlock` request is written, and the response is written at the tail.
*/

var _ sdk.StreamingService = &StreamingService{}

// StreamingService is a concrete implementation of StreamingService that writes state changes out to files
type StreamingService struct {
	listeners          map[sdk.StoreKey][]types.WriteListener // the listeners that will be initialized with BaseApp
	srcChan            <-chan []byte                          // the channel that all of the WriteListeners write their data out to
	filePrefix         string                                 // optional prefix for each of the generated files
	writeDir           string                                 // directory to write files into
	marshaller         codec.BinaryMarshaler                  // marshaller used for re-marshalling the ABCI messages to write them out to the destination files
	stateCache         [][]byte                               // cache the protobuf binary encoded StoreKVPairs in the order they are received
	currentBlockNumber int64                                  // the current block number
	currentTxIndex     int64                                  // the index of the current tx
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

// ListenBeginBlock satisfies the Hook interface
// It writes out the received BeginBlock request and response and the resulting state changes out to a file as described
// in the above the naming schema
func (fss *StreamingService) ListenBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	// generate the new file
	dstFile, err := fss.openBeginBlockFile(req)
	if err != nil {
		return err
	}
	// write req to file
	lengthPrefixedReqBytes, err := fss.marshaller.MarshalBinaryLengthPrefixed(&req)
	if err != nil {
		return err
	}
	if _, err = dstFile.Write(lengthPrefixedReqBytes); err != nil {
		return err
	}
	// write all state changes cached for this stage to file
	for _, stateChange := range fss.stateCache {
		if _, err = dstFile.Write(stateChange); err != nil {
			return err
		}
	}
	// reset cache
	fss.stateCache = nil
	// write res to file
	lengthPrefixedResBytes, err := fss.marshaller.MarshalBinaryLengthPrefixed(&res)
	if _, err = dstFile.Write(lengthPrefixedResBytes); err != nil {
		return err
	}
	// close file
	return dstFile.Close()
}

func (fss *StreamingService) openBeginBlockFile(req abci.RequestBeginBlock) (*os.File, error) {
	fss.currentBlockNumber = req.GetHeader().Height
	fss.currentTxIndex = 0
	fileName := fmt.Sprintf("block-%d-begin", fss.currentBlockNumber)
	return os.OpenFile(filepath.Join(fss.writeDir, fileName), os.O_CREATE|os.O_WRONLY, 0600)
}

// ListenDeliverTx satisfies the Hook interface
// It writes out the received DeliverTx request and response and the resulting state changes out to a file as described
// in the above the naming schema
func (fss *StreamingService) ListenDeliverTx(ctx sdk.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	// generate the new file
	dstFile, err := fss.openDeliverTxFile()
	if err != nil {
		return err
	}
	// write req to file
	lengthPrefixedReqBytes, err := fss.marshaller.MarshalBinaryLengthPrefixed(&req)
	if err != nil {
		return err
	}
	if _, err = dstFile.Write(lengthPrefixedReqBytes); err != nil {
		return err
	}
	// write all state changes cached for this stage to file
	for _, stateChange := range fss.stateCache {
		if _, err = dstFile.Write(stateChange); err != nil {
			return err
		}
	}
	// reset cache
	fss.stateCache = nil
	// write res to file
	lengthPrefixedResBytes, err := fss.marshaller.MarshalBinaryLengthPrefixed(&res)
	if _, err = dstFile.Write(lengthPrefixedResBytes); err != nil {
		return err
	}
	// close file
	return dstFile.Close()
}

func (fss *StreamingService) openDeliverTxFile() (*os.File, error) {
	fileName := fmt.Sprintf("block-%d-tx-%d", fss.currentBlockNumber, fss.currentTxIndex)
	fss.currentTxIndex++
	return os.OpenFile(filepath.Join(fss.writeDir, fileName), os.O_CREATE|os.O_WRONLY, 0600)
}

// ListenEndBlock satisfies the Hook interface
// It writes out the received EndBlock request and response and the resulting state changes out to a file as described
// in the above the naming schema
func (fss *StreamingService) ListenEndBlock(ctx sdk.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	// generate the new file
	dstFile, err := fss.openEndBlockFile()
	if err != nil {
		return err
	}
	// write req to file
	lengthPrefixedReqBytes, err := fss.marshaller.MarshalBinaryLengthPrefixed(&req)
	if err != nil {
		return err
	}
	if _, err = dstFile.Write(lengthPrefixedReqBytes); err != nil {
		return err
	}
	// write all state changes cached for this stage to file
	for _, stateChange := range fss.stateCache {
		if _, err = dstFile.Write(stateChange); err != nil {
			return err
		}
	}
	// reset cache
	fss.stateCache = nil
	// write res to file
	lengthPrefixedResBytes, err := fss.marshaller.MarshalBinaryLengthPrefixed(&res)
	if _, err = dstFile.Write(lengthPrefixedResBytes); err != nil {
		return err
	}
	// close file
	return dstFile.Close()
}

func (fss *StreamingService) openEndBlockFile() (*os.File, error) {
	fileName := fmt.Sprintf("block-%d-end", fss.currentBlockNumber)
	return os.OpenFile(filepath.Join(fss.writeDir, fileName), os.O_CREATE|os.O_WRONLY, 0600)
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
