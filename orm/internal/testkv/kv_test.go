package testkv

import "testing"

func TestNew(t *testing.T) {
	kv := New()

	kv.Set(nil, []byte("key1"), []byte("value"))
	iter := kv.IteratePrefix(nil, []byte("key"))
	t.Logf("%v", iter.Valid())
	t.Logf("%s", iter.Key())
}
