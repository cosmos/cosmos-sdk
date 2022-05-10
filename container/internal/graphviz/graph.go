package graphviz

import "io"

type Graph struct {
	parent    *Graph
	nodes     map[string]*Node
	subgraphs map[string]*Graph
}

func NewGraph() *Graph {
	return &Graph{
		nodes: map[string]*Node{},
	}
}

func (g *Graph) FindOrCreateNode(name string) (node *Node, found bool) {
	if node, ok := g.nodes[name]; ok {
		return node, true
	}

	node = &Node{name: name}
	g.nodes[name] = node
	return node, false
}

func (n *Graph) SubGraph(name string) *Graph {
	return &Graph{}
}

func (g *Graph) SetLabel(label string) {

}

func (g *Graph) CreateEdge(name string, from, to *Node) *Edge {
	return &Edge{}
}

func (g *Graph) RenderDOT(writer io.Writer) {

}
