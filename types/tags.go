package types

import (
	cmn "github.com/tendermint/tmlibs/common"
)

type Tag = cmn.KVPair

type Tags = cmn.KVPairs

// Append two lists of tags
func AppendTags(a, b Tags) Tags {
	return append(a, b...)
}

// New empty tags
func EmptyTags() Tags {
	return make(Tags, 0)
}

// Single tag to tags
func SingleTag(t Tag) Tags {
	return append(EmptyTags(), t)
}

// Make a tag from a key and a value
func MakeTag(k string, v []byte) Tag {
	return Tag{Key: []byte(k), Value: v}
}

// TODO: Deduplication?
