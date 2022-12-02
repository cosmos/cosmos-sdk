package file

import (
<<<<<<< HEAD
	"errors"
=======
	"bytes"
	"context"
>>>>>>> 1f91ee2ee (fix: state listener observe writes at wrong time (#13516))
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"sync"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
<<<<<<< HEAD
=======
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
>>>>>>> 1f91ee2ee (fix: state listener observe writes at wrong time (#13516))
)

var _ baseapp.StreamingService = &StreamingService{}

// StreamingService is a concrete implementation of StreamingService that writes state changes out to files
type StreamingService struct {
<<<<<<< HEAD
	listeners          map[types.StoreKey][]types.WriteListener // the listeners that will be initialized with BaseApp
	srcChan            <-chan []byte                            // the channel that all the WriteListeners write their data out to
	filePrefix         string                                   // optional prefix for each of the generated files
	writeDir           string                                   // directory to write files into
	codec              codec.BinaryCodec                        // marshaller used for re-marshalling the ABCI messages to write them out to the destination files
	stateCache         [][]byte                                 // cache the protobuf binary encoded StoreKVPairs in the order they are received
	stateCacheLock     *sync.Mutex                              // mutex for the state cache
	currentBlockNumber int64                                    // the current block number
	currentTxIndex     int64                                    // the index of the current tx
	quitChan           chan struct{}                            // channel to synchronize closure
}

// IntermediateWriter is used so that we do not need to update the underlying io.Writer
// inside the StoreKVPairWriteListener everytime we begin writing to a new file
type IntermediateWriter struct {
	outChan chan<- []byte
}

// NewIntermediateWriter create an instance of an intermediateWriter that sends to the provided channel
func NewIntermediateWriter(outChan chan<- []byte) *IntermediateWriter {
	return &IntermediateWriter{
		outChan: outChan,
	}
}

// Write satisfies io.Writer
func (iw *IntermediateWriter) Write(b []byte) (int, error) {
	iw.outChan <- b
	return len(b), nil
=======
	storeListeners []*types.MemoryListener // a series of KVStore listeners for each KVStore
	filePrefix     string                  // optional prefix for each of the generated files
	writeDir       string                  // directory to write files into
	codec          codec.BinaryCodec       // marshaller used for re-marshalling the ABCI messages to write them out to the destination files
	logger         log.Logger

	currentBlockNumber int64
	blockMetadata      types.BlockMetadata
	// if write the metadata file, otherwise only data file is outputted.
	outputMetadata bool
	// if true, when commit failed it will panic and stop the consensus state machine to ensure the
	// eventual consistency of the output, otherwise the error is ignored and have the risk of lossing data.
	stopNodeOnErr bool
	// if true, the file.Sync() is called to make sure the data is persisted onto disk, otherwise it risks lossing data when system crash.
	fsync bool
>>>>>>> 1f91ee2ee (fix: state listener observe writes at wrong time (#13516))
}

// NewStreamingService creates a new StreamingService for the provided writeDir, (optional) filePrefix, and storeKeys
func NewStreamingService(writeDir, filePrefix string, storeKeys []types.StoreKey, c codec.BinaryCodec, logger log.Logger, outputMetadata bool, stopNodeOnErr bool, fsync bool) (*StreamingService, error) {
	// sort storeKeys for deterministic output
	sort.SliceStable(storeKeys, func(i, j int) bool {
		return storeKeys[i].Name() < storeKeys[j].Name()
	})

	listeners := make([]*types.MemoryListener, len(storeKeys))
	// in this case, we are using the same listener for each Store
	for i, key := range storeKeys {
		listeners[i] = types.NewMemoryListener(key)
	}
	// check that the writeDir exists and is writeable so that we can catch the error here at initialization if it is not
	// we don't open a dstFile until we receive our first ABCI message
	if err := isDirWriteable(writeDir); err != nil {
		return nil, err
	}
	return &StreamingService{
		storeListeners: listeners,
		filePrefix:     filePrefix,
		writeDir:       writeDir,
		codec:          c,
		logger:         logger,
		outputMetadata: outputMetadata,
		stopNodeOnErr:  stopNodeOnErr,
		fsync:          fsync,
	}, nil
}

// Listeners satisfies the baseapp.StreamingService interface
// It returns the StreamingService's underlying WriteListeners
// Use for registering the underlying WriteListeners with the BaseApp
func (fss *StreamingService) Listeners() map[types.StoreKey][]types.WriteListener {
	listeners := make(map[types.StoreKey][]types.WriteListener, len(fss.storeListeners))
	for _, listener := range fss.storeListeners {
		listeners[listener.StoreKey()] = []types.WriteListener{listener}
	}
	return listeners
}

// ListenBeginBlock satisfies the baseapp.ABCIListener interface
// It writes the received BeginBlock request and response and the resulting state changes
// out to a file as described in the above the naming schema
<<<<<<< HEAD
func (fss *StreamingService) ListenBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	// generate the new file
	dstFile, err := fss.openBeginBlockFile(req)
	if err != nil {
		return err
	}
	// write req to file
	lengthPrefixedReqBytes, err := fss.codec.MarshalLengthPrefixed(&req)
	if err != nil {
		return err
	}
	if _, err = dstFile.Write(lengthPrefixedReqBytes); err != nil {
		return err
	}
	// write all state changes cached for this stage to file
	fss.stateCacheLock.Lock()
	for _, stateChange := range fss.stateCache {
		if _, err = dstFile.Write(stateChange); err != nil {
			fss.stateCache = nil
			fss.stateCacheLock.Unlock()
			return err
		}
	}
	// reset cache
	fss.stateCache = nil
	fss.stateCacheLock.Unlock()
	// write res to file
	lengthPrefixedResBytes, err := fss.codec.MarshalLengthPrefixed(&res)
	if err != nil {
		return err
	}
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
	if fss.filePrefix != "" {
		fileName = fmt.Sprintf("%s-%s", fss.filePrefix, fileName)
	}
	return os.OpenFile(filepath.Join(fss.writeDir, fileName), os.O_CREATE|os.O_WRONLY, 0o600)
=======
func (fss *StreamingService) ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) (rerr error) {
	fss.blockMetadata.RequestBeginBlock = &req
	fss.blockMetadata.ResponseBeginBlock = &res
	fss.currentBlockNumber = req.Header.Height
	return nil
>>>>>>> 1f91ee2ee (fix: state listener observe writes at wrong time (#13516))
}

// ListenDeliverTx satisfies the baseapp.ABCIListener interface
// It writes the received DeliverTx request and response and the resulting state changes
// out to a file as described in the above the naming schema
<<<<<<< HEAD
func (fss *StreamingService) ListenDeliverTx(ctx sdk.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	// generate the new file
	dstFile, err := fss.openDeliverTxFile()
	if err != nil {
		return err
	}
	// write req to file
	lengthPrefixedReqBytes, err := fss.codec.MarshalLengthPrefixed(&req)
=======
func (fss *StreamingService) ListenDeliverTx(ctx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) (rerr error) {
	fss.blockMetadata.DeliverTxs = append(fss.blockMetadata.DeliverTxs, &types.BlockMetadata_DeliverTx{
		Request:  &req,
		Response: &res,
	})
	return nil
}

// ListenEndBlock satisfies the baseapp.ABCIListener interface
// It writes the received EndBlock request and response and the resulting state changes
// out to a file as described in the above the naming schema
func (fss *StreamingService) ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) (rerr error) {
	fss.blockMetadata.RequestEndBlock = &req
	fss.blockMetadata.ResponseEndBlock = &res
	return nil
}

// ListenEndBlock satisfies the baseapp.ABCIListener interface
func (fss *StreamingService) ListenCommit(ctx context.Context, res abci.ResponseCommit) error {
	err := fss.doListenCommit(ctx, res)
>>>>>>> 1f91ee2ee (fix: state listener observe writes at wrong time (#13516))
	if err != nil {
		fss.logger.Error("Commit listening hook failed", "height", fss.currentBlockNumber, "err", err)
		if fss.stopNodeOnErr {
			return err
		}
	}
<<<<<<< HEAD
	// reset cache
	fss.stateCache = nil
	fss.stateCacheLock.Unlock()
	// write res to file
	lengthPrefixedResBytes, err := fss.codec.MarshalLengthPrefixed(&res)
	if err != nil {
		return err
	}
	if _, err = dstFile.Write(lengthPrefixedResBytes); err != nil {
		return err
	}
	// close file
	return dstFile.Close()
=======
	return nil
>>>>>>> 1f91ee2ee (fix: state listener observe writes at wrong time (#13516))
}

func (fss *StreamingService) doListenCommit(ctx context.Context, res abci.ResponseCommit) (err error) {
	fss.blockMetadata.ResponseCommit = &res

	// write to target files, the file size is written at the beginning, which can be used to detect completeness.
	metaFileName := fmt.Sprintf("block-%d-meta", fss.currentBlockNumber)
	dataFileName := fmt.Sprintf("block-%d-data", fss.currentBlockNumber)
	if fss.filePrefix != "" {
		metaFileName = fmt.Sprintf("%s-%s", fss.filePrefix, metaFileName)
		dataFileName = fmt.Sprintf("%s-%s", fss.filePrefix, dataFileName)
	}

<<<<<<< HEAD
// ListenEndBlock satisfies the baseapp.ABCIListener interface
// It writes the received EndBlock request and response and the resulting state changes
// out to a file as described in the above the naming schema
func (fss *StreamingService) ListenEndBlock(ctx sdk.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	// generate the new file
	dstFile, err := fss.openEndBlockFile()
	if err != nil {
		return err
	}
	// write req to file
	lengthPrefixedReqBytes, err := fss.codec.MarshalLengthPrefixed(&req)
	if err != nil {
		return err
	}
	if _, err = dstFile.Write(lengthPrefixedReqBytes); err != nil {
		return err
	}
	// write all state changes cached for this stage to file
	fss.stateCacheLock.Lock()
	for _, stateChange := range fss.stateCache {
		if _, err = dstFile.Write(stateChange); err != nil {
			fss.stateCache = nil
			fss.stateCacheLock.Unlock()
=======
	if fss.outputMetadata {
		bz, err := fss.codec.Marshal(&fss.blockMetadata)
		if err != nil {
			return err
		}
		if err := writeLengthPrefixedFile(path.Join(fss.writeDir, metaFileName), bz, fss.fsync); err != nil {
>>>>>>> 1f91ee2ee (fix: state listener observe writes at wrong time (#13516))
			return err
		}
	}

	var buf bytes.Buffer
	if err := fss.writeBlockData(&buf); err != nil {
		return err
	}
<<<<<<< HEAD
	if _, err = dstFile.Write(lengthPrefixedResBytes); err != nil {
		return err
	}
	// close file
	return dstFile.Close()
=======
	return writeLengthPrefixedFile(path.Join(fss.writeDir, dataFileName), buf.Bytes(), fss.fsync)
>>>>>>> 1f91ee2ee (fix: state listener observe writes at wrong time (#13516))
}

func (fss *StreamingService) writeBlockData(writer io.Writer) error {
	for _, listener := range fss.storeListeners {
		cache := listener.PopStateCache()
		for i := range cache {
			bz, err := fss.codec.MarshalLengthPrefixed(&cache[i])
			if err != nil {
				return err
			}
			if _, err = writer.Write(bz); err != nil {
				return err
			}
		}
	}
	return nil
}

// Stream satisfies the baseapp.StreamingService interface
func (fss *StreamingService) Stream(wg *sync.WaitGroup) error {
	return nil
}

// Close satisfies the io.Closer interface, which satisfies the baseapp.StreamingService interface
func (fss *StreamingService) Close() error {
	return nil
}

// isDirWriteable checks if dir is writable by writing and removing a file
// to dir. It returns nil if dir is writable.
func isDirWriteable(dir string) error {
	f := path.Join(dir, ".touch")
	if err := os.WriteFile(f, []byte(""), 0o600); err != nil {
		return err
	}
	return os.Remove(f)
}

func writeLengthPrefixedFile(path string, data []byte, fsync bool) (err error) {
	var f *os.File
	f, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return sdkerrors.Wrapf(err, "open file failed: %s", path)
	}
	defer func() {
		// avoid overriding the real error with file close error
		if err1 := f.Close(); err1 != nil && err == nil {
			err = sdkerrors.Wrapf(err, "close file failed: %s", path)
		}
	}()
	_, err = f.Write(sdk.Uint64ToBigEndian(uint64(len(data))))
	if err != nil {
		return sdkerrors.Wrapf(err, "write length prefix failed: %s", path)
	}
	_, err = f.Write(data)
	if err != nil {
		return sdkerrors.Wrapf(err, "write block data failed: %s", path)
	}
	if fsync {
		err = f.Sync()
		if err != nil {
			return sdkerrors.Wrapf(err, "fsync failed: %s", path)
		}
	}
	return
}
