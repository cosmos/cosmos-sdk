package anycompress

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTrieGetSet(t *testing.T) {
	nt := newTrie()
	nt.set([]byte("/testutil/testdata.proto.Animal"), []byte("texas"))
	nt.set([]byte("/testutil/testdata.proto.Animal.fox"), []byte("california"))
	got, err := nt.get([]byte("/testutil/testdata.proto.Animal"))
	if err != nil {
		t.Fatalf("Unexpectedly got an error: %v", err)
	}
	want := []byte("texas")
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("Failed to match: got - want +\n%s", diff)
	}

	np, _ := nt.longestPrefix([]byte("/testutil/testdata.proto.Animal^$"))
	if np == nil {
		t.Fatal("Unexpectedly failed to find the longest matching prefix")
	}
	if diff := cmp.Diff(np.value, want); diff != "" {
		t.Fatalf("Longest prefix matching failed: got - want +\n%s", diff)
	}

	np, _ = nt.longestPrefix([]byte("/testutil/testdata.proto.Animal.fo\xff"))
	if np == nil {
		t.Fatal("Unexpectedly failed to find the longest matching prefix")
	}
	if diff := cmp.Diff(np.value, []byte(nil)); diff != "" {
		t.Fatalf("Longest prefix matching failed: got - want +\n%s", diff)
	}

}
