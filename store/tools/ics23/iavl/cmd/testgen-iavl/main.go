package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	ics23 "github.com/confio/ics23/go"

	tmproofs "github.com/cosmos/cosmos-sdk/store/internal/proofs"
	tools "github.com/cosmos/cosmos-sdk/store/tools/ics23"
	iavlproofs "github.com/cosmos/cosmos-sdk/store/tools/ics23/iavl"
	"github.com/cosmos/cosmos-sdk/store/tools/ics23/iavl/helpers"
)

/**
testgen-iavl will generate a json struct on stdout (meant to be saved to file for testdata).
this will be an auto-generated existence proof in the form:

{
	"root": "<hex encoded root hash of tree>",
	"key": "<hex encoded key to prove>",
	"value": "<hex encoded value to prove> (empty on non-existence)",
	"proof": "<hex encoded protobuf marshaling of a CommitmentProof>"
}

It accepts two or three arguments (optional size: default 400)

  testgen-iavl [exist|nonexist] [left|right|middle] <size>

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

  testgen-iavl [batch] [size] [num exist] [num nonexist]
**/

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: testgen-iavl batch [size] [# exist] [# nonexist]")
		fmt.Println("       testgen-iavl [exist|nonexist] [left|right|middle] <size>")
		os.Exit(1)
	}

	if os.Args[1] == "batch" {
		err := doBatch(os.Args[2:])
		if err != nil {
			fmt.Printf("%+v\n", err)
			fmt.Println("Usage: testgen-iavl [batch] [size] [# exist] [# nonexist]")
			os.Exit(1)
		}
		return
	}

	exist, loc, size, err := tools.ParseArgs(os.Args)
	if err != nil {
		fmt.Printf("%+v\n", err)
		fmt.Println("Usage: testgen-iavl [exist|nonexist] [left|right|middle] <size>")
		os.Exit(1)
	}

	tree, allkeys, err := helpers.BuildTree(size)
	if err != nil {
		fmt.Printf("%+v\n", err)
		fmt.Println("Usage: testgen-iavl [exist|nonexist] [left|right|middle] <size>")
		os.Exit(1)
	}
	root := tree.WorkingHash()

	var key, value []byte
	if exist {
		key = helpers.GetKey(allkeys, loc)
		_, value = tree.Get(key)
	} else {
		key = helpers.GetNonKey(allkeys, loc)
	}

	var proof *ics23.CommitmentProof
	if exist {
		proof, err = iavlproofs.CreateMembershipProof(tree, key)
	} else {
		proof, err = iavlproofs.CreateNonMembershipProof(tree, key)
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

type item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func doBatch(args []string) error {
	size, exist, nonexist, err := tools.ParseBatchArgs(args)
	if err != nil {
		return err
	}

	tree, allkeys, err := helpers.BuildTree(size)
	if err != nil {
		return err
	}
	root := tree.WorkingHash()

	items := []item{}
	proofs := []*ics23.CommitmentProof{}

	for i := 0; i < exist; i++ {
		key := []byte(helpers.GetKey(allkeys, tmproofs.Middle))
		_, value := tree.Get(key)
		proof, err := iavlproofs.CreateMembershipProof(tree, key)
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
		key := []byte(helpers.GetNonKey(allkeys, tmproofs.Middle))
		proof, err := iavlproofs.CreateNonMembershipProof(tree, key)
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
