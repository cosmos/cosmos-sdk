package mempool

import (
	"container/list"
	"fmt"

	huandu "github.com/huandu/skiplist"
)

type node struct {
	priority    int64
	nonce       uint64
	sender      string
	tx          Tx
	outPriority map[string]bool
	outNonce    map[string]bool
	inPriority  map[string]bool
	inNonce     map[string]bool
}

type graph struct {
	priorities *huandu.SkipList
	nodes      map[string]node
}

func (n node) key() string {
	return fmt.Sprintf("%s-%d-%d", n.sender, n.priority, n.nonce)
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

func NewGraph() *graph {
	return &graph{
		nodes:      make(map[string]node),
		priorities: huandu.New(huandu.Int64),
	}
}

func (g *graph) AddNode(n node) {
	if !g.ContainsNode(n) {
		g.nodes[n.key()] = n
	}
	pnode := g.priorities.Set(n.priority, n.key())
	if pnode.Prev() != nil {

	}
}

func (g *graph) ContainsNode(n node) bool {
	_, ok := g.nodes[n.key()]
	return ok
}

func (g *graph) TopologicalSort() ([]node, error) {
	sn := g.priorities.Back()
	var start node
	for sn != nil {
		start = g.nodes[sn.Value.(string)]
		if len(start.inPriority) == 0 && len(start.inNonce) == 0 {
			break
		}

		sn = sn.Prev()
	}
	sorted := list.New()
	err := g.visit(start, make(map[string]bool), make(map[string]bool), sorted)
	if err != nil {
		return nil, err
	}
	var res []node
	for e := sorted.Front(); e != nil; e = e.Next() {
		res = append(res, e.Value.(node))
	}

	return res, nil
}

/*
Kahn's Algorithm
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
func (g *graph) visit(n node, marked map[string]bool, tmp map[string]bool, sorted *list.List) error {
	if _, ok := marked[n.key()]; ok {
		return nil
	}
	if _, ok := tmp[n.key()]; ok {
		return fmt.Errorf("not a DAG, cycling on %s", n.key())
	}

	tmp[n.key()] = true

	for m := range n.outPriority {
		err := g.visit(g.nodes[m], marked, tmp, sorted)
		if err != nil {
			return err
		}
	}
	for m := range n.outNonce {
		err := g.visit(g.nodes[m], marked, tmp, sorted)
		if err != nil {
			return err
		}
	}

	delete(tmp, n.key())
	marked[n.key()] = true
	sorted.PushFront(n)

	return nil
}
