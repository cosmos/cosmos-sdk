package iavlx

import (
	"fmt"
	"io"
)

func DebugTraverseNode(nodePtr *NodePointer, onNode func(node, parent Node, direction string) error) error {
	if nodePtr == nil {
		return nil
	}

	var traverse func(np *NodePointer, parent Node, direction string) error
	traverse = func(np *NodePointer, parent Node, direction string) error {
		node, err := np.Resolve()
		if err != nil {
			return err
		}

		if err := onNode(node, parent, direction); err != nil {
			return err
		}

		if node.IsLeaf() {
			return nil
		}

		err = traverse(node.Left(), node, "l")
		if err != nil {
			return err
		}
		err = traverse(node.Right(), node, "r")
		if err != nil {
			return err
		}
		return nil
	}

	return traverse(nodePtr, nil, "")
}

var graphvizFillColors = []string{"purple", "green", "red", "blue", "yellow"}
var graphvizTextColors = []string{"white", "black", "white", "white", "black"}

func RenderNodeDotGraph(writer io.Writer, nodePtr *NodePointer) error {
	_, err := fmt.Fprintln(writer, "digraph G {\n\trankdir=BT")
	if err != nil {
		return err
	}
	finishGraph := func() error {
		_, err := fmt.Fprintln(writer, "}")
		return err
	}
	if nodePtr == nil {
		return finishGraph()
	}

	err = DebugTraverseNode(nodePtr, func(node, parent Node, direction string) error {
		key, err := node.Key()
		if err != nil {
			return err
		}

		version := node.Version()
		id := node.ID()

		idx := id.Index()
		label := fmt.Sprintf("ver: %d idx: %d key:0x%x ", version, idx, key)
		attrs := ""
		fillColor := graphvizFillColors[version%uint32(len(graphvizFillColors))]
		textColor := graphvizTextColors[version%uint32(len(graphvizTextColors))]
		attrs += fmt.Sprintf(" fillcolor=%s fontcolor=%s style=filled", fillColor, textColor)
		if node.IsLeaf() {
			value, err := node.Value()
			if err != nil {
				return err
			}

			label += fmt.Sprintf("val:0x%X", value)
			attrs += " shape=box"
		} else {
			label += fmt.Sprintf("ht:%d sz:%d", node.Height(), node.Size())
		}

		nodeName := graphvizNodeID(id)

		_, err = fmt.Fprintf(writer, "\t%s [id=%s label=\"%s\"%s];\n", nodeName, nodeName, label, attrs)
		if err != nil {
			return err
		}
		if parent != nil {
			parentName := graphvizNodeID(parent.ID())
			_, err = fmt.Fprintf(writer, "\t%s -> %s [label=\"%s\"];\n", parentName, nodeName, direction)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return finishGraph()
}

func RenderChangesetDotGraph(writer io.Writer, cs *Changeset, orphans map[NodeID]uint32) error {
	_, err := fmt.Fprintln(writer, "digraph G {\n\trankdir=LR")
	if err != nil {
		return err
	}
	finishGraph := func() error {
		_, err := fmt.Fprintln(writer, "}")
		return err
	}

	_, err = fmt.Fprintln(writer, "\tsubgraph cluster_branches {\t\nlabel=\"Branches\"")
	if err != nil {
		return err
	}

	numBranches := cs.branchesData.Count()
	curVersion := uint32(0)
	var lastBranchId NodeID
	for i := 0; i < numBranches; i++ {
		branchLayout := cs.branchesData.UnsafeItem(uint32(i))
		id := branchLayout.ID()
		nodeVersion := id.Version()
		if nodeVersion != curVersion {
			if curVersion != 0 {
				_, err = fmt.Fprintln(writer, "\t}")
			}
			fillColor := graphvizFillColors[nodeVersion%uint32(len(graphvizFillColors))]
			textColor := graphvizTextColors[nodeVersion%uint32(len(graphvizTextColors))]
			_, err = fmt.Fprintf(writer, "\tsubgraph cluster_B%d {\n\t\tlabel=\"Version %d\" color=%s style=filled fontcolor=%s node [fontcolor=%s]\n", nodeVersion, nodeVersion, fillColor, textColor, textColor)
		}
		curVersion = nodeVersion
		if lastBranchId.IsEmpty() {
			_, err = fmt.Fprintf(writer, "\t\t%s -> %s [style=invis];\n", graphvizNodeID(lastBranchId), graphvizNodeID(id))
		}
		lastBranchId = id

		nodeName := graphvizNodeID(id)
		idx := id.Index()
		label := fmt.Sprintf("idx: %d", idx)
		orphanVersion, isOrphan := orphans[id]
		if isOrphan {
			label += fmt.Sprintf("<BR/>orphaned: <B>%d</B>", orphanVersion)
		}
		attrs := ""
		if isOrphan {
			attrs = " style=dashed"
		}
		vi, err := cs.getVersionInfo(uint32(nodeVersion))
		if err != nil {
			return err
		}
		if vi.RootID == id {
			attrs += " shape=doublecircle"
		}
		_, err = fmt.Fprintf(writer, "\t\t%s [id=%s label=<%s>%s];\n", nodeName, nodeName, label, attrs)
		if err != nil {
			return err
		}
	}
	// finish last version subgraph
	_, err = fmt.Fprintln(writer, "\t}")
	if err != nil {
		return err
	}

	// finish branches subgraph
	_, err = fmt.Fprintln(writer, "\t}")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(writer, "\tsubgraph cluster_leaves {\t\nlabel=\"Leaves\"")
	if err != nil {
		return err
	}
	numLeaves := cs.leavesData.Count()
	curVersion = 0
	var lastLeafId NodeID
	for i := 0; i < numLeaves; i++ {
		leafLayout := cs.leavesData.UnsafeItem(uint32(i))
		id := leafLayout.ID()
		nodeVersion := id.Version()
		if nodeVersion != curVersion {
			if curVersion != 0 {
				_, err = fmt.Fprintln(writer, "\t}")
			}
			fillColor := graphvizFillColors[nodeVersion%uint32(len(graphvizFillColors))]
			textColor := graphvizTextColors[nodeVersion%uint32(len(graphvizTextColors))]
			_, err = fmt.Fprintf(writer, "\tsubgraph cluster_L%d {\n\t\tlabel=\"Version %d\" color=%s fontcolor=%s style=filled node [fontcolor=%s]\n", nodeVersion, nodeVersion, fillColor, textColor, textColor)
		}
		curVersion = nodeVersion
		if lastLeafId.IsEmpty() {
			_, err = fmt.Fprintf(writer, "\t\t%s -> %s [style=invis];\n", graphvizNodeID(lastLeafId), graphvizNodeID(id))
		}
		lastLeafId = id

		nodeName := graphvizNodeID(id)
		label := fmt.Sprintf("idx: %d", id.Index())
		orphanVersion, isOrphan := orphans[id]
		if isOrphan {
			label += fmt.Sprintf("<BR/>orphaned: <B>%d</B>", orphanVersion)
		}
		attrs := ""
		if isOrphan {
			attrs = " style=dashed"
		}
		_, err = fmt.Fprintf(writer, "\t\t%s [id=%s label=<%s> shape=box%s];\n", nodeName, nodeName, label, attrs)
		if err != nil {
			return err
		}
	}
	// finish last version subgraph
	_, err = fmt.Fprintln(writer, "\t}")
	if err != nil {
		return err
	}

	// finish leaves subgraph
	_, err = fmt.Fprintln(writer, "\t}")
	if err != nil {
		return err
	}

	return finishGraph()
}

func graphvizNodeID(node NodeID) string {
	if node.IsLeaf() {
		return fmt.Sprintf("L%d_%d", node.Version(), node.Index())
	} else {
		return fmt.Sprintf("B%d_%d", node.Version(), node.Index())
	}
}
