package internal

import (
	"errors"
	"fmt"

	"github.com/cosmos/iavl"

	"cosmossdk.io/log/v2"
)

// Importer builds an iavlx tree from a stream of ExportNodes in post-order (left subtree,
// right subtree, parent). This is the same order produced by TreeReader.Export and by iavl/v1's
// Exporter, making it compatible with both iavlx-to-iavlx and v1-to-iavlx migration.
//
// The import algorithm uses a stack:
//   - Leaf nodes are pushed directly onto the stack.
//   - Branch nodes pop their two children (left = stack[-2], right = stack[-1]), wire them up,
//     write them to disk, and push the branch onto the stack.
//   - When Finalize is called, exactly one node should remain on the stack — the root.
//
// The result is a single changeset directory containing one checkpoint with the full tree.
// After import, the tree can be opened normally via NewTreeStore/load().
//
// Used by the `iavlx import` CLI for one-time offline migration from iavl/v1.
type Importer struct {
	logger log.Logger

	branchCount   uint32
	leafCount     uint32
	stack         []*NodePointer
	stagedVersion uint32

	writer *ChangesetWriter
}

// NewImporter creates a new Importer that will write into treeDir. stagedVersion is the
// version of the tree being imported (e.g. the latest version from the source iavl/v1 tree);
// importing nodes with a version greater than stagedVersion is an error. The tree directory
// must be empty (no prior data).
func NewImporter(stagedVersion uint32, treeDir string, logger log.Logger) (*Importer, error) {
	ts, err := NewTreeStore(treeDir, TreeOptions{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create tree store: %w", err)
	}

	if ts.LatestVersion() != 0 {
		return nil, fmt.Errorf("cannot import into non-empty tree store, latest version is %d", ts.LatestVersion())
	}

	// we set staged version of the writer to 1 so that this is the changeset 1 which contains all historical nodes
	cw, err := NewChangesetWriter(treeDir, 1, ts)
	if err != nil {
		return nil, fmt.Errorf("failed to create changeset writer: %w", err)
	}

	// we must force the writer's checkpoint to 1 or else writing nodes will fail
	cw.checkpoint = 1

	return &Importer{
		logger:        logger,
		stagedVersion: stagedVersion,
		writer:        cw,
	}, nil
}

// ExportNode is a single node in the post-order export stream produced by TreeReader.Export
// and consumed by Importer.Add. Leaf nodes have Height == 0 and a non-nil Value; branch
// nodes have Height > 0 and a nil Value.
type ExportNode = struct {
	Key     []byte
	Value   []byte
	Version int64
	Height  int8
}

// Add processes one ExportNode in post-order sequence. Leaf nodes are pushed onto the
// stack; branch nodes pop their two children, wire them up, write them to disk, and push
// the assembled branch. Nodes must arrive in the same order produced by TreeReader.Export.
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

		// In post-order export (left subtree, right subtree, parent), the left child
		// is pushed first (stack[stackSize-2]) and the right child last (stack[stackSize-1]).
		leftPtr := i.stack[stackSize-2]
		rightPtr := i.stack[stackSize-1]
		left := leftPtr.Mem.Load()
		right := rightPtr.Mem.Load()
		if left == nil || right == nil {
			return fmt.Errorf("child node is nil for branch node at height %d", height)
		}
		// IAVL height = max(left.height, right.height) + 1, so the taller child must be
		// at exactly height-1. The shorter child may be at height-1 or height-2 (AVL balance).
		maxChildHeight := left.height
		if right.height > maxChildHeight {
			maxChildHeight = right.height
		}
		if maxChildHeight != height-1 {
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

// Finalize validates that exactly one root node remains on the stack, writes the final
// checkpoint, seals and syncs all changeset files. After Finalize, the tree can be opened
// normally via NewTreeStore/load().
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

	i.logger.Info("Successfully finalized import", "version", cpInfo.Version, "checkpoint", cpInfo.Checkpoint, "rootID", cpInfo.RootID, "branchCount", cpInfo.Branches.Count, "leafCount", cpInfo.Leaves.Count)

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
	if node.left != nil {
		node.left.Mem.Store(nil)
	}
	if node.right != nil {
		node.right.Mem.Store(nil)
	}

	return nil
}

// importExporter drains an iavl/v1 Exporter, feeding each exported node into Add until
// the exporter signals completion via ErrorExportDone.
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
