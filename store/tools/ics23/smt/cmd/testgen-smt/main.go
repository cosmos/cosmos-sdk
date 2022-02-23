package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	ics23 "github.com/confio/ics23/go"

	tmproofs "github.com/cosmos/cosmos-sdk/store/internal/proofs"
	tools "github.com/cosmos/cosmos-sdk/store/tools/ics23"
	smtproofs "github.com/cosmos/cosmos-sdk/store/tools/ics23/smt"
	"github.com/cosmos/cosmos-sdk/store/tools/ics23/smt/helpers"
)

/**
testgen-smt will generate a json struct on stdout (meant to be saved to file for testdata).
this will be an auto-generated existence proof in the form:

{
	"root": "<hex encoded root hash of tree>",
	"key": "<hex encoded key to prove>",
	"value": "<hex encoded value to prove> (empty on non-existence)",
	"proof": "<hex encoded protobuf marshaling of a CommitmentProof>"
}

It accepts two or three arguments (optional size: default 400)

  testgen-smt [exist|nonexist] [left|right|middle] <size>

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

  testgen-smt [batch] [size] [num exist] [num nonexist]
**/

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: testgen-smt batch [size] [# exist] [# nonexist]")
		fmt.Println("       testgen-smt [exist|nonexist] [left|right|middle] <size>")
		os.Exit(1)
	}

	if os.Args[1] == "batch" {
		size, exist, nonexist, err := tools.ParseBatchArgs(os.Args[2:])
		if err != nil {
			fmt.Printf("%+v\n", err)
			fmt.Println("Usage: testgen-smt batch [size] [# exist] [# nonexist]")
			os.Exit(1)
		}
		err = doBatch(size, exist, nonexist)
		if err != nil {
			fmt.Printf("Error: %+v\n", err)
			os.Exit(1)
		}
		return
	}

	exist, loc, size, err := tools.ParseArgs(os.Args)
	if err != nil {
		fmt.Printf("%+v\n", err)
		fmt.Println("Usage: testgen-smt [exist|nonexist] [left|right|middle] <size>")
		os.Exit(1)
	}
	err = doSingle(exist, loc, size)
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		os.Exit(1)
	}
}

func doSingle(exist bool, loc tmproofs.Where, size int) error {
	tree, preim, err := helpers.BuildTree(size)
	if err != nil {
		return err
	}
	root := tree.Root()

	var key, value []byte
	if exist {
		key = preim.GetKey(loc)
		value, err = tree.Get(key)
		if err != nil {
			return fmt.Errorf("get key: %w", err)
		}
	} else {
		key = preim.GetNonKey(loc)
	}

	var proof *ics23.CommitmentProof
	if exist {
		proof, err = smtproofs.CreateMembershipProof(tree, key)
	} else {
		proof, err = smtproofs.CreateNonMembershipProof(tree, key, preim)
	}
	if err != nil {
		return fmt.Errorf("create proof: %w", err)
	}

	binary, err := proof.Marshal()
	if err != nil {
		return fmt.Errorf("protobuf marshal: %w", err)
	}

	path := sha256.Sum256(key)
	res := map[string]interface{}{
		"root":  hex.EncodeToString(root),
		"key":   hex.EncodeToString(path[:]),
		"value": hex.EncodeToString(value),
		"proof": hex.EncodeToString(binary),
	}
	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("json encoding: %w", err)
	}

	fmt.Println(string(out))
	return nil
}

type item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func pickWhere(i int) tmproofs.Where {
	if i > 2 {
		return tmproofs.Middle
	}
	return tmproofs.Where(i)
}

func doBatch(size, exist, nonexist int) error {
	tree, preim, err := helpers.BuildTree(size)
	if err != nil {
		return err
	}
	root := tree.Root()

	items := []item{}
	proofs := []*ics23.CommitmentProof{}

	for i := 0; i < exist; i++ {
		where := pickWhere(i)
		key := []byte(preim.GetKey(where))
		value, err := tree.Get(key)
		if err != nil {
			return fmt.Errorf("get key: %w", err)
		}
		proof, err := smtproofs.CreateMembershipProof(tree, key)
		if err != nil {
			return fmt.Errorf("create proof: %w", err)
		}
		proofs = append(proofs, proof)
		path := sha256.Sum256(key)
		item := item{
			Key:   hex.EncodeToString(path[:]),
			Value: hex.EncodeToString(value),
		}
		items = append(items, item)
	}

	for i := 0; i < nonexist; i++ {
		where := pickWhere(i)
		key := []byte(preim.GetNonKey(where))
		proof, err := smtproofs.CreateNonMembershipProof(tree, key, preim)
		if err != nil {
			return fmt.Errorf("create proof: %w", err)
		}
		proofs = append(proofs, proof)
		path := sha256.Sum256(key)
		item := item{
			Key: hex.EncodeToString(path[:]),
		}
		items = append(items, item)
	}

	// combine all proofs into batch and compress
	proof, err := ics23.CombineProofs(proofs)
	if err != nil {
		return fmt.Errorf("combine proofs: %w", err)
	}

	binary, err := proof.Marshal()
	if err != nil {
		return fmt.Errorf("protobuf marshal: %w", err)
	}

	res := map[string]interface{}{
		"root":  hex.EncodeToString(root),
		"items": items,
		"proof": hex.EncodeToString(binary),
	}
	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("json encoding: %w", err)
	}

	fmt.Println(string(out))

	return nil
}
