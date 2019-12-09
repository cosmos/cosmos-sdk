package types

// ConvertIntArrayToInt64 converts array elements of type int to int64
func ConvertIntArrayToInt64(blockHeight []int) []int64 {
	blockHeightInt64 := make([]int64, len(blockHeight))
	for i, height := range blockHeight {
		blockHeightInt64[i] = int64(height)
	}
	return blockHeightInt64
}

// Contains returns true if the array contains the given element
func Contains(blockHeightArray []int64, skipHeight int64) bool {
	for _, height := range blockHeightArray {
		if height == skipHeight {
			return true
		}
	}
	return false
}
