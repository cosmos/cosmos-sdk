package mempool

import (
	"fmt"

	huandu "github.com/huandu/skiplist"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var _ Mempool = (*graph)(nil)

type node struct {
	priority int64
	nonce    uint64
	sender   string
	tx       Tx

	nonceNode    *huandu.Element
	priorityNode *huandu.Element
	in           map[string]bool
}

type senderGraph struct {
	byNonce    *huandu.SkipList
	byPriority *huandu.SkipList
}

// MaxPriorityEdge returns the maximum out of tree priority which the node with
// the given priority and nonce can draw edges to by finding the minimum priority
// in this tree with a lower nonce
func (sg senderGraph) MaxPriorityEdge(priority int64, nonce uint64) int64 {
	// optimization: if n is _sufficiently_ small we can just iterate the nonce list
	// otherwise we use the skip list by priority
	//
	min := sg.byPriority.Front()
	for min != nil {
		n := min.Value.(*node)
		if n.priority < priority && n.nonce < nonce {
			// the minimum priority in the tree with a lower nonce
			return n.priority
		}
		min = min.Next()
	}
	// otherwise we can draw to anything
	return priority - 1
}

func newSenderGraph() senderGraph {
	return senderGraph{
		byNonce:    huandu.New(huandu.Uint64),
		byPriority: huandu.New(huandu.Int64),
	}
}

type graph struct {
	priorities   *huandu.SkipList
	nodes        map[string]*node
	senderGraphs map[string]senderGraph
	iterations   int
}

func (g *graph) Insert(context sdk.Context, tx Tx) error {
	sigs, err := tx.(signing.SigVerifiableTx).GetSignaturesV2()
	if err != nil {
		return err
	}

	sig := sigs[0]
	sender := sig.PubKey.Address().String()
	nonce := sig.Sequence
	node := &node{priority: context.Priority(), nonce: nonce, sender: sender, tx: tx}
	g.AddNode(node)
	return nil
}

func (g *graph) AddNode(n *node) {
	g.nodes[n.key()] = n

	n.priorityNode = g.priorities.Set(txKey{priority: n.priority, sender: n.sender, nonce: n.nonce}, n)
	sg, ok := g.senderGraphs[n.sender]
	if !ok {
		sg = newSenderGraph()
		g.senderGraphs[n.sender] = sg
	}

	n.nonceNode = sg.byNonce.Set(n.nonce, n)
	sg.byPriority.Set(n.priority, n)
}

func (g *graph) Select(txs [][]byte, maxBytes int64) ([]Tx, error) {
	// todo collapse multiple iterations into kahns
	sorted, err := g.TopologicalSort()
	if err != nil {
		return nil, err
	}
	var res []Tx
	for _, n := range sorted {
		res = append(res, n.tx)
	}
	return res, nil
}

func (g *graph) CountTx() int {
	return len(g.nodes)
}

func (g *graph) Remove(tx Tx) error {
	//TODO implement me
	panic("implement me")
}

func (n node) key() string {
	return fmt.Sprintf("%d-%s-%d", n.priority, n.sender, n.nonce)
}

func (n node) String() string {
	return n.key()
}

func NewGraph() Mempool {
	return &graph{
		nodes:        make(map[string]*node),
		priorities:   huandu.New(huandu.LessThanFunc(txKeyLess)),
		senderGraphs: make(map[string]senderGraph),
	}
}

func (g *graph) ContainsNode(n node) bool {
	_, ok := g.nodes[n.key()]
	return ok
}

func (g *graph) TopologicalSort() ([]*node, error) {
	in, out := g.DrawPriorityEdges()
	var edgeless []*node
	for _, n := range g.nodes {
		g.iterations++
		nk := n.key()
		if _, ok := in[nk]; !ok || len(in[nk]) == 0 {
			edgeless = append(edgeless, n)
		}
	}
	sorted, err := g.kahns(edgeless, in, out)
	if err != nil {
		return nil, err
	}
	return sorted, nil
}

func nodeEdge(n *node, m *node) string {
	return fmt.Sprintf("%s->%s", n.key(), m.key())
}

/*

DFS
L ← Empty list that will contain the sorted nodes
while exists nodes without a permanent mark do
    select an unmarked node n
    visit(n)

function visit(node n)
    if n has a permanent mark then
        return
    if n has a temporary mark then
        stop   (not a DAG)

    mark n with a temporary mark

    for each node m with an edge from n to m do
        visit(m)

    remove temporary mark from n
    mark n with a permanent mark
    add n to head of L

*/

type nodeEdges map[string]map[string]bool

// DrawPriorityEdges is O(n^2) hopefully n is not too large
// Given n_a, need an answer the question:
// "Is there node n_b in my sender tree with n_b.nonce < n_a.nonce AND n_b.priority < n_a.priority?"
// If yes, don't draw any priority edges to nodes (or possibly just to nodes with a priority < n_b.priority)
func (g *graph) DrawPriorityEdges() (in nodeEdges, out nodeEdges) {
	pn := g.priorities.Front()
	in = make(nodeEdges)
	out = make(nodeEdges)

	for pn != nil {
		g.iterations++
		n := pn.Value.(*node)
		nk := n.key()
		out[nk] = make(map[string]bool)

		// draw nonce edge (if there is one)
		if n.nonceNode.Next() != nil {
			m := n.nonceNode.Next().Value.(*node)
			mk := m.key()
			if _, ok := in[mk]; !ok {
				in[mk] = make(map[string]bool)
			}

			out[nk][mk] = true
			in[mk][nk] = true
		}

		// beginning with the next lowest priority node, draw priority edges
		maxp := g.senderGraphs[n.sender].MaxPriorityEdge(n.priority, n.nonce)
		pm := pn.Next()
		for pm != nil {
			g.iterations++
			m := pm.Value.(*node)
			// skip these nodes
			if m.priority > maxp {
				pm = pm.Next()
				continue
			}
			mk := m.key()
			if _, ok := in[mk]; !ok {
				in[mk] = make(map[string]bool)
			}

			if n.sender != m.sender {
				out[nk][mk] = true
				in[mk][nk] = true
			}

			pm = pm.Next()
		}
		pn = pn.Next()
	}

	return in, out
}

func (g *graph) kahns(edgeless []*node, inEdges nodeEdges, outEdges nodeEdges) ([]*node, error) {
	var sorted []*node

	/*
		L ← Empty list that will contain the sorted elements
		S ← Set of all nodes with no incoming edge

		while S is not empty do
		    remove a node n from S
		    add n to L
		    for each node m with an edge e from n to m do
		        remove edge e from the graph
		        if m has no other incoming edges then
		            insert m into S

		if graph has edges then
		    return error   (graph has at least one cycle)
		else
		    return L   (a topologically sorted order)
	*/

	for i := 0; i < len(edgeless) && edgeless[i] != nil; i++ {
		n := edgeless[i]
		nk := n.key()
		sorted = append(sorted, n)

		for mk, _ := range outEdges[nk] {
			g.iterations++

			delete(outEdges[nk], mk)
			delete(inEdges[mk], nk)

			if len(inEdges[mk]) == 0 {
				edgeless = append(edgeless, g.nodes[mk])
			}
		}

	}

	return sorted, nil
}
