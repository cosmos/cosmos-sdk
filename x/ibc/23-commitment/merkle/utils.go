package merkle

/*

func (prefix Prefix) Key(key []byte) []byte {
	return join(prefix.KeyPrefix, key)
}

func join(a, b []byte) (res []byte) {
	res = make([]byte, len(a)+len(b))
	copy(res, a)
	copy(res[len(a):], b)
	return
}

func (prefix Prefix) Path() string {
	pathstr := ""
	for _, inter := range prefix.KeyPath {
		// The Queryable() stores uses slash-separated keypath format for querying
		pathstr = pathstr + "/" + string(inter)
	}

	return pathstr
}
*/
