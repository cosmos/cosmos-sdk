package tests

import (
	"encoding/json"

	"cosmossdk.io/indexer/postgres"
	"cosmossdk.io/schema/indexer"
)

func postgresConfigToIndexerConfig(cfg postgres.Config) (indexer.Config, error) {
	cfgBz, err := json.Marshal(cfg)
	if err != nil {
		return indexer.Config{}, err
	}

	var cfgMap map[string]interface{}
	err = json.Unmarshal(cfgBz, &cfgMap)
	if err != nil {
		return indexer.Config{}, err
	}

	return indexer.Config{
		Type:   "postgres",
		Config: cfgMap,
	}, nil
}
