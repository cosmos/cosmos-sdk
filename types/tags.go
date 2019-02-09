package types

import (
	"fmt"
	"strings"

	cmn "github.com/tendermint/tendermint/libs/common"
)

// Type synonym for convenience
type Tag = cmn.KVPair

// Type synonym for convenience
type Tags cmn.KVPairs

// New empty tags
func EmptyTags() Tags {
	return make(Tags, 0)
}

// Append a single tag
func (t Tags) AppendTag(k string, v string) Tags {
	return append(t, MakeTag(k, v))
}

// Append two lists of tags
func (t Tags) AppendTags(tags Tags) Tags {
	return append(t, tags...)
}

// Turn tags into KVPair list
func (t Tags) ToKVPairs() []cmn.KVPair {
	return []cmn.KVPair(t)
}

// New variadic tags, must be k string, v []byte repeating
func NewTags(tags ...interface{}) Tags {
	var ret Tags
	if len(tags)%2 != 0 {
		panic("must specify key-value pairs as varargs")
	}
	i := 0
	for {
		if i == len(tags) {
			break
		}
		ret = append(ret, Tag{Key: toBytes(tags[i]), Value: toBytes(tags[i+1])})
		i += 2
	}
	return ret
}

func toBytes(i interface{}) []byte {
	switch x := i.(type) {
	case []uint8:
		return x
	case string:
		return []byte(x)
	default:
		panic(i)
	}
}

// Make a tag from a key and a value
func MakeTag(k string, v string) Tag {
	return Tag{Key: []byte(k), Value: []byte(v)}
}

//__________________________________________________

// common tags
var (
	TagAction       = "action"
	TagSrcValidator = "source-validator"
	TagDstValidator = "destination-validator"
	TagDelegator    = "delegator"
)

// A KVPair where the Key and Value are both strings, rather than []byte
type StringTag struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

func (st StringTag) String() string {
	return fmt.Sprintf("%s = %s", st.Key, st.Value)
}

// A slice of StringTag
type StringTags []StringTag

func (st StringTags) String() string {
	var sb strings.Builder
	for _, t := range st {
		sb.WriteString(fmt.Sprintf("    - %s\n", t.String()))
	}
	return sb.String()
}

// Conversion function from a []byte tag to a string tag
func TagToStringTag(tag Tag) StringTag {
	return StringTag{
		Key:   string(tag.Key),
		Value: string(tag.Value),
	}
}

// Conversion function from Tags to a StringTags
func TagsToStringTags(tags Tags) StringTags {
	var stringTags StringTags
	for _, tag := range tags {
		stringTags = append(stringTags, TagToStringTag(tag))
	}
	return stringTags
}
