package types

// Convert int64 array into a map
func ConvertArrayToMap(arr []int64) map[int64]bool {
	eleMap := make(map[int64]bool)

	// put slice values into map
	for _, s := range arr {
		eleMap[s] = true
	}

	return eleMap
}
