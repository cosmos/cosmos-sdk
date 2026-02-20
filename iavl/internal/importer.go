package internal

import (
	"errors"
	"fmt"
	"path/filepath"

	"cosmossdk.io/log/v2"
	db "github.com/cometbft/cometbft-db"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/iavl"
	iavldb "github.com/cosmos/iavl/db"
)

type Importer struct {
	branchCount   uint32
	leafCount     uint32
	stack         []*NodePointer
	stagedVersion uint32

	writer *ChangesetWriter
}

func NewImporter(stagedVersion uint32, treeDir string, log log.Logger) (*Importer, error) {
	ts, err := NewTreeStore(treeDir, TreeOptions{}, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create tree store: %w", err)
	}

	if ts.LatestVersion() != 0 {
		return nil, fmt.Errorf("cannot import into non-empty tree store, latest version is %d", ts.LatestVersion())
	}

	cw, err := NewChangesetWriter(treeDir, stagedVersion, ts)
	if err != nil {
		return nil, fmt.Errorf("failed to create changeset writer: %w", err)
	}

	return &Importer{
		stagedVersion: stagedVersion,
		writer:        cw,
	}, nil
}

type ExportNode = struct {
	Key     []byte
	Value   []byte
	Version int64
	Height  int8
}

func (i *Importer) Add(exportNode *ExportNode) error {
	if exportNode == nil {
		return errors.New("node cannot be nil")
	}
	version := uint32(exportNode.Version)
	if version > i.stagedVersion {
		return fmt.Errorf("node version %v can't be greater than import version %v",
			exportNode.Version, i.stagedVersion)
	}

	node := &MemNode{
		key:     exportNode.Key,
		value:   exportNode.Value,
		height:  uint8(exportNode.Height),
		version: version,
	}

	// We build the tree from the bottom-left up. The stack is used to store unresolved left
	// children while constructing right children. When all children are built, the parent can
	// be constructed and the resolved children can be discarded from the stack. Using a stack
	// ensures that we can handle additional unresolved left children while building a right branch.
	//
	// We don't modify the stack until we've verified the built node, to avoid leaving the
	// importer in an inconsistent state when we return an error.
	np := NewNodePointer(node)
	stackSize := len(i.stack)
	height := node.height
	if height == 0 {
		i.leafCount++
		node.nodeId = NewNodeID(true, 1, i.leafCount)

		node.size = 1
	} else if stackSize >= 2 {
		i.branchCount++
		node.nodeId = NewNodeID(false, 1, i.branchCount)

		leftPtr := i.stack[stackSize-1]
		rightPtr := i.stack[stackSize-2]
		left := leftPtr.Mem.Load()
		right := rightPtr.Mem.Load()
		if left == nil || right == nil {
			return fmt.Errorf("child node is nil for branch node at height %d", height)
		}
		if left.height != height-1 || right.height != height-1 {
			return fmt.Errorf("invalid child stack for branch node at height %d: left child height %d, right child height %d", height, left.height, right.height)
		}

		node.size = left.size + right.size
		node.left = leftPtr
		node.right = rightPtr

		// write nodes
		if err := i.writeNode(leftPtr); err != nil {
			return err
		}
		if err := i.writeNode(rightPtr); err != nil {
			return err
		}

		// update stack to remove the two children
		i.stack = i.stack[:stackSize-2]
	}

	np.id = node.nodeId
	i.stack = append(i.stack, np)

	return nil
}

func (i *Importer) Finalize() error {
	var cpInfo CheckpointInfo
	cpInfo.Version = i.stagedVersion
	cpInfo.Checkpoint = 1

	switch len(i.stack) {
	case 0:
		cpInfo.RootID = NewEmptyTreeNodeID(1)
	case 1:
		rootPtr := i.stack[0]
		err := i.writeNode(rootPtr)
		if err != nil {
			return err
		}
		cpInfo.RootID = rootPtr.id
	default:
		return fmt.Errorf("invalid node structure, found stack size %v when committing",
			len(i.stack))
	}

	totalBranches := i.branchCount
	if totalBranches > 0 {
		cpInfo.Branches.StartIndex = 1
		cpInfo.Branches.Count = totalBranches
		cpInfo.Branches.EndIndex = totalBranches
	}
	totalLeaves := i.leafCount
	if totalLeaves > 0 {
		cpInfo.Leaves.StartIndex = 1
		cpInfo.Leaves.Count = totalLeaves
		cpInfo.Leaves.EndIndex = totalLeaves
	}

	// file integrity check data
	cpInfo.KVEndOffset = uint64(i.writer.kvWriter.Size())
	cpInfo.SetCRC32()

	// commit checkpoint info
	err := i.writer.checkpointsData.Append(&cpInfo)
	if err != nil {
		return fmt.Errorf("failed to write checkpoint info: %w", err)
	}

	err = i.writer.Seal()
	if err != nil {
		return fmt.Errorf("failed to seal changeset: %w", err)
	}

	// sync all files
	files := i.writer.Changeset().Files()
	err = errors.Join(
		files.LeavesFile().Sync(),
		files.BranchesFile().Sync(),
		files.KVDataFile().Sync(),
		files.CheckpointsFile().Sync(),
	)
	if err != nil {
		return fmt.Errorf("failed to sync changeset files: %w", err)
	}

	err = i.writer.Changeset().Close()
	if err != nil {
		return fmt.Errorf("failed to close changeset: %w", err)
	}

	return nil
}

func (i *Importer) writeNode(np *NodePointer) error {
	node := np.Mem.Load()
	if node == nil {
		return fmt.Errorf("node is nil when writing to disk")
	}

	// compute hash
	_, err := node.ComputeHash(SyncHashScheduler{})
	if err != nil {
		return fmt.Errorf("failed to compute left child hash: %w", err)
	}

	// write node to disk
	if err := i.writer.writeNode(np); err != nil {
		return err
	}

	// remove the recursive references to avoid memory leak
	node.left.Mem.Store(nil)
	node.right.Mem.Store(nil)

	return nil
}

func (i *Importer) importExporter(exporter *iavl.Exporter) error {
	for {
		exportNode, err := exporter.Next()
		if err != nil {
			if errors.Is(err, iavl.ErrorExportDone) {
				break
			}
			return fmt.Errorf("failed to get next exported node: %w", err)
		}
		err = i.Add((*ExportNode)(exportNode))
		if err != nil {
			return fmt.Errorf("failed to add exported node: %w", err)
		}
	}
	return nil
}

func ImportIAVLV1Store(v1Db db.DB, multiStoreDir string, log log.Logger) error {
	panic("TODO")
}

func importIAVLV1Tree(v1Db dbm.DB, store, multiStoreDir string, log log.Logger) error {
	treeDir := filepath.Join(multiStoreDir, "stores", fmt.Sprintf("%s.iavl", store))

	v1Prefix := "s/k:" + store + "/"
	v1Db = dbm.NewPrefixDB(v1Db, []byte(v1Prefix))
	tree := iavl.NewMutableTree(iavldb.NewWrapper(v1Db), 0, false, log)
	_, err := tree.Load()
	if err != nil {
		return fmt.Errorf("failed to load IAVL tree: %w", err)
	}

	version, err := tree.GetLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to get latest version of IAVL tree: %w", err)
	}

	imTree, err := tree.GetImmutable(version)
	if err != nil {
		return fmt.Errorf("failed to get immutable tree for version %d: %w", version, err)
	}

	exporter, err := imTree.Export()
	if err != nil {
		return fmt.Errorf("failed to create exporter for version %d: %w", version, err)
	}

	importer, err := NewImporter(uint32(version), treeDir, log)
	if err != nil {
		return fmt.Errorf("failed to create importer for version %d: %w", version, err)
	}

	err = importer.importExporter(exporter)
	if err != nil {
		return fmt.Errorf("failed to import exported nodes for version %d: %w", version, err)
	}

	return importer.Finalize()
}
