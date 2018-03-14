package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IBCMapper struct {
	ibcKey sdk.StoreKey
}

func NewIBCMapper(ibcKey sdk.StoreKey) IBCMapper {
	return IBCMapper{
		ibcKey: ibcKey,
	}
}

func GetIngressKey(srcChain string) []byte {
	return []byte(fmt.Sprintf("%s", srcChain))
}

func GetEgressKey(destChain string, index int64) []byte {
	return []byte(fmt.Sprintf("%s/%d", destChain, index))
}
