package mapping

type Queue struct {
	Indexer
}

func NewQueue(base Base, prefix []byte, enc IndexEncoding) Queue {
	return Queue{
		Indexer: NewIndexer(base, prefix, enc),
	}
}

func (q Queue) Prefix(prefix []byte) Queue {
	return Queue{
		Indexer: q.Indexer.Prefix(prefix),
	}
}

func (q Queue) Peek(ctx Context, ptr interface{}) (ok bool) {
	_, ok = q.First(ctx, ptr)
	return
}

func (q Queue) Pop(ctx Context, ptr interface{}) (ok bool) {
	key, ok := q.First(ctx, ptr)
	if !ok {
		return
	}

	q.Delete(ctx, key)
	return
}

func (q Queue) Push(ctx Context, o interface{}) {
	key, ok := q.Last(ctx, nil)
	if !ok {
		key = 0
	} else {
		key = key + 1
	}

	q.Set(ctx, key, o)
}
