package store

import (
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// nolint: reexport
type (
	CommitMultiStore = types.CommitMultiStore
)

// nolint: reexport
func NewCommitMultiStore(db dbm.DB) *rootmulti.Store {
	return rootmulti.NewStore(db)
}
func ErrUnknownRequest(msg string) types.Error {
	return types.ErrUnknownRequest(msg)
}
