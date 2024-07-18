package decoding

import (
	"errors"
	"testing"

	"cosmossdk.io/schema"
)

type modA struct{}

func (m modA) ModuleCodec() (schema.ModuleCodec, error) {
	return schema.ModuleCodec{
		Schema: schema.ModuleSchema{ObjectTypes: []schema.ObjectType{{Name: "A"}}},
	}, nil
}

type modB struct{}

func (m modB) ModuleCodec() (schema.ModuleCodec, error) {
	return schema.ModuleCodec{
		Schema: schema.ModuleSchema{ObjectTypes: []schema.ObjectType{{Name: "B"}}},
	}, nil
}

type modC struct{}

var moduleSet = map[string]interface{}{
	"modA": modA{},
	"modB": modB{},
	"modC": modC{},
}

var testResolver = ModuleSetDecoderResolver(moduleSet)

func TestModuleSetDecoderResolver_IterateAll(t *testing.T) {
	objectTypes := map[string]bool{}
	err := testResolver.IterateAll(func(moduleName string, cdc schema.ModuleCodec) error {
		objectTypes[cdc.Schema.ObjectTypes[0].Name] = true
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(objectTypes) != 2 {
		t.Fatalf("expected 2 object types, got %d", len(objectTypes))
	}

	if !objectTypes["A"] {
		t.Fatalf("expected object type A")
	}

	if !objectTypes["B"] {
		t.Fatalf("expected object type B")
	}
}

func TestModuleSetDecoderResolver_LookupDecoder(t *testing.T) {
	decoder, found, err := testResolver.LookupDecoder("modA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !found {
		t.Fatalf("expected to find decoder for modA")
	}

	if decoder.Schema.ObjectTypes[0].Name != "A" {
		t.Fatalf("expected object type A, got %s", decoder.Schema.ObjectTypes[0].Name)
	}

	decoder, found, err = testResolver.LookupDecoder("modB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !found {
		t.Fatalf("expected to find decoder for modB")
	}

	if decoder.Schema.ObjectTypes[0].Name != "B" {
		t.Fatalf("expected object type B, got %s", decoder.Schema.ObjectTypes[0].Name)
	}

	decoder, found, err = testResolver.LookupDecoder("modC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if found {
		t.Fatalf("expected not to find decoder")
	}

	decoder, found, err = testResolver.LookupDecoder("modD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if found {
		t.Fatalf("expected not to find decoder")
	}
}

type modD struct{}

func (m modD) ModuleCodec() (schema.ModuleCodec, error) {
	return schema.ModuleCodec{}, errors.New("an error")
}

func TestModuleSetDecoderResolver_IterateAll_Error(t *testing.T) {
	resolver := ModuleSetDecoderResolver(map[string]interface{}{
		"modD": modD{},
	})
	err := resolver.IterateAll(func(moduleName string, cdc schema.ModuleCodec) error {
		if moduleName == "modD" {
			t.Fatalf("expected error")
		}
		return nil
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}
