package testing

func PrefixedClientKey(clientID string, key []byte) []byte {
	return append([]byte("clients/"+clientID+"/"), key...)
}
