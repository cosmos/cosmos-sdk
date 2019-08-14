package ibc

import "strconv"

const Version int64 = 1

func VersionPrefix(version int64) []byte {
	return []byte(strconv.FormatInt(version, 10) + "/")
}
