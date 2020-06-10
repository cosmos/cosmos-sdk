package testing

func prefixedClientKey(clientID string, key []byte) []byte {
	return append([]byte("clients/"+clientID+"/"), key...)
}
