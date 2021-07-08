package grpc

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/fkocik/fsnotify"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/streaming/file"
	"github.com/cosmos/cosmos-sdk/store/streaming/file/server/config"
	pb "github.com/cosmos/cosmos-sdk/store/streaming/file/server/v1beta"
	"github.com/cosmos/cosmos-sdk/store/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

// StateFileBackend performs the state file reading and filtering to service Handler requests
type StateFileBackend struct {
	conf       *config.StateFileServerConfig
	codec      *codec.ProtoCodec
	logger     log.Logger
	trimPrefix string
	quitChan   <-chan struct{}
}

// NewStateFileBackend returns a new StateFileBackend
func NewStateFileBackend(conf *config.StateFileServerConfig, codec *codec.ProtoCodec, logger log.Logger, quitChan <-chan struct{}) *StateFileBackend {
	trimPrefix := "block-"
	if conf.FilePrefix != "" {
		trimPrefix = fmt.Sprintf("%s-%s", conf.FilePrefix, trimPrefix)
	}
	return &StateFileBackend{
		conf:       conf,
		codec:      codec,
		trimPrefix: trimPrefix,
		logger:     logger,
		quitChan:   quitChan,
	}
}

// StreamData streams the requested state file data
// this streams new data as it is written to disk
func (sfb *StateFileBackend) StreamData(req *pb.StreamRequest, res chan<- *pb.StreamResponse) (error, <-chan struct{}) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err, nil
	}

	done := make(chan struct{})
	go func() {
		defer w.Close()
		defer close(done)
		for {
			select {
			case event, ok := <-w.Events:
				if !ok || event.Op != fsnotify.CloseWrite {
					continue
				}

				fileName := event.Name
				if sfb.conf.FilePrefix != "" && !strings.HasPrefix(fileName, sfb.conf.FilePrefix) {
					continue
				}

				switch {
				case strings.Contains(fileName, "begin") && req.BeginBlock:
					res <- sfb.formBeginBlockResponse(fileName, req.StoreKeys)
				case strings.Contains(fileName, "tx") && req.DeliverTx:
					res <- sfb.formDeliverTxResponse(fileName, req.StoreKeys)
				case strings.Contains(fileName, "end") && req.EndBlock:
					res <- sfb.formEndBlockResponse(fileName, req.StoreKeys)
				default:
					continue
				}

				if sfb.conf.RemoveAfter {
					if err := os.Remove(filepath.Join(sfb.conf.ReadDir, fileName)); err != nil {
						sfb.logger.Error("unable to remove state change file", "err", err)
					}
				}
			case err, ok := <-w.Errors:
				if !ok {
					continue
				}
				sfb.logger.Error("fsnotify watcher error", "err", err)
			case <-sfb.quitChan:
				sfb.logger.Info("quiting StateFileBackend StreamData process")
				return
			}
		}
	}()
	return nil, done
}

// BackFillData stream the requested state file data
// this stream data that is already written to disk
func (sfb *StateFileBackend) BackFillData(req *pb.StreamRequest, res chan<- *pb.StreamResponse) (error, <-chan struct{}) {
	f, err := os.Open(sfb.conf.ReadDir)
	if err != nil {
		return err, nil
	}
	files, err := f.Readdir(-1)
	if err != nil {
		return err, nil
	}
	sort.Sort(filesByTimeModified(files))

	done := make(chan struct{})
	go func() {
		defer close(done)
		for _, f := range files {
			select { // short circuit if the parent processes are shutting down
			case <-sfb.quitChan:
				sfb.logger.Info("quiting StateFileBackend BackFillData process")
				return
			default:
			}

			if f.IsDir() {
				continue
			}

			fileName := f.Name()
			if sfb.conf.FilePrefix != "" && !strings.HasPrefix(fileName, sfb.conf.FilePrefix) {
				continue
			}

			switch {
			case strings.Contains(fileName, "begin") && req.BeginBlock:
				res <- sfb.formBeginBlockResponse(fileName, req.StoreKeys)
			case strings.Contains(fileName, "tx") && req.DeliverTx:
				res <- sfb.formDeliverTxResponse(fileName, req.StoreKeys)
			case strings.Contains(fileName, "end") && req.EndBlock:
				res <- sfb.formEndBlockResponse(fileName, req.StoreKeys)
			default:
				continue
			}

			if sfb.conf.RemoveAfter {
				if err := os.Remove(filepath.Join(sfb.conf.ReadDir, fileName)); err != nil {
					sfb.logger.Error("unable to remove state change file", "err", err)
				}
			}
		}
	}()
	return nil, done
}

// BeginBlockDataAt returns a BeginBlockPayload for the provided BeginBlockRequest
func (sfb *StateFileBackend) BeginBlockDataAt(ctx context.Context, req *pb.BeginBlockRequest) (*pb.BeginBlockPayload, error) {
	fileName := fmt.Sprintf("block-%d-begin", req.Height)
	if sfb.conf.FilePrefix != "" {
		fileName = fmt.Sprintf("%s-%s", sfb.conf.FilePrefix, fileName)
	}
	return sfb.formBeginBlockPayload(fileName, req.StoreKeys)
}

// DeliverTxDataAt returns a DeliverTxPayload for the provided BeginBlockRequest
func (sfb *StateFileBackend) DeliverTxDataAt(ctx context.Context, req *pb.DeliverTxRequest) (*pb.DeliverTxPayload, error) {
	fileName := fmt.Sprintf("block-%d-tx-%d", req.Height, req.Index)
	if sfb.conf.FilePrefix != "" {
		fileName = fmt.Sprintf("%s-%s", sfb.conf.FilePrefix, fileName)
	}
	return sfb.formDeliverTxPayload(fileName, req.StoreKeys)
}

// EndBlockDataAt returns a EndBlockPayload for the provided EndBlockRequest
func (sfb *StateFileBackend) EndBlockDataAt(ctx context.Context, req *pb.EndBlockRequest) (*pb.EndBlockPayload, error) {
	fileName := fmt.Sprintf("block-%d-end", req.Height)
	if sfb.conf.FilePrefix != "" {
		fileName = fmt.Sprintf("%s-%s", sfb.conf.FilePrefix, fileName)
	}
	return sfb.formEndBlockPayload(fileName, req.StoreKeys)
}

func (sfb *StateFileBackend) formBeginBlockResponse(fileName string, storeKeys []string) *pb.StreamResponse {
	res := new(pb.StreamResponse)
	res.ChainId = sfb.conf.ChainID
	blockHeightStr := string(strings.TrimPrefix(fileName, sfb.trimPrefix)[0])
	blockHeight, err := strconv.ParseInt(blockHeightStr, 10, 64)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	res.Height = blockHeight

	bbp, err := sfb.formBeginBlockPayload(fileName, storeKeys)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	res.BeginBlockPayload = bbp
	return res
}

func (sfb *StateFileBackend) formBeginBlockPayload(fileName string, storeKeys []string) (*pb.BeginBlockPayload, error) {
	fileBytes, err := readFile(sfb.conf.ReadDir, fileName)
	if err != nil {
		return nil, err
	}
	messageBytes, err := file.SegmentBytes(fileBytes)
	if err != nil {
		return nil, err
	}
	if len(messageBytes) < 2 {
		return nil, fmt.Errorf("expected at least two protobuf messages, got %d", len(messageBytes))
	}

	beginBlockReq := new(abci.RequestBeginBlock)
	if err := sfb.codec.Unmarshal(messageBytes[0], beginBlockReq); err != nil {
		return nil, err
	}

	beginBlockRes := new(abci.ResponseBeginBlock)
	if err := sfb.codec.Unmarshal(messageBytes[len(messageBytes)-1], beginBlockRes); err != nil {
		return nil, err
	}

	kvPairs := make([]*types.StoreKVPair, 0, len(messageBytes[1:len(messageBytes)-2]))
	for i := 1; i < len(messageBytes)-1; i++ {
		kvPair := new(types.StoreKVPair)
		if err := sfb.codec.Unmarshal(messageBytes[i], kvPair); err != nil {
			return nil, err
		}
		if listIsEmptyOrContains(storeKeys, kvPair.StoreKey) {
			kvPairs = append(kvPairs, kvPair)
		}
	}

	return &pb.BeginBlockPayload{
		Request:      beginBlockReq,
		Response:     beginBlockRes,
		StateChanges: kvPairs,
	}, nil
}

func (sfb *StateFileBackend) formDeliverTxResponse(fileName string, storeKeys []string) *pb.StreamResponse {
	res := new(pb.StreamResponse)
	res.ChainId = sfb.conf.ChainID
	blockHeightStr := string(strings.TrimPrefix(fileName, sfb.trimPrefix)[0])
	blockHeight, err := strconv.ParseInt(blockHeightStr, 10, 64)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	res.Height = blockHeight

	dtp, err := sfb.formDeliverTxPayload(fileName, storeKeys)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	res.DeliverTxPayload = dtp
	return res
}

func (sfb *StateFileBackend) formDeliverTxPayload(fileName string, storeKeys []string) (*pb.DeliverTxPayload, error) {
	fileBytes, err := readFile(sfb.conf.ReadDir, fileName)
	if err != nil {
		return nil, err
	}
	messageBytes, err := file.SegmentBytes(fileBytes)
	if err != nil {
		return nil, err
	}
	if len(messageBytes) < 2 {
		return nil, fmt.Errorf("expected at least two protobuf messages, got %d", len(messageBytes))
	}

	deliverTxReq := new(abci.RequestDeliverTx)
	if err := sfb.codec.Unmarshal(messageBytes[0], deliverTxReq); err != nil {
		return nil, err
	}

	deliverTxRes := new(abci.ResponseDeliverTx)
	if err := sfb.codec.Unmarshal(messageBytes[len(messageBytes)-1], deliverTxRes); err != nil {
		return nil, err
	}

	kvPairs := make([]*types.StoreKVPair, 0, len(messageBytes[1:len(messageBytes)-2]))
	for i := 1; i < len(messageBytes)-1; i++ {
		kvPair := new(types.StoreKVPair)
		if err := sfb.codec.Unmarshal(messageBytes[i], kvPair); err != nil {
			return nil, err
		}
		if listIsEmptyOrContains(storeKeys, kvPair.StoreKey) {
			kvPairs = append(kvPairs, kvPair)
		}
	}

	return &pb.DeliverTxPayload{
		Request:      deliverTxReq,
		Response:     deliverTxRes,
		StateChanges: kvPairs,
	}, nil
}

func (sfb *StateFileBackend) formEndBlockResponse(fileName string, storeKeys []string) *pb.StreamResponse {
	res := new(pb.StreamResponse)
	res.ChainId = sfb.conf.ChainID
	blockHeightStr := string(strings.TrimPrefix(fileName, sfb.trimPrefix)[0])
	blockHeight, err := strconv.ParseInt(blockHeightStr, 10, 64)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	res.Height = blockHeight

	ebp, err := sfb.formEndBlockPayload(fileName, storeKeys)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	res.EndBlockPayload = ebp
	return res
}

func (sfb *StateFileBackend) formEndBlockPayload(fileName string, storeKeys []string) (*pb.EndBlockPayload, error) {
	fileBytes, err := readFile(sfb.conf.ReadDir, fileName)
	if err != nil {
		return nil, err
	}
	messageBytes, err := file.SegmentBytes(fileBytes)
	if err != nil {
		return nil, err
	}
	if len(messageBytes) < 2 {
		return nil, fmt.Errorf("expected at least two protobuf messages, got %d", len(messageBytes))
	}

	endBlockReq := new(abci.RequestEndBlock)
	if err := sfb.codec.Unmarshal(messageBytes[0], endBlockReq); err != nil {
		return nil, err
	}

	endBlockRes := new(abci.ResponseEndBlock)
	if err := sfb.codec.Unmarshal(messageBytes[len(messageBytes)-1], endBlockRes); err != nil {
		return nil, err
	}

	kvPairs := make([]*types.StoreKVPair, 0, len(messageBytes[1:len(messageBytes)-2]))
	for i := 1; i < len(messageBytes)-1; i++ {
		kvPair := new(types.StoreKVPair)
		if err := sfb.codec.Unmarshal(messageBytes[i], kvPair); err != nil {
			return nil, err
		}
		if listIsEmptyOrContains(storeKeys, kvPair.StoreKey) {
			kvPairs = append(kvPairs, kvPair)
		}
	}

	return &pb.EndBlockPayload{
		Request:      endBlockReq,
		Response:     endBlockRes,
		StateChanges: kvPairs,
	}, nil
}

func readFile(dir, fileName string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(dir, fileName))
}

func listIsEmptyOrContains(list []string, str string) bool {
	if len(list) == 0 {
		return true
	}
	for _, element := range list {
		if element == str {
			return true
		}
	}
	return false
}

type filesByTimeModified []os.FileInfo

// Len satisfies sort.Interface
func (fs filesByTimeModified) Len() int {
	return len(fs)
}

// Swap satisfies sort.Interface
func (fs filesByTimeModified) Swap(i, j int) {
	fs[i], fs[j] = fs[j], fs[i]
}

// Less satisfies sort.Interface
func (fs filesByTimeModified) Less(i, j int) bool {
	return fs[i].ModTime().Before(fs[j].ModTime())
}
