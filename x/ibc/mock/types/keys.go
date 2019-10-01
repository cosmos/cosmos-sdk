package types

func SequenceKey(chanid string) []byte {
	return []byte("sequence/" + chanid)
}