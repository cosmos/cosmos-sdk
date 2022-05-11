// Package graphviz
package graphviz

import (
	"bytes"
	"fmt"
	"io"
)

// Graph represents a graphviz digraph.
type Graph struct {
	*Attributes

	// name is the optional name of this graph
	name string

	// parent is non-nil if this is a sub-graph
	parent *Graph

	// allNodes includes all nodes in the graph and its sub-graphs.
	// It is set to the same map in parent and sub-graphs.
	allNodes map[string]*Node

	// myNodes are the nodes in this graph (whether it's a root or sub-graph)
	myNodes map[string]*Node

	subgraphs map[string]*Graph

	edges []*Edge
}

// NewGraph creates a new Graph instance.
func NewGraph() *Graph {
	return &Graph{
		Attributes: NewAttributes(),
		name:       "",
		parent:     nil,
		allNodes:   map[string]*Node{},
		myNodes:    map[string]*Node{},
		subgraphs:  map[string]*Graph{},
		edges:      nil,
	}
}

// FindOrCreateNode finds or creates the node with the provided name.
func (g *Graph) FindOrCreateNode(name string) (node *Node, found bool) {
	if node, ok := g.allNodes[name]; ok {
		return node, true
	}

	node = &Node{
		Attributes: NewAttributes(),
		name:       name,
	}
	g.allNodes[name] = node
	g.myNodes[name] = node
	return node, false
}

// SubGraph finds or creates the subgraph with the provided name.
func (g *Graph) SubGraph(name string) *Graph {
	if sub, ok := g.subgraphs[name]; ok {
		return sub
	}

	n := &Graph{
		Attributes: NewAttributes(),
		name:       name,
		parent:     g,
		allNodes:   g.allNodes,
		myNodes:    map[string]*Node{},
		subgraphs:  map[string]*Graph{},
		edges:      nil,
	}
	g.subgraphs[name] = n
	return n
}

// CreateEdge creates a new graphviz edge.
func (g *Graph) CreateEdge(from, to *Node) *Edge {
	edge := &Edge{
		Attributes: NewAttributes(),
		from:       from,
		to:         to,
	}
	g.edges = append(g.edges, edge)
	return edge
}

// RenderDOT renders the graph to DOT format.
func (g *Graph) RenderDOT(w io.Writer) error {
	return g.render(w, "")
}

func (g *Graph) render(w io.Writer, indent string) error {
	if g.parent == nil {
		_, err := fmt.Fprintf(w, "%sdigraph %q {\n", indent, g.name)
		if err != nil {
			return err
		}
	} else {
		_, err := fmt.Fprintf(w, "%ssubgraph %q {\n", indent, g.name)
		if err != nil {
			return err
		}
	}

	{
		subIndent := indent + "  "

		if attrStr := g.Attributes.String(); attrStr != "" {
			_, err := fmt.Fprintf(w, "%sgraph %s;\n", subIndent, attrStr)
			if err != nil {
				return err
			}
		}

		for _, subgraph := range g.subgraphs {
			err := subgraph.render(w, subIndent+"  ")
			if err != nil {
				return err
			}
		}

		for _, node := range g.myNodes {
			err := node.render(w, subIndent)
			if err != nil {
				return err
			}
		}

		for _, edge := range g.edges {
			err := edge.render(w, subIndent)
			if err != nil {
				return err
			}
		}
	}

	_, err := fmt.Fprintf(w, "%s}\n\n", indent)
	return err
}

// String returns the graph in DOT format.
func (g *Graph) String() string {
	buf := &bytes.Buffer{}
	err := g.RenderDOT(buf)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
