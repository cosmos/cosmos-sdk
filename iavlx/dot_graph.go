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

var graphvizColors = []string{"purple", "green", "red", "blue", "yellow"}

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

		label := fmt.Sprintf("ver: %d idx: %d key:0x%x ", version, id.Index(), key)
		attrs := ""
		color := graphvizColors[version%uint32(len(graphvizColors))]
		attrs += fmt.Sprintf(" color=%s", color)
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

		nodeName := fmt.Sprintf("n%p", node)

		_, err = fmt.Fprintf(writer, "\t%s [label=\"%s\"%s];\n", nodeName, label, attrs)
		if err != nil {
			return err
		}
		if parent != nil {
			parentName := fmt.Sprintf("n%p", parent)
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
	curVersion := uint64(0)
	var lastBranchId NodeID
	for i := 0; i < numBranches; i++ {
		branchLayout := cs.branchesData.UnsafeItem(uint32(i))
		id := branchLayout.ID()
		nodeVersion := id.Version()
		if nodeVersion != curVersion {
			if curVersion != 0 {
				_, err = fmt.Fprintln(writer, "\t}")
			}
			color := graphvizColors[nodeVersion%uint64(len(graphvizColors))]
			_, err = fmt.Fprintf(writer, "\tsubgraph cluster_B%d {\n\t\tlabel=\"Version %d\" color=%s style=filled\n", nodeVersion, nodeVersion, color)
		}
		curVersion = nodeVersion
		if lastBranchId != 0 {
			_, err = fmt.Fprintf(writer, "\t\tN%d -> N%d [style=invis];\n", lastBranchId, id)
		}
		lastBranchId = id

		nodeName := fmt.Sprintf("N%d", id)
		label := fmt.Sprintf("idx: %d", id.Index())
		orphanVersion, isOrphan := orphans[id]
		if isOrphan {
			label += fmt.Sprintf("<BR/>orphaned: <B>%d</B>", orphanVersion)
		}
		attrs := ""
		if isOrphan {
			attrs = " style=dashed"
		}
		_, err = fmt.Fprintf(writer, "\t\t%s [label=<%s>%s];\n", nodeName, label, attrs)
		if err != nil {
			return err
		}

		//leftNodeName := fmt.Sprintf("N%d", branchLayout.Left)
		//rightNodeName := fmt.Sprintf("N%d", branchLayout.Right)
		//_, err = fmt.Fprintf(writer, "\t\t%s -> %s [constraint=false style=dashed label=\"L\"];\n", nodeName, leftNodeName)
		//if err != nil {
		//	return err
		//}
		//_, err = fmt.Fprintf(writer, "\t\t%s -> %s [constraint=false style=dashed label=\"R\"];\n", nodeName, rightNodeName)
		//if err != nil {
		//	return err
		//}
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
			color := graphvizColors[nodeVersion%uint64(len(graphvizColors))]
			_, err = fmt.Fprintf(writer, "\tsubgraph cluster_L%d {\n\t\tlabel=\"Version %d\" color=%s style=filled\n", nodeVersion, nodeVersion, color)
		}
		curVersion = nodeVersion
		if lastLeafId != 0 {
			_, err = fmt.Fprintf(writer, "\t\tN%d -> N%d [style=invis];\n", lastLeafId, id)
		}
		lastLeafId = id

		nodeName := fmt.Sprintf("N%d", id)
		label := fmt.Sprintf("idx: %d", id.Index())
		orphanVersion, isOrphan := orphans[id]
		if isOrphan {
			label += fmt.Sprintf("<BR/>orphaned: <B>%d</B>", orphanVersion)
		}
		attrs := ""
		if isOrphan {
			attrs = " style=dashed"
		}
		_, err = fmt.Fprintf(writer, "\t\t%s [label=<%s> shape=box%s];\n", nodeName, label, attrs)
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
