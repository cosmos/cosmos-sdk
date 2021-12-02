package types

import (
	"encoding/hex"
	"encoding/json"
	"math"
	"strings"

	"github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

func (gi GasInfo) String() string {
	bz, _ := codec.MarshalYAML(codec.NewProtoCodec(nil), &gi)
	return string(bz)
}

func (r Result) String() string {
	bz, _ := codec.MarshalYAML(codec.NewProtoCodec(nil), &r)
	return string(bz)
}

func (r Result) GetEvents() Events {
	events := make(Events, len(r.Events))
	for i, e := range r.Events {
		events[i] = Event(e)
	}

	return events
}

// ABCIMessageLogs represents a slice of ABCIMessageLog.
type ABCIMessageLogs []ABCIMessageLog

func NewABCIMessageLog(i uint32, log string, events Events) ABCIMessageLog {
	return ABCIMessageLog{
		MsgIndex: i,
		Log:      log,
		Events:   StringifyEvents(events.ToABCIEvents()),
	}
}

// String implements the fmt.Stringer interface for the ABCIMessageLogs type.
func (logs ABCIMessageLogs) String() (str string) {
	if logs != nil {
		raw, err := json.Marshal(logs)
		if err == nil {
			str = string(raw)
		}
	}

	return str
}

// NewResponseResultTx returns a TxResponse given a ResultTx from tendermint
func NewResponseResultTx(res *coretypes.ResultTx, anyTx *codectypes.Any, timestamp string) *TxResponse {
	if res == nil {
		return nil
	}

	parsedLogs, _ := ParseABCILogs(res.TxResult.Log)

	return &TxResponse{
		TxHash:    res.Hash.String(),
		Height:    res.Height,
		Codespace: res.TxResult.Codespace,
		Code:      res.TxResult.Code,
		Data:      strings.ToUpper(hex.EncodeToString(res.TxResult.Data)),
		RawLog:    res.TxResult.Log,
		Logs:      parsedLogs,
		Info:      res.TxResult.Info,
		GasWanted: res.TxResult.GasWanted,
		GasUsed:   res.TxResult.GasUsed,
		Tx:        anyTx,
		Timestamp: timestamp,
		Events:    res.TxResult.Events,
	}
}

// NewResponseResultBlock returns a BlockResponse given a ResultBlock from CometBFT
func NewResponseResultBlock(res *coretypes.ResultBlock, timestamp string) *cmtproto.Block {
	if res == nil {
		return nil
	}

	if !res.CheckTx.IsOK() {
		return newTxResponseCheckTx(res)
	}

	return newTxResponseDeliverTx(res)
}

func newTxResponseCheckTx(res *ctypes.ResultBroadcastTxCommit) *TxResponse {
	if res == nil {
		return nil
	}

	var txHash string
	if res.Hash != nil {
		txHash = res.Hash.String()
	}

	parsedLogs, _ := ParseABCILogs(res.CheckTx.Log)

	return &TxResponse{
		Height:    res.Height,
		TxHash:    txHash,
		Codespace: res.CheckTx.Codespace,
		Code:      res.CheckTx.Code,
		Data:      strings.ToUpper(hex.EncodeToString(res.CheckTx.Data)),
		RawLog:    res.CheckTx.Log,
		Logs:      parsedLogs,
		Info:      res.CheckTx.Info,
		GasWanted: res.CheckTx.GasWanted,
		GasUsed:   res.CheckTx.GasUsed,
		Events:    res.CheckTx.Events,
	}
}

func newTxResponseDeliverTx(res *ctypes.ResultBroadcastTxCommit) *TxResponse {
	if res == nil {
		return nil
	}

	var txHash string
	if res.Hash != nil {
		txHash = res.Hash.String()
	}

	parsedLogs, _ := ParseABCILogs(res.DeliverTx.Log)

	return &TxResponse{
		Height:    res.Height,
		TxHash:    txHash,
		Codespace: res.DeliverTx.Codespace,
		Code:      res.DeliverTx.Code,
		Data:      strings.ToUpper(hex.EncodeToString(res.DeliverTx.Data)),
		RawLog:    res.DeliverTx.Log,
		Logs:      parsedLogs,
		Info:      res.DeliverTx.Info,
		GasWanted: res.DeliverTx.GasWanted,
		GasUsed:   res.DeliverTx.GasUsed,
		Events:    res.DeliverTx.Events,
	}
}

// NewResponseFormatBroadcastTx returns a TxResponse given a ResultBroadcastTx from tendermint
func NewResponseFormatBroadcastTx(res *coretypes.ResultBroadcastTx) *TxResponse {
	if res == nil {
		return nil
	}

	parsedLogs, _ := ParseABCILogs(res.Log)

	return &TxResponse{
		Code:      res.Code,
		Codespace: res.Codespace,
		Data:      res.Data.String(),
		RawLog:    res.Log,
		Logs:      parsedLogs,
		TxHash:    res.Hash.String(),
	}
}

func (r TxResponse) String() string {
	bz, _ := codec.MarshalYAML(codec.NewProtoCodec(nil), &r)
	return string(bz)
}

// Empty returns true if the response is empty
func (r TxResponse) Empty() bool {
	return r.TxHash == "" && r.Logs == nil
}

func NewSearchTxsResult(totalCount, count, page, limit uint64, txs []*TxResponse) *SearchTxsResult {
	totalPages := calcTotalPages(int64(totalCount), int64(limit))

	return &SearchTxsResult{
		TotalCount: totalCount,
		Count:      count,
		PageNumber: page,
		PageTotal:  uint64(totalPages),
		Limit:      limit,
		Txs:        txs,
	}
}

func NewSearchBlocksResult(totalCount, count, page, limit int64, blocks []*cmtproto.Block) *SearchBlocksResult {
	totalPages := calcTotalPages(totalCount, limit)

	return &SearchBlocksResult{
		TotalCount: totalCount,
		Count:      count,
		PageNumber: page,
		PageTotal:  totalPages,
		Limit:      limit,
		Blocks:     blocks,
	}
}

// ParseABCILogs attempts to parse a stringified ABCI tx log into a slice of
// ABCIMessageLog types. It returns an error upon JSON decoding failure.
func ParseABCILogs(logs string) (res ABCIMessageLogs, err error) {
	err = json.Unmarshal([]byte(logs), &res)
	return res, err
}

var _, _ gogoprotoany.UnpackInterfacesMessage = SearchTxsResult{}, TxResponse{}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
//
// types.UnpackInterfaces needs to be called for each nested Tx because
// there are generally interfaces to unpack in Tx's
func (s SearchTxsResult) UnpackInterfaces(unpacker gogoprotoany.AnyUnpacker) error {
	for _, tx := range s.Txs {
		err := codectypes.UnpackInterfaces(tx, unpacker)
		if err != nil {
			return err
		}
	}
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (r TxResponse) UnpackInterfaces(unpacker gogoprotoany.AnyUnpacker) error {
	if r.Tx != nil {
		var tx HasMsgs
		return unpacker.UnpackAny(r.Tx, &tx)
	}
	return nil
}

// GetTx unpacks the Tx from within a TxResponse and returns it
func (r TxResponse) GetTx() HasMsgs {
	if tx, ok := r.Tx.GetCachedValue().(HasMsgs); ok {
		return tx
	}
	return nil
}

// calculate total pages in an overflow safe manner
func calcTotalPages(totalCount, limit int64) int64 {
	totalPages := int64(0)
	if totalCount != 0 && limit != 0 {
		if totalCount%limit > 0 {
			totalPages = totalCount/limit + 1
		} else {
			totalPages = totalCount / limit
		}
	}
	return totalPages
}
