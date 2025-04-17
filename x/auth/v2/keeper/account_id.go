package keeper

import "github.com/cosmos/cosmos-sdk/types"

type AccountID struct {
	value []byte
}

// gogo proto custom type support
var _ types.CustomProtobufType = &AccountID{}

func (a AccountID) Marshal() ([]byte, error)                 { panic("implement me") }
func (a AccountID) MarshalTo(data []byte) (n int, err error) { panic("implement me") }
func (a AccountID) Unmarshal(data []byte) error              { panic("implement me") }
func (a AccountID) Size() int                                { panic("implement me") }
func (a AccountID) MarshalJSON() ([]byte, error)             { panic("implement me") }
func (a AccountID) UnmarshalJSON(data []byte) error          { panic("implement me") }
