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

	tx          Tx
	outPriority map[string]bool
	outNonce    map[string]bool
	inPriority  map[string]bool
	inNonce     map[string]bool

	pElement *huandu.Element
	nElement *huandu.Element
}

type graph struct {
	priorities  *huandu.SkipList
	nodes       map[string]*node
	senderNodes map[string]*huandu.SkipList
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

func (g *graph) AddEdge(from node, to node) {
	// TODO transition in* to a count? only used in finding the start node
	// or some other method for finding the top most node
	if from.sender == to.sender {
		from.outNonce[to.key()] = true
		to.inNonce[from.key()] = true
	} else {
		from.outPriority[to.key()] = true
		to.inPriority[from.key()] = true
	}
}

type nodePriorityKey struct {
	priority int64
	sender   string
	nonce    uint64
}

func nodePriorityKeyLess(a, b interface{}) int {
	keyA := a.(nodePriorityKey)
	keyB := b.(nodePriorityKey)
	res := huandu.Int64.Compare(keyA.priority, keyB.priority)
	if res != 0 {
		return res
	}

	res = huandu.Uint64.Compare(keyA.nonce, keyB.nonce)
	if res != 0 {
		return res
	}

	return huandu.String.Compare(keyA.sender, keyB.sender)
}

func NewGraph() *graph {
	return &graph{
		nodes:       make(map[string]*node),
		priorities:  huandu.New(huandu.GreaterThanFunc(nodePriorityKeyLess)),
		senderNodes: make(map[string]*huandu.SkipList),
	}
}

func (g *graph) AddNode(n *node) {
	g.nodes[n.key()] = n

	pnode := g.priorities.Set(nodePriorityKey{priority: n.priority, sender: n.sender, nonce: n.nonce}, n)
	sgs, ok := g.senderNodes[n.sender]
	if !ok {
		sgs = huandu.New(huandu.Uint64)
		g.senderNodes[n.sender] = sgs
	}

	nnode := sgs.Set(n.nonce, n)
	n.pElement = pnode
	n.nElement = nnode
}

func (g *graph) ContainsNode(n node) bool {
	_, ok := g.nodes[n.key()]
	return ok
}

func (g *graph) TopologicalSort() ([]*node, error) {
	maxPriority := g.priorities.Back().Value.(*node)
	start := g.senderNodes[maxPriority.sender].Front().Value.(*node)
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

	// priority edge rules:
	// - a node has an incoming priority edge if the next priority node in ascending order has p > this.p AND a different sender (in another tree)
	//    OR
	// - a node has an incoming priority edge if n.priority < latest L_n.priority.
	// - a node has an outgoing priority edge if the next priority node in descending order has p < this.p AND a different sender (in another tree)

	priorityCursor := edgeless[0].pElement
	for i := 0; i < len(edgeless) && edgeless[i] != nil; i++ {
		n := edgeless[i]
		//nextPriority = n.pElement.Next().Value.(*node).priority
		sorted = append(sorted, n)
		if n.priority == priorityCursor.Value.(*node).priority {
			priorityCursor = n.pElement
		}

		// nonce edge
		nextNonceNode := n.nElement.Next()
		if nextNonceNode != nil {
			m := nextNonceNode.Value.(*node)
			nonceEdge := nodeEdge(n, m)
			visited[nonceEdge] = true
			if !hasIncomingEdges(m, priorityCursor, visited) {
				edgeless = append(edgeless, m)
			}
		}

		// priority edge
		nextPriorityNode := n.pElement.Prev()
		if nextPriorityNode != nil {
			m := nextPriorityNode.Value.(*node)
			if m.sender != n.sender &&
				// no edge where priority is equal
				m.priority < n.priority {
				fmt.Println(nodeEdge(n, m))
				visited[nodeEdge(n, m)] = true
				if !hasIncomingEdges(m, priorityCursor, visited) {
					edgeless = append(edgeless, m)
				}
			}
		}
	}

	return sorted, nil
}

func hasIncomingEdges(n *node, pcursor *huandu.Element, visited map[string]bool) bool {
	prevNonceNode := n.nElement.Prev()
	if prevNonceNode != nil {
		m := prevNonceNode.Value.(*node)
		// if edge has not been visited return true
		incoming := !visited[nodeEdge(m, n)]
		if incoming {
			return true
		}
	}

	// priority edge
	if pcursor != nil {
		m := pcursor.Prev().Value.(*node)
		if m.sender != n.sender &&
			// no edge where priority is equal
			m.priority > n.priority {
			incoming := !visited[nodeEdge(m, n)]
			return incoming
		}
	}

	return false
}
