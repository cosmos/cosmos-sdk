package store

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"sort"
	"sync"
	"time"

	"github.com/tendermint/iavl"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/snapshot"
	sm "github.com/tendermint/tendermint/state"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	snapshotWorkingQueueSize  = 5
	snapshotToRemoveQueueSize = 5
	snapshotRetry             = 5
)

type incompleteChunkItem struct {
	chunkIdx int
	chunk    *abci.AppStateChunk
	nodePart []byte
}

type chunkItemSorter struct {
	chunkItems []incompleteChunkItem
}

func (cis *chunkItemSorter) Len() int {
	return len(cis.chunkItems)
}

func (cis *chunkItemSorter) Swap(i, j int) {
	cis.chunkItems[i], cis.chunkItems[j] = cis.chunkItems[j], cis.chunkItems[i]
}

func (cis *chunkItemSorter) Less(i, j int) bool {
	return cis.chunkItems[i].chunkIdx < cis.chunkItems[j].chunkIdx
}

type PrefixNodeDB struct {
	startIdxInclusive int64
	endIdxExclusive   int64
	storeName         string
	*iavl.NodeDB
}

type StateSyncHelper struct {
	SnapshotHeights chan int64
	HeightsToDelete chan int64

	logger   log.Logger
	commitMS sdk.CommitMultiStore
	db       dbm.DB
	cdc      *codec.Codec

	manifest            *abci.Manifest
	stateSyncStoreInfos []StoreInfo

	hashesToIdx      map[abci.SHA256Sum]int          // chunkhash -> idx in manifest
	incompleteChunks map[int64][]incompleteChunkItem // node idx -> incomplete chunk items, for caching incomplete nodes temporally
	prefixNodeDBs    []PrefixNodeDB

	storeKeys []sdk.StoreKey

	reloadingMtx sync.RWMutex // guard below fields to make sure no concurrent load snapshot and response snapshot, and they should be updated atomically

	snapshotManager *snapshot.SnapshotManager
}

func NewStateSyncHelper(
	logger log.Logger,
	db dbm.DB,
	cms sdk.CommitMultiStore,
	cdc *codec.Codec) *StateSyncHelper {
	var helper StateSyncHelper
	helper.logger = logger
	helper.db = db
	helper.commitMS = cms
	helper.cdc = cdc

	kvStores := cms.GetCommitKVStores()
	names := make([]string, 0, len(kvStores))
	nameToKey := make(map[string]sdk.StoreKey, len(kvStores))
	for key, store := range cms.GetCommitKVStores() {
		switch store.(type) {
		case *IavlStore:
			nameToKey[key.Name()] = key
			names = append(names, key.Name())
		default:
			// deliberately do nothing other store type doesn't effect app hash
		}
	}
	sort.Strings(names)
	for _, name := range names {
		helper.storeKeys = append(helper.storeKeys, nameToKey[name])
	}

	helper.SnapshotHeights = make(chan int64, snapshotWorkingQueueSize)
	helper.HeightsToDelete = make(chan int64, snapshotToRemoveQueueSize)

	return &helper
}

// Split Init method and NewStateSyncHelper for snapshot command
func (helper *StateSyncHelper) Init(lastBreatheBlockHeight int64) {
	go helper.ReloadSnapshotRoutine(lastBreatheBlockHeight, 0)
	go func() {
		for height := range helper.SnapshotHeights {
			helper.ReloadSnapshotRoutine(height, snapshotRetry)
		}
	}()
	go func() {
		for height := range helper.HeightsToDelete {
			helper.DeleteSnapshot(height)
		}
	}()
}

func (helper *StateSyncHelper) StartRecovery(manifest *abci.Manifest) error {
	helper.logger.Info("start recovery")

	helper.manifest = manifest
	helper.stateSyncStoreInfos = make([]StoreInfo, 0, len(helper.storeKeys))
	helper.hashesToIdx = make(map[abci.SHA256Sum]int, len(manifest.AppStateHashes))
	helper.incompleteChunks = make(map[int64][]incompleteChunkItem, 0)
	helper.prefixNodeDBs = make([]PrefixNodeDB, 0, len(helper.storeKeys))

	idxOfChunk := 0
	for _, h := range manifest.AppStateHashes {
		helper.hashesToIdx[h] = idxOfChunk
		idxOfChunk++
	}

	if len(manifest.NumKeys) != len(helper.storeKeys) {
		return fmt.Errorf("sub store count in manifest %d does not match local %d", len(manifest.NumKeys), len(helper.storeKeys))
	}

	var startIdxForEachStore int64
	for idx, numOfKeys := range manifest.NumKeys {
		db := dbm.NewPrefixDB(helper.db, []byte("s/k:"+helper.storeKeys[idx].Name()+"/"))
		nodeDB := iavl.NewNodeDB(db, 10000)
		helper.prefixNodeDBs = append(helper.prefixNodeDBs,
			PrefixNodeDB{
				startIdxForEachStore,
				startIdxForEachStore + numOfKeys,
				helper.storeKeys[idx].Name(),
				nodeDB})
		startIdxForEachStore += numOfKeys
	}

	return nil
}

func (helper *StateSyncHelper) WriteRecoveryChunk(hash abci.SHA256Sum, chunk *abci.AppStateChunk, isComplete bool) (err error) {
	helper.reloadingMtx.Lock()
	defer helper.reloadingMtx.Unlock()

	if chunk != nil {
		numOfNodes := len(chunk.Nodes)
		nodes := make([]*iavl.Node, 0, numOfNodes)

		helper.logger.Info("start write recovery chunk", "isComplete", isComplete, "hash", fmt.Sprintf("%x", hash), "startIdx", chunk.StartIdx, "numOfNodes", numOfNodes, "chunkCompletion", chunk.Completeness)

		switch chunk.Completeness {
		case 0: // chunk is independent and complete
			for idx := 0; idx < numOfNodes; idx++ {
				node, _ := iavl.MakeNode(chunk.Nodes[idx])
				iavl.Hash(node)
				nodes = append(nodes, node)
			}
		case 1:
			for idx := 0; idx < numOfNodes-1; idx++ {
				if node, err := iavl.MakeNode(chunk.Nodes[idx]); err == nil {
					iavl.Hash(node)
					nodes = append(nodes, node)
				} else {
					return err
				}
			}
			nodeIdx := chunk.StartIdx + int64(numOfNodes-1)
			helper.incompleteChunks[nodeIdx] = append(helper.incompleteChunks[nodeIdx],
				incompleteChunkItem{
					helper.hashesToIdx[hash],
					chunk,
					chunk.Nodes[numOfNodes-1]})
		case 2, 3:
			if numOfNodes != 1 {
				helper.logger.Error("incomplete chunk should has only one node", "hash", hash, "startIdx", chunk.StartIdx, "completeness", chunk.Completeness, "numOfNodes", numOfNodes)
			}

			helper.incompleteChunks[chunk.StartIdx] = append(helper.incompleteChunks[chunk.StartIdx], incompleteChunkItem{helper.hashesToIdx[hash], chunk, chunk.Nodes[0]})
		default:
			helper.logger.Error("unknown completeness status", "hash", hash, "startIdx", chunk.StartIdx, "completeness", chunk.Completeness, "numOfNodes", numOfNodes)
		}

		// write complete nodes right now
		for idx, node := range nodes {
			nodeIdx := chunk.StartIdx + int64(idx)
			helper.saveNode(nodeIdx, node)
		}

		helper.logger.Info("finished write recovery chunk", "isComplete", isComplete, "hash", fmt.Sprintf("%x", hash), "startIdx", chunk.StartIdx, "numOfNodes", numOfNodes, "chunkCompletion", chunk.Completeness)
	}

	if isComplete {
		err = helper.finishCompleteChunkWrite()
	}

	return err
}

func (helper *StateSyncHelper) DeleteSnapshot(height int64) error {
	err := snapshot.ManagerAt(height).Delete()
	helper.logger.Info("deleted snapshot", "height", height, "err", err)
	return err
}

func (helper *StateSyncHelper) finishCompleteChunkWrite() error {
	helper.prepareEmptyStores()
	if err := helper.saveIncompleteChunks(); err != nil {
		return err
	}
	if err := helper.commitDB(); err != nil {
		return err
	}

	return nil
}

func (helper *StateSyncHelper) prepareEmptyStores() {
	for _, nodeDB := range helper.prefixNodeDBs {
		if nodeDB.endIdxExclusive == nodeDB.startIdxInclusive {
			nodeDB.NodeDB.SaveEmptyRoot(helper.manifest.Height, true)
			helper.stateSyncStoreInfos = append(helper.stateSyncStoreInfos, StoreInfo{
				Name: nodeDB.storeName,
				Core: StoreCore{
					CommitID: CommitID{
						Version: helper.manifest.Height,
						Hash:    nil,
					},
				},
			})
		}
	}
}

func (helper *StateSyncHelper) saveIncompleteChunks() error {
	for nodeIdx, chunkItems := range helper.incompleteChunks {
		helper.logger.Debug("processing incomplete node", "nodeIdx", nodeIdx)
		// sort and check chunkItems are valid
		sort.Sort(&chunkItemSorter{chunkItems})

		expectedNodeParts := chunkItems[len(chunkItems)-1].chunkIdx - chunkItems[0].chunkIdx + 1
		if expectedNodeParts != len(chunkItems) {
			return fmt.Errorf("node parts are not complete, should be %d, but have %d, nodeIdx: %d", expectedNodeParts, len(chunkItems), nodeIdx)
		}

		var completeNode bytes.Buffer
		for idx, chunkItem := range chunkItems {
			if idx == 0 {
				if chunkItem.chunk.Completeness != 1 {
					return fmt.Errorf("first node part containing chunk's completeness %d is wrong, should be 1, nodeIdx: %d", chunkItem.chunk.Completeness, nodeIdx)
				}
			} else if idx == len(chunkItems)-1 {
				if chunkItem.chunk.Completeness != 3 {
					return fmt.Errorf("last node part containing chunk's completeness %d is wrong, should be 3, nodeIdx: %d", chunkItem.chunk.Completeness, nodeIdx)
				}
			} else {
				if chunkItem.chunk.Completeness != 2 {
					return fmt.Errorf("middle node part containing chunk's completeness %d is wrong, should be 2, nodeIdx: %d", chunkItem.chunk.Completeness, nodeIdx)
				}
			}
			completeNode.Write(chunkItem.nodePart)
		}

		if node, err := iavl.MakeNode(completeNode.Bytes()); err == nil {
			iavl.Hash(node)
			helper.saveNode(nodeIdx, node)
		} else {
			return err
		}
	}
	return nil
}

func (helper *StateSyncHelper) commitDB() error {
	height := helper.manifest.Height

	// TODO: revisit would it be problem commit too late? would there be memory or performance issue?
	// probably we need commit as soon as store is complete
	for _, db := range helper.prefixNodeDBs {
		db.NodeDB.Commit()
	}

	// simulate setLatestversion key
	batch := helper.db.NewBatch()
	latestBytes, _ := helper.cdc.MarshalBinaryLengthPrefixed(height) // Does not error
	batch.Set([]byte("s/latest"), latestBytes)

	ci := CommitInfo{
		Version:    height,
		StoreInfos: helper.stateSyncStoreInfos,
	}
	if cInfoBytes, err := helper.cdc.MarshalBinaryLengthPrefixed(ci); err != nil {
		return err
	} else {
		cInfoKey := fmt.Sprintf("s/%d", height)
		batch.Set([]byte(cInfoKey), cInfoBytes)
		batch.WriteSync()
		return nil
	}
}

func (helper *StateSyncHelper) saveNode(nodeIdx int64, node *iavl.Node) {
	for _, nodeDB := range helper.prefixNodeDBs {
		if nodeIdx < nodeDB.endIdxExclusive {
			if nodeIdx == nodeDB.startIdxInclusive {
				nodeDB.NodeDB.SaveRoot(node, helper.manifest.Height, true)
				rootHash := iavl.Hash(node)
				helper.stateSyncStoreInfos = append(helper.stateSyncStoreInfos, StoreInfo{
					Name: nodeDB.storeName,
					Core: StoreCore{
						CommitID: CommitID{
							Version: helper.manifest.Height,
							Hash:    rootHash,
						},
					},
				})
				helper.logger.Info("save root hash", "store", nodeDB.storeName, "hash", fmt.Sprintf("%X", rootHash))
			}
			iavl.Hash(node)
			nodeDB.NodeDB.SaveNode(node)
			helper.logger.Debug("saved node to store", "nodeIdx", nodeIdx, "store", nodeDB.storeName)
			break
		}
	}
}

// the method might take quite a while, BETTER to be called concurrently
// so we only do it once a day after breathe block
func (helper StateSyncHelper) ReloadSnapshotRoutine(height int64, retry int) {
	helper.reloadingMtx.Lock()
	defer helper.reloadingMtx.Unlock()

	helper.takeSnapshotImpl(height, retry)
}

func (helper *StateSyncHelper) takeSnapshotImpl(height int64, retry int) {
	defer func() {
		if r := recover(); r != nil {
			log := fmt.Sprintf("recovered: %v\nstack:\n%v", r, string(debug.Stack()))
			helper.logger.Error("failed loading latest snapshot", "err", log)
		}
	}()
	helper.logger.Info("reload latest snapshot", "height", height)

	for {
		if mgr := snapshot.ManagerAt(height); mgr != nil {
			helper.snapshotManager = mgr
			break
		} else {
			helper.logger.Debug("waiting base snapshot manager is initialized")
			time.Sleep(100 * time.Millisecond)
		}
	}

	if helper.snapshotManager.IsFinalized() {
		return
	}

	failed := true
	for failed {
		failed = false
		if state := sm.LoadStateForHeight(helper.snapshotManager.GetStateDB(), height); state == nil {
			helper.logger.Info("expected state has not committed yet", "height", height)
			failed = true
			time.Sleep(1 * time.Second) // Endblocker has notified this reload snapshot,
			// wait for 1 sec after commit finish
			if retry > 0 {
				retry--
				continue
			} else {
				return
			}
		}

		totalKeys := int64(0)
		numKeys := make([]int64, 0, len(helper.storeKeys))
		currChunkNodes := make([][]byte, 0, 40000) // one account leaf node is around 100 bytes according to testnet experiment, non-leaf node should be less, 40000 should be a bit less than 4M
		var currStartIdx int64
		var currChunkTotalBytes int
		for _, key := range helper.storeKeys {
			var currStoreKeys int64
			store := helper.commitMS.GetKVStore(key)
			// TODO: use Iterator method of store interface, no longer rely on implementation of KVStore
			// as we only append storeKeys for IavlStore at constructor, so this type assertion should never fail
			mutableTree := store.(*IavlStore).Tree
			if tree, err := mutableTree.GetImmutable(height); err == nil {
				tree.IterateFirst(func(nodeBytes []byte) {
					nodeBytesLength := len(nodeBytes)

					if currChunkTotalBytes+nodeBytesLength <= abci.ChunkPayloadMaxBytes {
						currChunkNodes = append(currChunkNodes, nodeBytes)
						currChunkTotalBytes += nodeBytesLength
					} else {
						helper.finalizeAppStateChunk(currStartIdx, abci.Complete, currChunkNodes)
						currStartIdx += int64(len(currChunkNodes))
						currChunkNodes = currChunkNodes[:0]
						currChunkTotalBytes = 0

						// One chunk should have AT MOST one incomplete node
						// For a large node, we at most waste one chunk (the last finalized one)
						if nodeBytesLength > abci.ChunkPayloadMaxBytes {
							firstPart := nodeBytes[:abci.ChunkPayloadMaxBytes]
							currChunkNodes = append(currChunkNodes, firstPart)
							helper.finalizeAppStateChunk(currStartIdx, abci.InComplete_First, currChunkNodes)

							startCutIdx := len(firstPart)
							for ; startCutIdx+abci.ChunkPayloadMaxBytes < nodeBytesLength; startCutIdx += abci.ChunkPayloadMaxBytes {
								helper.finalizeAppStateChunk(totalKeys+currStoreKeys, abci.InComplete_Mid, [][]byte{nodeBytes[startCutIdx : startCutIdx+abci.ChunkPayloadMaxBytes]})
							}

							lastPart := nodeBytes[startCutIdx:]
							helper.finalizeAppStateChunk(totalKeys+currStoreKeys, abci.InComplete_Last, [][]byte{lastPart})

							currStartIdx = totalKeys + currStoreKeys + 1
							currChunkNodes = currChunkNodes[:0]
							currChunkTotalBytes = 0
						} else {
							currChunkNodes = append(currChunkNodes, nodeBytes)
							currChunkTotalBytes += nodeBytesLength
						}
					}

					currStoreKeys++
				})
				helper.logger.Info("snapshoted a substore", "storeName", key, "numOfKeys", currStoreKeys)
			} else {
				helper.logger.Error("failed to load immutable tree", "err", err, "store", key)
				return
			}
			totalKeys += currStoreKeys
			numKeys = append(numKeys, currStoreKeys)
		}

		if !failed {
			if len(currChunkNodes) > 0 {
				helper.finalizeAppStateChunk(currStartIdx, abci.Complete, currChunkNodes)
			}
			if err := helper.snapshotManager.SelfFinalize(numKeys); err == nil {
				helper.logger.Info("finish read snapshot chunk", "height", height, "keys", totalKeys)
			} else {
				helper.logger.Error("failed read snapshot chunk", "height", height, "keys", totalKeys, "err", err)
			}
		}
	}
}

func (helper *StateSyncHelper) finalizeAppStateChunk(startIdx int64, completeness uint8, nodes [][]byte) error {
	return helper.snapshotManager.WriteAppStateChunk(&abci.AppStateChunk{startIdx, completeness, nodes})
}
