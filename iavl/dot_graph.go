package iavlx

import (
	"fmt"
	"io"
)

func DebugTraverse(nodePtr *NodePointer, onNode func(node Node, parent Node, direction string) error) error {
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

func RenderDotGraph(writer io.Writer, nodePtr *NodePointer) error {
	_, err := fmt.Fprintln(writer, "digraph G {")
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

	err = DebugTraverse(nodePtr, func(node Node, parent Node, direction string) error {
		key, err := node.Key()
		if err != nil {
			return err
		}

		version := node.Version()

		label := fmt.Sprintf("ver: %d key:0x%x ", version, key)
		if node.IsLeaf() {
			value, err := node.Value()
			if err != nil {
				return err
			}

			label += fmt.Sprintf("val:0x%X", value)
		} else {
			label += fmt.Sprintf("ht:%d sz:%d", node.Height(), node.Size())
		}

		nodeName := fmt.Sprintf("n%p", node)

		_, err = fmt.Fprintf(writer, "%s [label=\"%s\"];\n", nodeName, label)
		if err != nil {
			return err
		}
		if parent != nil {
			parentName := fmt.Sprintf("n%p", parent)
			_, err = fmt.Fprintf(writer, "%s -> %s [label=\"%s\"];\n", parentName, nodeName, direction)
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
