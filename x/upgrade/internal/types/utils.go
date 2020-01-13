package types

import (
	"os"
	"path/filepath"
)

func EnsureConfigExists(upgradeInfoFileDir string) (string, error) {
	if _, err := os.Stat(upgradeInfoFileDir); os.IsNotExist(err) {
		err = os.Mkdir(upgradeInfoFileDir, os.ModePerm)
		if err != nil {
			return "", err
		}
		return filepath.Join(upgradeInfoFileDir, "upgrade-info.json"), nil
	}
	return filepath.Join(upgradeInfoFileDir, "upgrade-info.json"), nil
}
