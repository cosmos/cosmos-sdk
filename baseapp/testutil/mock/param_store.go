package mock

import (
	"context"
	"encoding/json"
	"errors"
	

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"

	

	"github.com/cosmos/cosmos-sdk/baseapp"
	
)

var ParamStoreKey = []byte("paramstore")

type ParamStore struct {
	db *dbm.MemDB
}

var _ baseapp.ParamStore = (*ParamStore)(nil)

func NewMockParamStore(db *dbm.MemDB) *ParamStore {
	return &ParamStore{db: db}
}

func (ps ParamStore) Set(_ context.Context, value cmtproto.ConsensusParams) error {
	bz, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return ps.db.Set(ParamStoreKey, bz)
}

func (ps ParamStore) Has(_ context.Context) (bool, error) {
	return ps.db.Has(ParamStoreKey)
}

func (ps ParamStore) Get(_ context.Context) (cmtproto.ConsensusParams, error) {
	bz, err := ps.db.Get(ParamStoreKey)
	if err != nil {
		return cmtproto.ConsensusParams{}, err
	}

	if len(bz) == 0 {
		return cmtproto.ConsensusParams{}, errors.New("params not found")
	}

	var params cmtproto.ConsensusParams
	if err := json.Unmarshal(bz, &params); err != nil {
		return cmtproto.ConsensusParams{}, err
	}

	return params, nil
}