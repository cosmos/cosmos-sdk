package grpc

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/fkocik/fsnotify"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/state_file_server/config"
	pb "github.com/cosmos/cosmos-sdk/state_file_server/grpc/v1beta"
	"github.com/cosmos/cosmos-sdk/store/streaming/file"
	"github.com/cosmos/cosmos-sdk/store/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

type StateFileBackend struct {
	conf config.StateServerBackendConfig

	codec *codec.ProtoCodec
	logger log.Logger
	trimPrefix string
}

func NewStateFileBackend(conf config.StateServerBackendConfig, codec *codec.ProtoCodec, logger log.Logger) *StateFileBackend {
	trimPrefix := "block-"
	if conf.FilePrefix != "" {
		trimPrefix = fmt.Sprintf("%s-%s", conf.FilePrefix, trimPrefix)
	}
	return &StateFileBackend{
		conf: conf,
		codec: codec,
		trimPrefix: trimPrefix,
		logger: logger,
	}
}

func (sfb *StateFileBackend) Stream(req *pb.StreamRequest, res chan <-*pb.StreamResponse, doneChan chan <-struct{}) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		close(doneChan)
		return err
	}
	go func() {
		defer w.Close()
		defer close(doneChan)
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					continue
				}
				if event.Op == fsnotify.CloseWrite {
					fileName := event.Name
					readFlag := false
					switch {
					case strings.Contains(fileName, "begin") && req.BeginBlock:
						readFlag = true
						res <- sfb.formBeginBlockResponse(fileName)
					case strings.Contains(fileName, "end") && req.EndBlock:
						readFlag = true
						res <- sfb.formBeginBlockResponse(fileName)
					case strings.Contains(fileName, "tx") && req.DeliverTx:
						readFlag = true
						res <- sfb.formBeginBlockResponse(fileName)
					default:
					}
					if !sfb.conf.Persist && readFlag {
						if err := os.Remove(filepath.Join(sfb.conf.ReadDir, fileName)); err != nil {
							sfb.logger.Error("unable to remove state change file", "err", err)
						}
					}
				}
			case err, ok := <-w.Errors:
				if !ok {
					continue
				}
				sfb.logger.Error("fsnotify watcher error", "err", err)
			}
		}
	}()
	return nil
}

func (sfb *StateFileBackend) BackFill(req *pb.StreamRequest, res chan <-*pb.StreamResponse, doneChan chan <-struct{}) error {
	f, err := os.Open(sfb.conf.ReadDir)
	if err != nil {
		return err
	}
	files, err := f.Readdir(-1)
	if err != nil {
		return err
	}
	sort.Sort(filesByTimeModified(files))
	go func() {
		defer close(doneChan)
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			fileName := f.Name()
			readFlag := false
			switch {
			case strings.Contains(fileName, "begin") && req.BeginBlock:
				readFlag = true
				res <- sfb.formBeginBlockResponse(fileName)
			case strings.Contains(fileName, "end") && req.EndBlock:
				readFlag = true
				res <- sfb.formBeginBlockResponse(fileName)
			case strings.Contains(fileName, "tx") && req.DeliverTx:
				readFlag = true
				res <- sfb.formBeginBlockResponse(fileName)
			default:
			}
			if !sfb.conf.Persist && readFlag {
				if err := os.Remove(filepath.Join(sfb.conf.ReadDir, fileName)); err != nil {
					sfb.logger.Error("unable to remove state change file", "err", err)
				}
			}
		}
	}()
	return nil
}

type filesByTimeModified []os.FileInfo

func (fs filesByTimeModified) Len() int {
	return len(fs)
}

func (fs filesByTimeModified) Swap(i, j int) {
	fs[i], fs[j] = fs[j], fs[i]
}

func (fs filesByTimeModified) Less(i, j int) bool {
	return fs[i].ModTime().Before(fs[j].ModTime())
}

func (sfb *StateFileBackend) formBeginBlockResponse(fileName string) *pb.StreamResponse {
	res := new(pb.StreamResponse)
	res.ChainId = sfb.conf.ChainID
	blockHeightStr := string(strings.TrimPrefix(fileName, sfb.trimPrefix)[0])
	blockHeight, err := strconv.ParseInt(blockHeightStr, 10, 64)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	res.Height = blockHeight
	fileBytes, err := ioutil.ReadFile(filepath.Join(sfb.conf.ReadDir, fileName))
	if err != nil {
		res.Err = err.Error()
		return res
	}
	messageBytes, err := file.SegmentBytes(fileBytes)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	if len(messageBytes) < 2 {
		res.Err = fmt.Sprintf("expected at least two protobuf messages, got %d", len(messageBytes))
		return res
	}
	beginBlockReq := new(abci.RequestBeginBlock)
	if err := sfb.codec.Unmarshal(messageBytes[0], beginBlockReq); err != nil {
		res.Err = err.Error()
		return res
	}
	beginBlockRes := new(abci.ResponseBeginBlock)
	if err := sfb.codec.Unmarshal(messageBytes[len(messageBytes)-1], beginBlockRes); err != nil {
		res.Err = err.Error()
		return res
	}
	kvPairs := make([]*types.StoreKVPair, len(messageBytes[1:len(messageBytes)-2]))
	for i := 1; i < len(messageBytes) - 1; i++ {
		kvPair := new(types.StoreKVPair)
		if err := sfb.codec.Unmarshal(messageBytes[i], kvPair); err != nil {
			res.Err = err.Error()
			return res
		}
		kvPairs[i-1] = kvPair
	}
	res.BeginBlockPayload = &pb.BeginBlockPayload{
		Request: beginBlockReq,
		Response: beginBlockRes,
		StateChanges: kvPairs,
	}
	return res
}

func (sfb *StateFileBackend) formDeliverTxResponse(fileName string) *pb.StreamResponse {
	res := new(pb.StreamResponse)
	res.ChainId = sfb.conf.ChainID
	blockHeightStr := string(strings.TrimPrefix(fileName, sfb.trimPrefix)[0])
	blockHeight, err := strconv.ParseInt(blockHeightStr, 10, 64)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	res.Height = blockHeight
	fileBytes, err := ioutil.ReadFile(filepath.Join(sfb.conf.ReadDir, fileName))
	if err != nil {
		res.Err = err.Error()
		return res
	}
	messageBytes, err := file.SegmentBytes(fileBytes)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	if len(messageBytes) < 2 {
		res.Err = fmt.Sprintf("expected at least two protobuf messages, got %d", len(messageBytes))
		return res
	}
	deliverTxReq := new(abci.RequestDeliverTx)
	if err := sfb.codec.Unmarshal(messageBytes[0], deliverTxReq); err != nil {
		res.Err = err.Error()
		return res
	}
	deliverTxRes := new(abci.ResponseDeliverTx)
	if err := sfb.codec.Unmarshal(messageBytes[len(messageBytes)-1], deliverTxRes); err != nil {
		res.Err = err.Error()
		return res
	}
	kvPairs := make([]*types.StoreKVPair, len(messageBytes[1:len(messageBytes)-2]))
	for i := 1; i < len(messageBytes) - 1; i++ {
		kvPair := new(types.StoreKVPair)
		if err := sfb.codec.Unmarshal(messageBytes[i], kvPair); err != nil {
			res.Err = err.Error()
			return res
		}
		kvPairs[i-1] = kvPair
	}
	res.DeliverTxPayload = &pb.DeliverTxPayload{
		Request: deliverTxReq,
		Response: deliverTxRes,
		StateChanges: kvPairs,
	}
	return res
}

func (sfb *StateFileBackend) formEndBlockResponse(fileName string) *pb.StreamResponse {
	res := new(pb.StreamResponse)
	res.ChainId = sfb.conf.ChainID
	blockHeightStr := string(strings.TrimPrefix(fileName, sfb.trimPrefix)[0])
	blockHeight, err := strconv.ParseInt(blockHeightStr, 10, 64)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	res.Height = blockHeight
	fileBytes, err := ioutil.ReadFile(filepath.Join(sfb.conf.ReadDir, fileName))
	if err != nil {
		res.Err = err.Error()
		return res
	}
	messageBytes, err := file.SegmentBytes(fileBytes)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	if len(messageBytes) < 2 {
		res.Err = fmt.Sprintf("expected at least two protobuf messages, got %d", len(messageBytes))
		return res
	}
	endBlockReq := new(abci.RequestEndBlock)
	if err := sfb.codec.Unmarshal(messageBytes[0], endBlockReq); err != nil {
		res.Err = err.Error()
		return res
	}
	endBlockRes := new(abci.ResponseEndBlock)
	if err := sfb.codec.Unmarshal(messageBytes[len(messageBytes)-1], endBlockRes); err != nil {
		res.Err = err.Error()
		return res
	}
	kvPairs := make([]*types.StoreKVPair, len(messageBytes[1:len(messageBytes)-2]))
	for i := 1; i < len(messageBytes) - 1; i++ {
		kvPair := new(types.StoreKVPair)
		if err := sfb.codec.Unmarshal(messageBytes[i], kvPair); err != nil {
			res.Err = err.Error()
			return res
		}
		kvPairs[i-1] = kvPair
	}
	res.EndBlockPayload = &pb.EndBlockPayload{
		Request: endBlockReq,
		Response: endBlockRes,
		StateChanges: kvPairs,
	}
	return res
}