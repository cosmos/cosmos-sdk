package address

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/internal/conv"
	sdkAddress "github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var errEmptyAddress = errors.New("empty address string is not allowed")

type Option func(*Options)

type Options struct {
	mu    *sync.Mutex
	cache *simplelru.LRU
}

func WithLRU(cache *simplelru.LRU) Option {
	return func(o *Options) {
		o.cache = cache
	}
}

func WithMutex(mu *sync.Mutex) Option {
	return func(o *Options) {
		o.mu = mu
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

func NewBech32Codec(prefix string, opts ...Option) address.Codec {
	options := Options{}
	for _, optionFn := range opts {
		optionFn(&options)
	}

	ac := Bech32Codec{prefix}
	if options.mu == nil || options.cache == nil {
		return ac
	}

	return cachedBech32Codec{
		codec: ac,
		cache: options.cache,
		mu:    options.mu,
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

	key := conv.UnsafeBytesToStr(bz)
	cbc.mu.Lock()
	defer cbc.mu.Unlock()

	addrs, ok := cbc.cache.Get(key)
	if !ok {
		addrs = make(map[string]string)
		cbc.cache.Add(key, addrs)
	}

	addrMap, ok := addrs.(map[string]string)
	if !ok {
		return "", fmt.Errorf("cache contains non-map[string]string value for key %s", key)
	}

	addr, ok := addrMap[cbc.codec.Bech32Prefix]
	if !ok {
		var err error
		addr, err = cbc.codec.BytesToString(bz)
		if err != nil {
			return "", err
		}
		addrMap[cbc.codec.Bech32Prefix] = addr
	}

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
