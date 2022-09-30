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

func (g *senderGraph) canDrawEdge(priority int64, nonce uint64) bool {
	// optimization: if n is _sufficiently_ small we can just iterate the nonce list
	// otherwise we use the skip list by priority
	//
	min := g.byPriority.Front()
	for min != nil {
		n := min.Value.(*node)
		if n.priority > priority {
			return true
		}
		if n.priority < priority && n.nonce < nonce {
			break
		}
	}
}

type graph struct {
	priorities   *huandu.SkipList
	nodes        map[string]*node
	senderGraphs map[string]*huandu.SkipList
}

func (g *graph) Insert(context sdk.Context, tx Tx) error {
	senders := tx.(signing.SigVerifiableTx).GetSigners()
	nonces, err := tx.(signing.SigVerifiableTx).GetSignaturesV2()

	if err != nil {
		return err
	} else if len(senders) != len(nonces) {
		return fmt.Errorf("number of senders (%d) does not match number of nonces (%d)", len(senders), len(nonces))
	}

	// TODO multiple senders
	sender := senders[0].String()
	nonce := nonces[0].Sequence
	node := &node{priority: context.Priority(), nonce: nonce, sender: sender, tx: tx}
	g.AddNode(node)
	return nil
}

func (g *graph) AddNode(n *node) {
	g.nodes[n.key()] = n

	n.priorityNode = g.priorities.Set(txKey{priority: n.priority, sender: n.sender, nonce: n.nonce}, n)
	sgs, ok := g.senderGraphs[n.sender]
	if !ok {
		sgs = huandu.New(huandu.Uint64)
		g.senderGraphs[n.sender] = sgs
	}

	n.nonceNode = sgs.Set(n.nonce, n)
}

func (g *graph) Select(ctx sdk.Context, txs [][]byte, maxBytes int) ([]Tx, error) {
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
	//TODO implement me
	panic("implement me")
}

func (g *graph) Remove(context sdk.Context, tx Tx) error {
	//TODO implement me
	panic("implement me")
}

func (n node) key() string {
	return fmt.Sprintf("%d-%s-%d", n.priority, n.sender, n.nonce)
}

func (n node) String() string {
	return n.key()
}

func NewGraph() *graph {
	return &graph{
		nodes:        make(map[string]*node),
		priorities:   huandu.New(huandu.LessThanFunc(txKeyLess)),
		senderGraphs: make(map[string]*huandu.SkipList),
	}
}

func (g *graph) ContainsNode(n node) bool {
	_, ok := g.nodes[n.key()]
	return ok
}

func (g *graph) TopologicalSort() ([]*node, error) {
	maxPriority := g.priorities.Front().Value.(*node)
	start := g.senderGraphs[maxPriority.sender].Front().Value.(*node)
	edgeless := []*node{start}
	sorted, err := g.kahns(edgeless)
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

// DrawPriorityEdges is O(n^2) hopefully n is not too large
// Given n_a, need an answer the question:
// "Is there node n_b in my sender tree with n_b.nonce < n_a.nonce AND n_b.priority < n_a.priority?"
// If yes, don't draw any priority edges to nodes (or possibly just to nodes with a priority < n_b.priority)
//
func (g *graph) DrawPriorityEdges() (in map[string]map[string]bool, out map[string]map[string]bool) {
	pn := g.priorities.Front()
	in = make(map[string]map[string]bool)
	out = make(map[string]map[string]bool)

	for pn != nil {
		n := pn.Value.(*node)
		nk := n.key()
		out[nk] = make(map[string]bool)
		if n.nonceNode.Next() != nil {
			m := n.nonceNode.Next().Value.(*node)
			mk := m.key()
			if _, ok := in[mk]; !ok {
				in[mk] = make(map[string]bool)
			}

			out[nk][mk] = true
			in[mk][nk] = true
		}

		pm := pn.Next()
		for pm != nil {
			m := pm.Value.(*node)
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

func (g *graph) kahns(edgeless []*node) ([]*node, error) {
	var sorted []*node
	inEdges, outEdges := g.DrawPriorityEdges()

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
			delete(outEdges[nk], mk)
			delete(inEdges[mk], nk)

			if len(inEdges[mk]) == 0 {
				edgeless = append(edgeless, g.nodes[mk])
			}
		}

	}

	return sorted, nil
}
