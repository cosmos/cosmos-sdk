package address

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/golang-lru/simplelru"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/internal/conv"
	sdkAddress "github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// TODO: ideally sdk.GetBech32PrefixValAddr("") should be used but currently there's a cyclical import.
	// 	Once globals are deleted the cyclical import won't happen.
	suffixValAddr  = "valoper"
	suffixConsAddr = "valcons"
)

var (
	errEmptyAddress = errors.New("empty address string is not allowed")
)

// cache variables
var (
	accAddrMu     sync.Mutex
	accAddrCache  *simplelru.LRU
	consAddrMu    sync.Mutex
	consAddrCache *simplelru.LRU
	valAddrMu     sync.Mutex
	valAddrCache  *simplelru.LRU

	isCachingEnabled atomic.Bool
)

func init() {
	var err error
	isCachingEnabled.Store(true)

	// in total the cache size is 61k entries. Key is 32 bytes and value is around 50-70 bytes.
	// That will make around 92 * 61k * 2 (LRU) bytes ~ 11 MB
	if accAddrCache, err = simplelru.NewLRU(60000, nil); err != nil {
		panic(err)
	}
	if consAddrCache, err = simplelru.NewLRU(500, nil); err != nil {
		panic(err)
	}
	if valAddrCache, err = simplelru.NewLRU(500, nil); err != nil {
		panic(err)
	}
}

type Bech32Codec struct {
	Bech32Prefix string
}

type cachedBech32Codec struct {
	codec Bech32Codec
	mu    *sync.Mutex
	cache *simplelru.LRU
}

var (
	_ address.Codec = &Bech32Codec{}
	_ address.Codec = &cachedBech32Codec{}
)

func NewBech32Codec(prefix string) address.Codec {
	ac := Bech32Codec{prefix}
	if !isCachingEnabled.Load() {
		return ac
	}

	lru := accAddrCache
	mu := &accAddrMu
	if strings.HasSuffix(prefix, suffixValAddr) {
		lru = valAddrCache
		mu = &valAddrMu
	} else if strings.HasSuffix(prefix, suffixConsAddr) {
		lru = consAddrCache
		mu = &consAddrMu
	}

	return cachedBech32Codec{
		codec: ac,
		cache: lru,
		mu:    mu,
	}
}

// StringToBytes encodes text to bytes
func (bc Bech32Codec) StringToBytes(text string) ([]byte, error) {
	if len(strings.TrimSpace(text)) == 0 {
		return []byte{}, errEmptyAddress
	}

	hrp, bz, err := bech32.DecodeAndConvert(text)
	if err != nil {
		return nil, err
	}

	if len(bz) > sdkAddress.MaxAddrLen {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "address max length is %d, got %d", sdkAddress.MaxAddrLen, len(bz))
	}

	if hrp != bc.Bech32Prefix {
		return nil, errorsmod.Wrapf(sdkerrors.ErrLogic, "hrp does not match bech32 prefix: expected '%s' got '%s'", bc.Bech32Prefix, hrp)
	}

	return bz, nil
}

// BytesToString decodes bytes to text
func (bc Bech32Codec) BytesToString(bz []byte) (string, error) {
	if len(bz) == 0 {
		return "", nil
	}

	text, err := bech32.ConvertAndEncode(bc.Bech32Prefix, bz)
	if err != nil {
		return "", err
	}

	if len(bz) > sdkAddress.MaxAddrLen {
		return "", errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "address max length is %d, got %d", sdkAddress.MaxAddrLen, len(bz))
	}

	return text, nil
}

func (cbc cachedBech32Codec) BytesToString(bz []byte) (string, error) {
	if len(bz) == 0 {
		return "", nil
	}

	// caches prefix is added to the key to make sure keys are unique in case codecs with different bech32 prefix are defined.
	key := cbc.codec.Bech32Prefix + conv.UnsafeBytesToStr(bz)
	cbc.mu.Lock()
	defer cbc.mu.Unlock()

	if addr, ok := cbc.cache.Get(key); ok {
		return addr.(string), nil
	}

	addr, err := cbc.codec.BytesToString(bz)
	if err != nil {
		return "", err
	}
	cbc.cache.Add(key, addr)

	return addr, nil
}

func (cbc cachedBech32Codec) StringToBytes(text string) ([]byte, error) {
	cbc.mu.Lock()
	defer cbc.mu.Unlock()

	if addr, ok := cbc.cache.Get(text); ok {
		return addr.([]byte), nil
	}

	addr, err := cbc.codec.StringToBytes(text)
	if err != nil {
		return nil, err
	}
	cbc.cache.Add(text, addr)

	return addr, nil
}
