package indexer

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestUnmarshalIndexingConfig(t *testing.T) {
	cfg := &IndexingConfig{Target: map[string]Config{"target": {Type: "type"}}}
	jsonBz, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("json", func(t *testing.T) {
		res, err := unmarshalIndexingConfig(json.RawMessage(jsonBz))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(res, cfg) {
			t.Fatalf("expected %v, got %v", cfg, res)
		}
	})

	t.Run("map", func(t *testing.T) {
		var m map[string]interface{}
		err := json.Unmarshal(jsonBz, &m)
		if err != nil {
			t.Fatal(err)
		}

		res, err := unmarshalIndexingConfig(m)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(res, cfg) {
			t.Fatalf("expected %v, got %v", cfg, res)
		}
	})

	t.Run("ptr", func(t *testing.T) {
		res, err := unmarshalIndexingConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if res != cfg {
			t.Fatalf("expected %v, got %v", cfg, res)
		}
	})

	t.Run("struct", func(t *testing.T) {
		res, err := unmarshalIndexingConfig(*cfg)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(res, cfg) {
			t.Fatalf("expected %v, got %v", cfg, res)
		}
	})
}

func TestUnmarshalIndexerConfig(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		cfg := testConfig{SomeParam: "foobar"}
		cfg2, err := unmarshalIndexerCustomConfig(cfg, testConfig{})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cfg, cfg2) {
			t.Fatalf("expected %v, got %v", cfg, cfg2)
		}
	})

	t.Run("ptr", func(t *testing.T) {
		cfg := &testConfig{SomeParam: "foobar"}
		cfg2, err := unmarshalIndexerCustomConfig(cfg, &testConfig{})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cfg, cfg2) {
			t.Fatalf("expected %v, got %v", cfg, cfg2)
		}
	})

	t.Run("map -> struct", func(t *testing.T) {
		cfg := testConfig{SomeParam: "foobar"}
		jzonBz, err := json.Marshal(cfg)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]interface{}
		err = json.Unmarshal(jzonBz, &m)
		if err != nil {
			t.Fatal(err)
		}
		cfg2, err := unmarshalIndexerCustomConfig(m, testConfig{})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cfg, cfg2) {
			t.Fatalf("expected %v, got %v", cfg, cfg2)
		}
	})

	t.Run("map -> ptr", func(t *testing.T) {
		cfg := &testConfig{SomeParam: "foobar"}
		jzonBz, err := json.Marshal(cfg)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]interface{}
		err = json.Unmarshal(jzonBz, &m)
		if err != nil {
			t.Fatal(err)
		}
		cfg2, err := unmarshalIndexerCustomConfig(m, &testConfig{})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cfg, cfg2) {
			t.Fatalf("expected %v, got %v", cfg, cfg2)
		}
	})
}

type testConfig struct {
	SomeParam string `json:"some_param"`
}
