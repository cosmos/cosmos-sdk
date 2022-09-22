package mempool

import (
	"testing"
)

func initGraph() *graph {
	return NewGraph()
}

func initNodes(ns []*node) []node {
	var nodes []node
	// TODO what this API look like?
	for _, n := range ns {
		n.inNonce = make(map[string]bool)
		n.inPriority = make(map[string]bool)
		n.outNonce = make(map[string]bool)
		n.outPriority = make(map[string]bool)
		nodes = append(nodes, *n)
	}

	return nodes
}

func TestPoolCase(t *testing.T) {
	ns := []*node{
		{priority: 21, nonce: 4, sender: "a"}, // tx0
		{priority: 6, nonce: 3, sender: "a"},  // tx1
		{priority: 8, nonce: 2, sender: "a"},  // tx2
		{priority: 15, nonce: 1, sender: "b"}, // tx3
		{priority: 20, nonce: 1, sender: "a"}, // tx4
		{priority: 7, nonce: 2, sender: "b"},  // tx5
	}

	nodes := initNodes(ns)
	tests := []struct {
		name     string
		limit    int
		edges    [][]int
		expected []int
	}{
		{"case 1", 5,
			[][]int{{4, 2}, {4, 3}, {3, 2}, {2, 1}, {1, 0}},
			[]int{4, 3, 2, 1, 0},
		}, {
			"case 2", 6,
			[][]int{{4, 2}, {4, 3}, {3, 2}, {2, 1}, {1, 0}, {4, 5}, {2, 5}, {5, 1}},
			[]int{4, 3, 2, 5, 1, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := initGraph()
			for i := 0; i < tt.limit; i++ {
				//graph.AddNode(nodes[i])
			}
			for _, e := range tt.edges {
				graph.AddEdge(nodes[e[0]], nodes[e[1]])
			}

			results, err := graph.TopologicalSort()

			if err != nil {
				t.Error(err)
				return
			}
			if len(results) != len(tt.expected) {
				t.Errorf("Wrong number of results: %v", results)
				return
			}

			for i := 0; i < len(tt.expected); i++ {
				if results[i].key() != nodes[tt.expected[i]].key() {
					t.Errorf("Wrong sort order: %v", results)
					break
				}
			}
		})
	}
}
