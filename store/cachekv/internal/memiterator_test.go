package internal

import (
	"testing"
)

func TestMemIterator_Ascending(t *testing.T) {
	db := NewBTree()
	// db.set()
	db.Set([]byte("a"), []byte("value_a"))
	db.Set([]byte("b"), []byte("value_b"))
	db.Set([]byte("c"), []byte("value_c"))

	iterator := newMemIterator([]byte("a"), []byte("c"), db, true)

	var result []string
	for iterator.Valid() {
		result = append(result, string(iterator.Key()))
		iterator.Next()
	}

	expected := []string{"a", "b", "c"}
	for i, key := range result {
		if key != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], key)
		}
	}

	if iterator.Valid() {
		t.Errorf("Iterator should be invalid after last item")
	}
}

func TestMemIterator_Descending(t *testing.T) {
	db := NewBTree()

	db.Set([]byte("a"), []byte("value_a"))
	db.Set([]byte("b"), []byte("value_b"))
	db.Set([]byte("c"), []byte("value_c"))
	db.Set([]byte("d"), []byte("value_d"))

	iterator := newMemIterator([]byte("a"), []byte("d"), db, false)

	var result []string
	for iterator.Valid() {
		result = append(result, string(iterator.Key()))
		iterator.Next()
	}

	expected := []string{"c", "b", "a"}
	for i, key := range result {
		if key != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], key)
		}
	}

	if iterator.Valid() {
		t.Errorf("Iterator should be invalid after last item")
	}
}

func TestMemIterator_EmptyRange(t *testing.T) {
	db := NewBTree()
	db.Set([]byte("a"), []byte("value_a"))
	db.Set([]byte("b"), []byte("value_b"))
	db.Set([]byte("c"), []byte("value_c"))

	iterator := newMemIterator([]byte("d"), []byte("e"), db, true)

	if iterator.Valid() {
		t.Errorf("Iterator should be invalid for empty range")
	}
}
