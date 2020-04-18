package types

import (
	"time"
)

// DONTCOVER

// The Double Sign Jail period ends at Max Time supported by Amino
// (Dec 31, 9999 - 23:59:59 GMT).
var DoubleSignJailEndTime = time.Unix(253402300799, 0)
