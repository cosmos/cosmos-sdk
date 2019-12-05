package subspace

func concatKeys(key, subkey []byte) []byte {
	res := make([]byte, len(key)+1+len(subkey))
	copy(res, key)
	res[len(key)] = '/'
	copy(res[len(key)+1:], subkey)

	return res
}

func isAlphaNumeric(key []byte) bool {
	for _, b := range key {
		if !((48 <= b && b <= 57) || // numeric
			(65 <= b && b <= 90) || // upper case
			(97 <= b && b <= 122)) { // lower case
			return false
		}
	}

	return true
}
