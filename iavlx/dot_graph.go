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

	colors := []string{"red", "green", "blue", "orange", "purple"}
	err = DebugTraverseNode(nodePtr, func(node, parent Node, direction string) error {
		key, err := node.Key()
		if err != nil {
			return err
		}

		version := node.Version()
		id := node.ID()

		label := fmt.Sprintf("ver: %d idx: %d key:0x%x ", version, id.Index(), key)
		attrs := ""
		color := colors[version%uint32(len(colors))]
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

func RenderChangesetDotGraph(writer io.Writer, cs *Changeset) error {
	_, err := fmt.Fprintln(writer, "digraph G {")
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
	for i := 0; i < numBranches; i++ {
		branchLayout := cs.branchesData.UnsafeItem(uint32(i))
		id := branchLayout.ID()
		nodeVersion := id.Version()
		if nodeVersion != curVersion {
			if curVersion != 0 {
				_, err = fmt.Fprintln(writer, "\t}")
			}
			_, err = fmt.Fprintf(writer, "\tsubgraph cluster_B%d {\nlabel=\"Version %d\"\n", nodeVersion, nodeVersion)
		}
		curVersion = nodeVersion

		nodeName := fmt.Sprintf("N%d", id)
		label := fmt.Sprintf("idx: %d", id.Index())
		_, err = fmt.Fprintf(writer, "\t%s [label=\"%s\"];\n", nodeName, label)
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

	return finishGraph()
}
