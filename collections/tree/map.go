package tree

import (
	"bytes"
	"context"
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/codec"
)

const (
	RootSuffix       uint8 = 0x0
	ValuesSuffix     uint8 = 0x1
	TreeSuffix       uint8 = 0x2
	RootNameSuffix         = "_root"
	ValuesNameSuffix       = "_values"
	TreeNameSuffix         = "_tree"
)

type Map[K, V any] struct {
	root   collections.Item[uint64]
	values collections.Map[K, V]
	tree   collections.Vec[Node[K]]
}

func NewMap[K, V any](sb *collections.SchemaBuilder, prefix collections.Prefix, namespace string, kc codec.KeyCodec[K], vc codec.ValueCodec[V]) Map[K, V] {
	root := collections.NewItem(sb, collections.NewPrefix(0), namespace+RootNameSuffix, collections.Uint64Value)
	values := collections.NewMap(sb, collections.NewPrefix(1), namespace+ValuesNameSuffix, kc, vc)
	tree := collections.NewVec(sb, collections.NewPrefix(2), namespace+TreeNameSuffix, NewNodeEncoder(kc))

	return Map[K, V]{
		root:   root,
		values: values,
		tree:   tree,
	}
}

func (t Map[K, V]) len(ctx context.Context) (uint64, error) {
	return t.tree.Len(ctx)
}

func (t Map[K, V]) getNode(ctx context.Context, id uint64) (Node[K], error) {
	return t.tree.Get(ctx, id)
}

// saveNode will replace node n at the provided index, or will push it to the
// tree.
func (t Map[K, V]) saveNode(ctx context.Context, n Node[K]) error {
	length, err := t.len(ctx)
	if err != nil {
		return err
	}
	if n.Id < length {
		return t.tree.Replace(ctx, n.Id, n)
	}
	return t.tree.Push(ctx, n)
}

// Has reports if the tree has the provided key.
func (t Map[K, V]) Has(ctx context.Context, key K) (bool, error) {
	return t.values.Has(ctx, key)
}

// Get returns the value in the tree.
func (t Map[K, V]) Get(ctx context.Context, key K) (V, error) {
	return t.values.Get(ctx, key)
}

// Set adds to, or updates the tree with the provided key value pair.
func (t Map[K, V]) Set(ctx context.Context, key K, value V) error {
	exists, err := t.Has(ctx, key)
	if err != nil {
		return err
	}
	if !exists {
		length, err := t.len(ctx)
		if err != nil {
			return err
		}
		root, err := t.getRoot(ctx)
		if err != nil {
			return err
		}
		newRoot, err := t.setAt(ctx, root, length, key)
		if err != nil {
			return err
		}
		err = t.root.Set(ctx, newRoot) // TODO: check if equal old root and new root
		if err != nil {
			return err
		}
	}
	return t.values.Set(ctx, key, value)
}

// Min returns the lowest key in the tree.
func (t Map[K, V]) Min(ctx context.Context) (k K, err error) {
	root, err := t.getRoot(ctx)
	if err != nil {
		return k, err
	}
	minNode, _, err := t.minAt(ctx, root, root)
	if err != nil {
		return k, err
	}
	return minNode.Key, nil
}

// Max returns the biggest key in the tree.
func (t Map[K, V]) Max(ctx context.Context) (k K, err error) {
	root, err := t.getRoot(ctx)
	if err != nil {
		return k, err
	}
	maxNode, _, err := t.maxAt(ctx, root, root)
	if err != nil {
		return k, err
	}
	return maxNode.Key, nil
}

// Ceil returns the lowest key which is greater or equal than the provided one.
func (t Map[K, V]) Ceil(ctx context.Context, key K) (k *K, err error) {
	has, err := t.Has(ctx, key)
	if err != nil {
		return nil, err
	}
	if has {
		return &key, nil
	}
	return t.Higher(ctx, key)
}

// Higher finds the lowest key in the tree which is bigger than the key provided.
// A nil return means no higher key exists.
func (t Map[K, V]) Higher(ctx context.Context, key K) (k *K, err error) {
	root, err := t.getRoot(ctx)
	if err != nil {
		return
	}
	return t.aboveAt(ctx, root, key)
}

// Floor returns the smallest key which
func (t Map[K, V]) Floor(ctx context.Context, key K) (*K, error) {
	panic("impl")
}

// Lower finds the biggest key which is smaller than the key provided.
func (t Map[K, V]) Lower(ctx context.Context, key K) (k *K, err error) {
	root, err := t.getRoot(ctx)
	if err != nil {
		return nil, err
	}
	return t.belowAt(ctx, root, key)
}

func (t Map[K, V]) minAt(ctx context.Context, atId, parentId uint64) (Node[K], Node[K], error) {
	parentNode, err := t.getNode(ctx, parentId)
	if err != nil {
		return Node[K]{}, Node[K]{}, err
	}
	for {
		node, err := t.getNode(ctx, atId)
		if err != nil {
			return Node[K]{}, Node[K]{}, err
		}

		if node.Left != nil {
			atId = *node.Left
			parentId = node.Id
		} else {
			return node, parentNode, nil
		}
	}
}

func (t Map[K, V]) maxAt(ctx context.Context, atId, parentId uint64) (Node[K], Node[K], error) {
	parentNode, err := t.getNode(ctx, parentId)
	if err != nil {
		return Node[K]{}, Node[K]{}, err
	}
	for {
		node, err := t.getNode(ctx, atId)
		if err != nil {
			return Node[K]{}, Node[K]{}, err
		}
		if node.Right != nil {
			parentNode = node
			atId = *node.Right
		} else {
			return node, parentNode, nil
		}
	}
}

func (t Map[K, V]) aboveAt(ctx context.Context, at uint64, key K) (*K, error) {
	var seen *K
	for {
		node, err := t.getNode(ctx, at)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				break
			}
			return nil, err
		}
		cmp, err := t.cmp(node.Key, key)
		if err != nil {
			return nil, err
		}
		if cmp == cmpLower || cmp == cmpEqual {
			if node.Right != nil {
				at = *node.Right
			} else {
				break
			}
		} else {
			seen = &node.Key
			if node.Left != nil {
				at = *node.Left
			} else {
				break
			}
		}
	}
	return seen, nil
}

func (t Map[K, V]) belowAt(ctx context.Context, at uint64, key K) (seen *K, err error) {
	for {
		node, err := t.getNode(ctx, at)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				break
			}
			return nil, err
		}

		cmp, err := t.cmp(key, node.Key)
		if err != nil {
			return nil, err
		}
		if cmp == cmpGreater {
			seen = &node.Key
			if node.Right == nil {
				break
			} else {
				at = *node.Right
			}
		} else {
			if node.Left != nil {
				at = *node.Left
			} else {
				break
			}
		}
	}
	return seen, nil
}

func (t Map[K, V]) setAt(ctx context.Context, at uint64, id uint64, key K) (uint64, error) {
	node, err := t.getNode(ctx, at)
	if err != nil {
		// it is some error we cannot handle.
		if !errors.Is(err, collections.ErrNotFound) {
			return 0, err
		}
		// the node was not found, then we simply save the node
		// and return the current root, this also means this is the first
		// node, wo-oho.
		node = Node[K]{
			Id:     id,
			Left:   nil,
			Right:  nil,
			Height: 1,
			Key:    key,
		}
		err = t.saveNode(ctx, node)
		if err != nil {
			return 0, err
		}
		return at, err
	}

	// if the node already exists, then we're updating.
	cmp, err := t.cmp(key, node.Key)
	if err != nil {
		return 0, err
	}
	// if it's equal then the tree does not need to be updated
	// only the value does, but that is stored in a separate
	// partition of state.
	if cmp == cmpEqual {
		return at, nil
	}
	// if lower
	if cmp == cmpLower {
		setAt := id
		if node.Left != nil {
			setAt = *node.Left
		}
		idx, err := t.setAt(ctx, setAt, id, key)
		if err != nil {
			return 0, err
		}
		node.Left = &idx
	}
	// if higher
	if cmp == cmpGreater {
		setAt := id
		if node.Right != nil {
			setAt = *node.Right
		}
		idx, err := t.setAt(ctx, setAt, id, key)
		if err != nil {
			return 0, err
		}
		node.Right = &idx
	}
	err = t.updateHeight(ctx, &node)
	if err != nil {
		return 0, err
	}
	return t.enforceBalance(ctx, node)
}

// updateHeight calculates and saves the height of a subtree at node `at`.
// height[at] = 1 + max(height[at.Left], height[at.Right])
func (t Map[K, V]) updateHeight(ctx context.Context, node *Node[K]) (err error) {
	var lft, rgt uint64
	if node.Left != nil {
		lftNode, err := t.getNode(ctx, *node.Left)
		if err != nil {
			return err
		}
		lft = lftNode.Height
	}
	if node.Right != nil {
		rightNode, err := t.getNode(ctx, *node.Right)
		if err != nil {
			return err
		}
		rgt = rightNode.Height
	}

	node.Height = 1 + max(lft, rgt)
	return t.saveNode(ctx, *node)
}

func (t Map[K, V]) enforceBalance(ctx context.Context, node Node[K]) (uint64, error) {
	balance, err := t.getBalance(ctx, node)
	if err != nil {
		return 0, err
	}
	if balance > 1 {
		if node.Left != nil {
			lft, err := t.getNode(ctx, *node.Left)
			if err != nil {
				return 0, err
			}
			lftBalance, err := t.getBalance(ctx, lft)
			if err != nil {
				return 0, err
			}
			if lftBalance < 0 {
				rotated, err := t.rotateRight(ctx, &lft)
				if err != nil {
					return 0, err
				}
				node.Left = &rotated
			}
		}
		return t.rotateLeft(ctx, &node)
	}
	if balance < -1 {
		if node.Right != nil {
			rgt, err := t.getNode(ctx, *node.Right)
			if err != nil {
				return 0, err
			}
			rightBalance, err := t.getBalance(ctx, rgt)
			if err != nil {
				return 0, err
			}
			if rightBalance > 0 {
				rotated, err := t.rotateLeft(ctx, &rgt)
				if err != nil {
					return 0, err
				}
				node.Right = &rotated
			}
		}
		return t.rotateRight(ctx, &node)
	}
	return node.Id, nil
}

// getBalance calculates the balance factor of a node, defined as the
// difference in heights between left and right subtrees.
func (t Map[K, V]) getBalance(ctx context.Context, node Node[K]) (int64, error) {
	var leftHeight, rightHeight uint64
	if node.Left != nil {
		leftNode, err := t.getNode(ctx, *node.Left)
		if err != nil {
			return 0, err
		}
		leftHeight = leftNode.Height
	}
	if node.Right != nil {
		rightNode, err := t.getNode(ctx, *node.Right)
		if err != nil {
			return 0, err
		}
		rightHeight = rightNode.Height
	}
	return int64(leftHeight - rightHeight), nil
}

// rotateLeft performs a left rotation on an AVL tree subtree rooted at the given node.
// It returns the new root of the subtree, and the caller is responsible for updating the proper link from the parent.
func (t Map[K, V]) rotateLeft(ctx context.Context, node *Node[K]) (uint64, error) {
	leftNode, err := t.getNode(ctx, *node.Left)
	if err != nil {
		return 0, err
	}
	leftRight := leftNode.Right
	node.Left = leftRight
	leftNode.Right = &node.Id

	err = t.updateHeight(ctx, node)
	if err != nil {
		return 0, err
	}

	err = t.updateHeight(ctx, &leftNode)
	if err != nil {
		return 0, err
	}
	return leftNode.Id, nil
}

func (t Map[K, V]) rotateRight(ctx context.Context, node *Node[K]) (uint64, error) {
	rightNode, err := t.getNode(ctx, *node.Right)
	if err != nil {
		return 0, err
	}
	rightLeft := rightNode.Left
	node.Right = rightLeft
	rightNode.Left = &node.Id
	err = t.updateHeight(ctx, node)
	if err != nil {
		return 0, err
	}
	err = t.updateHeight(ctx, &rightNode)
	if err != nil {
		return 0, err
	}
	return rightNode.Id, nil
}

// cmp compares two keys and returns the comparison factor.
func (t Map[K, V]) cmp(key1 K, key2 K) (cmp, error) {
	key1B, err := collections.EncodeKeyWithPrefix(nil, t.values.KeyCodec(), key1)
	if err != nil {
		return 0, err
	}
	key2B, err := collections.EncodeKeyWithPrefix(nil, t.values.KeyCodec(), key2)
	if err != nil {
		return 0, err
	}

	switch bytes.Compare(key1B, key2B) {
	case 0:
		return cmpEqual, nil
	case -1:
		return cmpLower, nil
	case +1:
		return cmpGreater, nil
	}
	panic("unreachable")
}

func (t Map[K, V]) getRoot(ctx context.Context) (uint64, error) {
	root, err := t.root.Get(ctx)
	switch {
	case err == nil:
		return root, nil
	case errors.Is(err, collections.ErrNotFound):
		return 0, nil
	default:
		return 0, err
	}
}

type cmp uint8

const (
	cmpEqual cmp = iota
	cmpGreater
	cmpLower
)

type mapIter[K, V any] struct {
	t          Map[K, V]
	startKey   K
	endKey     K
	currentKey K
}

// forwardConsume implements forward iteration.
func (v mapIter[K, V]) forwardConsume(ctx context.Context, fn func(key K, value V) (stop bool, err error)) error {
	for {
		// get value
		value, err := v.t.values.Get(ctx, v.currentKey)
		if err != nil {
			return err
		}
		// do callback
		stop, err := fn(v.currentKey, value)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
		// find next value: TODO: consider using Map.aboveAt to keep track of the latest node, is this traversing the tree on every call to next???? :O
		k, err := v.t.Higher(ctx, v.currentKey)
		if err != nil {
			return err
		}
		// iteration finished, because there are no more values.
		if k == nil {
			return nil
		}
		// check if next value is > than end
		cmp, err := v.t.cmp(*k, v.endKey)
		if err != nil {
			return err
		}
		// iteration finished, next value is greater or equal to end.
		if cmp == cmpGreater {
			return nil
		}
		// set next key
		v.currentKey = *k
	}
}

func (t Map[K, V]) Walk(ctx context.Context, rng collections.Ranger[K], walkFn func(key K, value V) (stop bool, err error)) (err error) {
	var (
		start *collections.RangeKey[K]
		end   *collections.RangeKey[K]
		order = collections.OrderAscending
	)
	if rng != nil {
		start, end, order, err = rng.RangeValues()
		if err != nil {
			return err
		}
	}
	// get start key
	startKey, err := t.getStartKey(ctx, start)
	if err != nil {
		return err
	}
	// invalid iteration.
	if startKey == nil {
		return nil
	}
	// get end key
	endKey, err := t.getEndKey(ctx, end)
	if err != nil {
		return err
	}

	iter := mapIter[K, V]{
		t:          t,
		endKey:     *endKey,
		currentKey: *startKey,
	}
	if order == collections.OrderAscending {
		return iter.forwardConsume(ctx, walkFn)
	} else {
		panic("not impl")
	}
}

// TODO: this uses the tree to find start and end, I do not think this is ideal, we should
// implement something in KeyCodec that already does the bytes manipulation needed.
func (t Map[K, V]) getStartKey(ctx context.Context, start *collections.RangeKey[K]) (k *K, err error) {
	if start == nil {
		m, err := t.Min(ctx)
		return &m, err
	}
	if start.Kind() == collections.RangeKeyExactKind {
		// find start key, iteration by design is start inclusive,
		// so we want as first key either the provided one or the
		// one exactly after.
		return t.Ceil(ctx, start.Key())
	}
	// TODO: what's missing here is exclusive
	panic("not impl")
}

// look getStartKey
func (t Map[K, V]) getEndKey(ctx context.Context, end *collections.RangeKey[K]) (k *K, err error) {
	if end == nil {
		// return maximum
		m, err := t.Max(ctx)
		return &m, err
	}
	// if exact it means that we need to find the key below the current one.
	if end.Kind() == collections.RangeKeyExactKind {
		return t.Lower(ctx, end.Key())
	}
	panic("impl")
}
