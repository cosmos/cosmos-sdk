package store

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestVerifyMultiStoreQueryProof(t *testing.T) {

	// Create main tree for testing.
	db := dbm.NewMemDB()
	store_i, err := LoadIAVLStore(db, CommitID{}, sdk.PruneNothing)
	store := store_i.(*iavlStore)
	require.Nil(t, err)
	store.Set([]byte("MYKEY"), []byte("MYVALUE"))
	cid := store.Commit()

	/*
		var storeInfos = []storeInfo{
			storeInfo{
				Name: "tree",
				Core: storeCore{
					CommitID: cid,
				},
			},
			storeInfo{
				Name: "otherTree",
				Core: storeCore{
					CommitID: CommitID{
						Version: 689,
						Hash:    []byte("otherHash"),
					},
				},
			},
		}
	*/

	// Get Proof
	res := store.Query(abci.RequestQuery{
		Path:  "/key", // required path to get key/value+proof
		Data:  []byte("MYKEY"),
		Prove: true,
	})
	fmt.Println("result", spew.Sdump(res))
	require.NotNil(t, res.Proof)

	prt := DefaultProofRuntime()
	err = prt.VerifyValue(res.Proof, cid.Hash, "/MYKEY", []byte("MYVALUE"))
	require.Nil(t, err)
}
