package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	tmproofs "github.com/cosmos/cosmos-sdk/store/internal/proofs"

	ics23 "github.com/confio/ics23/go"
)

/**
testgen-simple will generate a json struct on stdout (meant to be saved to file for testdata).
this will be an auto-generated existence proof in the form:

{
	"root": "<hex encoded root hash of tree>",
	"key": "<hex encoded key to prove>",
	"value": "<hex encoded value to prove> (empty on non-existence)",
	"proof": "<hex encoded protobuf marshaling of a CommitmentProof>"
}

It accepts two or three arguments (optional size: default 400)

  testgen-simple [exist|nonexist] [left|right|middle] <size>

If you make a batch, we have a different format:

{
	"root": "<hex encoded root hash of tree>",
	"proof": "<hex encoded protobuf marshaling of a CommitmentProof (Compressed Batch)>",
	"items": [{
		"key": "<hex encoded key to prove>",
		"value": "<hex encoded value to prove> (empty on non-existence)",
	}, ...]
}

The batch variant accepts 5 arguments:

  testgen-simple [batch] [size] [num exist] [num nonexist]
**/

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: testgen-simple batch [size] [# exist] [# nonexist]")
		fmt.Println("       testgen-simple [exist|nonexist] [left|right|middle] <size>")
		os.Exit(1)
	}

	if os.Args[1] == "batch" {
		err := doBatch(os.Args[2:])
		if err != nil {
			fmt.Printf("%+v\n", err)
			fmt.Println("Usage: testgen-simple [batch] [size] [# exist] [# nonexist]")
			os.Exit(1)
		}
		return
	}

	exist, loc, size, err := parseArgs(os.Args)
	if err != nil {
		fmt.Printf("%+v\n", err)
		fmt.Println("Usage: testgen-simple [exist|nonexist] [left|right|middle] <size>")
		os.Exit(1)
	}

	data := tmproofs.BuildMap(size)
	allkeys := tmproofs.SortedKeys(data)
	root := tmproofs.CalcRoot(data)

	var key, value []byte
	if exist {
		key = []byte(tmproofs.GetKey(allkeys, loc))
		value = data[string(key)]
	} else {
		key = []byte(tmproofs.GetNonKey(allkeys, loc))
	}

	var proof *ics23.CommitmentProof
	if exist {
		proof, err = tmproofs.CreateMembershipProof(data, key)
	} else {
		proof, err = tmproofs.CreateNonMembershipProof(data, key)
	}
	if err != nil {
		fmt.Printf("Error: create proof: %+v\n", err)
		os.Exit(1)
	}

	binary, err := proof.Marshal()
	if err != nil {
		fmt.Printf("Error: protobuf marshal: %+v\n", err)
		os.Exit(1)
	}

	res := map[string]interface{}{
		"root":  hex.EncodeToString(root),
		"key":   hex.EncodeToString(key),
		"value": hex.EncodeToString(value),
		"proof": hex.EncodeToString(binary),
	}
	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		fmt.Printf("Error: json encoding: %+v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(out))
}

func parseArgs(args []string) (exist bool, loc tmproofs.Where, size int, err error) {
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

type item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func doBatch(args []string) error {
	size, exist, nonexist, err := parseBatchArgs(args)
	if err != nil {
		return err
	}

	data := tmproofs.BuildMap(size)
	allkeys := tmproofs.SortedKeys(data)
	root := tmproofs.CalcRoot(data)

	items := []item{}
	proofs := []*ics23.CommitmentProof{}

	for i := 0; i < exist; i++ {
		key := []byte(tmproofs.GetKey(allkeys, tmproofs.Middle))
		value := data[string(key)]
		proof, err := tmproofs.CreateMembershipProof(data, key)
		if err != nil {
			return fmt.Errorf("create proof: %+v", err)
		}
		proofs = append(proofs, proof)
		item := item{
			Key:   hex.EncodeToString(key),
			Value: hex.EncodeToString(value),
		}
		items = append(items, item)
	}

	for i := 0; i < nonexist; i++ {
		key := []byte(tmproofs.GetNonKey(allkeys, tmproofs.Middle))
		proof, err := tmproofs.CreateNonMembershipProof(data, key)
		if err != nil {
			return fmt.Errorf("create proof: %+v", err)
		}
		proofs = append(proofs, proof)
		item := item{
			Key: hex.EncodeToString(key),
		}
		items = append(items, item)
	}

	// combine all proofs into batch and compress
	proof, err := ics23.CombineProofs(proofs)
	if err != nil {
		fmt.Printf("Error: combine proofs: %+v\n", err)
		os.Exit(1)
	}

	binary, err := proof.Marshal()
	if err != nil {
		fmt.Printf("Error: protobuf marshal: %+v\n", err)
		os.Exit(1)
	}

	res := map[string]interface{}{
		"root":  hex.EncodeToString(root),
		"items": items,
		"proof": hex.EncodeToString(binary),
	}
	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		fmt.Printf("Error: json encoding: %+v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(out))

	return nil
}

func parseBatchArgs(args []string) (size int, exist int, nonexist int, err error) {
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
