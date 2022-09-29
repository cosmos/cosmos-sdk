package mempool

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	huandu "github.com/huandu/skiplist"
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

func (g *graph) DrawPriorityEdges(n *node) (edges []*node) {
	pnode := n.nonceNode.Prev().Value.(*node).priorityNode
	for pnode != nil {
		node := pnode.Value.(*node)
		if node.sender != n.sender {
			edges = append(edges, node)
		}

		pnode = pnode.Prev()
	}
	return edges
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
		priorities:   huandu.New(huandu.GreaterThanFunc(txKeyLess)),
		senderGraphs: make(map[string]*huandu.SkipList),
	}
}

func (g *graph) ContainsNode(n node) bool {
	_, ok := g.nodes[n.key()]
	return ok
}

func (g *graph) TopologicalSort() ([]*node, error) {
	maxPriority := g.priorities.Back().Value.(*node)
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
func (g *graph) kahns(edgeless []*node) ([]*node, error) {
	var sorted []*node
	visited := make(map[string]bool)

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
	var n *node
	//n := edgeless[0]
	for i := 0; i < len(edgeless) && edgeless[i] != nil; i++ {
		// enumerate the priority list drawing priority edges along the way
		pnodes := g.DrawPriorityEdges(n.priorityNode.Value.(*node))
		for _, m := range pnodes {
			edge := nodeEdge(n, m)
			if !visited[edge] {
				visited[edge] = true
			}
		}
	}

	return sorted, nil
}
