package iavl

type Update struct {
	Key, Value []byte
	Delete     bool
}
