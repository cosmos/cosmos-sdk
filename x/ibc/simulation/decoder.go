package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
	clientsim "github.com/cosmos/cosmos-sdk/x/ibc/02-client/simulation"
	connectionsim "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/simulation"
	channelsim "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/simulation"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding ibc type.
func NewDecodeStore(cdc codec.BinaryMarshaler, aminoCdc *codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		if res, found := clientsim.NewDecodeStore(aminoCdc, kvA, kvB); found {
			return res
		}

		if res, found := connectionsim.NewDecodeStore(cdc, kvA, kvB); found {
			return res
		}

		if res, found := channelsim.NewDecodeStore(cdc, kvA, kvB); found {
			return res
		}

		panic(fmt.Sprintf("invalid %s key prefix: %s", host.ModuleName, string(kvA.Key)))
	}
}
