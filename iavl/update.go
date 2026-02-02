package iavl

type KVUpdate = struct {
	Key, Value []byte
	Delete     bool
}
