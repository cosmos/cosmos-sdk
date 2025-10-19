// SPDX-License-Identifier: MPL-2.0
// Copyright © 2025 Timechain-Arweave-LunCoSim Contributors

package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/timechain/types"
)

// NewDecodeStore returns a decoder function for the module
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		default:
			panic(fmt.Sprintf("invalid %s key %X", types.ModuleName, kvA.Key))
		}
	}
}
