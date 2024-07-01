package decode

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	gogoproto "github.com/cosmos/gogoproto/proto"
	lru "github.com/hashicorp/golang-lru/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	v1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/core/transaction"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/tx/signing"
)

const (
	// DefaultTxCacheSize is the default capacity of TxCache
	DefaultTxCacheSize uint = 1000
)

// DecodedTx contains the decoded transaction, its signers, and other flags.
type DecodedTx struct {
	DynamicMessages              []proto.Message
	Messages                     []gogoproto.Message
	Tx                           *v1beta1.Tx
	TxRaw                        *v1beta1.TxRaw
	Signers                      [][]byte
	TxBodyHasUnknownNonCriticals bool

	// Cache for hash and full bytes
	cachedHash   [32]byte
	cachedBytes  []byte
	cachedHashed bool
}

var _ transaction.Tx = &DecodedTx{}

type gogoProtoCodec interface {
	Unmarshal([]byte, gogoproto.Message) error
}

// cacheEntry wraps a decoded transaction and an error
type cacheEntry struct {
	decodedTx *DecodedTx
	err       error
}

// Decoder contains the dependencies required for decoding transactions.
type Decoder struct {
	signingCtx *signing.Context
	codec      gogoProtoCodec
	txCache    *lru.Cache[string, cacheEntry]
}

// Options are options for creating a Decoder.
type Options struct {
	SigningContext *signing.Context
	ProtoCodec     gogoProtoCodec
	TxCacheSize    uint
}

// NewDecoder creates a new Decoder for decoding transactions.
func NewDecoder(options Options) (*Decoder, error) {
	if options.SigningContext == nil {
		return nil, errors.New("signing context is required")
	}
	if options.ProtoCodec == nil {
		return nil, errors.New("proto codec is required for unmarshalling gogoproto messages")
	}

	txCacheSize := options.TxCacheSize
	if options.TxCacheSize == 0 {
		txCacheSize = DefaultTxCacheSize
	}

	// Create a new LRU tx decoder cache
	txCache, err := lru.New[string, cacheEntry](int(txCacheSize))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tx decoder cache: %w", err)
	}

	return &Decoder{
		signingCtx: options.SigningContext,
		codec:      options.ProtoCodec,
		txCache:    txCache,
	}, nil
}

// Decode decodes raw protobuf encoded transaction bytes into a DecodedTx.
func (d *Decoder) Decode(txBytes []byte) (*DecodedTx, error) {
	// generate cache key using txBytes
	cacheKey := *(*string)(unsafe.Pointer(&txBytes))

	// check if the tx is already present the cache
	if entry, found := d.txCache.Get(cacheKey); found {
		return entry.decodedTx, entry.err
	}

	// Make sure txBytes follow ADR-027.
	err := rejectNonADR027TxRaw(txBytes)
	if err != nil {
		return d.returnDecodeError(cacheKey, errorsmod.Wrap(ErrTxDecode, err.Error()))
	}

	var raw v1beta1.TxRaw

	// reject all unknown proto fields in the root TxRaw
	fileResolver := d.signingCtx.FileResolver()
	err = RejectUnknownFieldsStrict(txBytes, raw.ProtoReflect().Descriptor(), fileResolver)
	if err != nil {
		return d.returnDecodeError(cacheKey, errorsmod.Wrap(ErrTxDecode, err.Error()))
	}

	err = proto.Unmarshal(txBytes, &raw)
	if err != nil {
		return d.returnDecodeError(cacheKey, err)
	}

	var body v1beta1.TxBody

	// allow non-critical unknown fields in TxBody
	txBodyHasUnknownNonCriticals, err := RejectUnknownFields(raw.BodyBytes, body.ProtoReflect().Descriptor(), true, fileResolver)
	if err != nil {
		return d.returnDecodeError(cacheKey, errorsmod.Wrap(ErrTxDecode, err.Error()))
	}

	err = proto.Unmarshal(raw.BodyBytes, &body)
	if err != nil {
		return d.returnDecodeError(cacheKey, errorsmod.Wrap(ErrTxDecode, err.Error()))
	}

	var authInfo v1beta1.AuthInfo

	// reject all unknown proto fields in AuthInfo
	err = RejectUnknownFieldsStrict(raw.AuthInfoBytes, authInfo.ProtoReflect().Descriptor(), fileResolver)
	if err != nil {
		return d.returnDecodeError(cacheKey, errorsmod.Wrap(ErrTxDecode, err.Error()))
	}

	err = proto.Unmarshal(raw.AuthInfoBytes, &authInfo)
	if err != nil {
		return d.returnDecodeError(cacheKey, errorsmod.Wrap(ErrTxDecode, err.Error()))
	}

	theTx := &v1beta1.Tx{
		Body:       &body,
		AuthInfo:   &authInfo,
		Signatures: raw.Signatures,
	}

	var (
		signers     [][]byte
		dynamicMsgs []proto.Message
		msgs        []gogoproto.Message
	)
	seenSigners := map[string]struct{}{}
	for _, anyMsg := range body.Messages {
		typeURL := strings.TrimPrefix(anyMsg.TypeUrl, "/")

		// unmarshal into dynamic message
		msgDesc, err := fileResolver.FindDescriptorByName(protoreflect.FullName(typeURL))
		if err != nil {
			return d.returnDecodeError(cacheKey, fmt.Errorf("protoFiles does not have descriptor %s: %w", anyMsg.TypeUrl, err))
		}
		dynamicMsg := dynamicpb.NewMessageType(msgDesc.(protoreflect.MessageDescriptor)).New().Interface()
		err = anyMsg.UnmarshalTo(dynamicMsg)
		if err != nil {
			return d.returnDecodeError(cacheKey, err)
		}
		dynamicMsgs = append(dynamicMsgs, dynamicMsg)

		// unmarshal into gogoproto message
		gogoType := gogoproto.MessageType(typeURL)
		if gogoType == nil {
			return d.returnDecodeError(cacheKey, fmt.Errorf("cannot find type: %s", anyMsg.TypeUrl))
		}
		msg := reflect.New(gogoType.Elem()).Interface().(gogoproto.Message)
		err = d.codec.Unmarshal(anyMsg.Value, msg)
		if err != nil {
			return d.returnDecodeError(cacheKey, err)
		}
		msgs = append(msgs, msg)

		// fetch signers with dynamic message
		ss, signerErr := d.signingCtx.GetSigners(dynamicMsg)
		if signerErr != nil {
			return d.returnDecodeError(cacheKey, errorsmod.Wrap(ErrTxDecode, signerErr.Error()))
		}
		for _, s := range ss {
			_, seen := seenSigners[string(s)]
			if seen {
				continue
			}
			signers = append(signers, s)
			seenSigners[string(s)] = struct{}{}
		}
	}

	decodedTx := &DecodedTx{
		Messages:                     msgs,
		DynamicMessages:              dynamicMsgs,
		Tx:                           theTx,
		TxRaw:                        &raw,
		TxBodyHasUnknownNonCriticals: txBodyHasUnknownNonCriticals,
		Signers:                      signers,
	}

	// store the decoded tx in the tx decoder cache
	d.txCache.Add(cacheKey, cacheEntry{decodedTx: decodedTx, err: err})

	return decodedTx, nil
}

func (d *Decoder) returnDecodeError(cacheKey string, err error) (*DecodedTx, error) {
	// store the error in the tx decoder cache
	d.txCache.Add(cacheKey, cacheEntry{decodedTx: nil, err: err})
	return nil, err
}

// Hash implements the interface for the Tx interface.
func (dtx *DecodedTx) Hash() [32]byte {
	if !dtx.cachedHashed {
		dtx.computeHashAndBytes()
	}
	return dtx.cachedHash
}

func (dtx *DecodedTx) GetGasLimit() (uint64, error) {
	if dtx == nil || dtx.Tx == nil || dtx.Tx.AuthInfo == nil || dtx.Tx.AuthInfo.Fee == nil {
		return 0, errors.New("gas limit not available or one or more required fields are nil")
	}
	return dtx.Tx.AuthInfo.Fee.GasLimit, nil
}

func (dtx *DecodedTx) GetMessages() ([]transaction.Msg, error) {
	if dtx == nil || dtx.Messages == nil {
		return nil, errors.New("messages not available or are nil")
	}

	return dtx.Messages, nil
}

func (dtx *DecodedTx) GetSenders() ([][]byte, error) {
	if dtx == nil || dtx.Signers == nil {
		return nil, errors.New("senders not available or are nil")
	}
	return dtx.Signers, nil
}

func (dtx *DecodedTx) Bytes() []byte {
	if !dtx.cachedHashed {
		dtx.computeHashAndBytes()
	}
	return dtx.cachedBytes
}

func (dtx *DecodedTx) computeHashAndBytes() {
	bz, err := proto.Marshal(dtx.TxRaw)
	if err != nil {
		panic(err)
	}

	dtx.cachedBytes = bz
	dtx.cachedHash = sha256.Sum256(bz)
	dtx.cachedHashed = true
}
