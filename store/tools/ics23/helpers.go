package ics23_tools

import (
	"fmt"
	"strconv"

	tmproofs "github.com/cosmos/cosmos-sdk/store/internal/proofs"
)

func ParseArgs(args []string) (exist bool, loc tmproofs.Where, size int, err error) {
	if len(args) != 3 && len(args) != 4 {
		err = fmt.Errorf("Insufficient args")
		return
	}

	switch args[1] {
	case "exist":
		exist = true
	case "nonexist":
		exist = false
	default:
		err = fmt.Errorf("Invalid arg: %s", args[1])
		return
	}

	switch args[2] {
	case "left":
		loc = tmproofs.Left
	case "middle":
		loc = tmproofs.Middle
	case "right":
		loc = tmproofs.Right
	default:
		err = fmt.Errorf("Invalid arg: %s", args[2])
		return
	}

	size = 400
	if len(args) == 4 {
		size, err = strconv.Atoi(args[3])
	}

	return
}

func ParseBatchArgs(args []string) (size int, exist int, nonexist int, err error) {
	if len(args) != 3 {
		err = fmt.Errorf("Insufficient args")
		return
	}

	size, err = strconv.Atoi(args[0])
	if err != nil {
		return
	}
	exist, err = strconv.Atoi(args[1])
	if err != nil {
		return
	}
	nonexist, err = strconv.Atoi(args[2])
	return
}
