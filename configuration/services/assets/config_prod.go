//go:build !dev

package assets

import (
	"embed"
	"jellyfin-duplicate/constants"

	"github.com/sirupsen/logrus"
)

//go:embed  config.prod.json
var embedFS embed.FS

func GetConfigFile() ([]byte, error) {
	if config, err := embedFS.ReadFile("config.prod.json"); err != nil {
		logrus.Warnf("Failed to read embedded config: %v", err)
		return nil, err
	} else {
		return config, nil
	}
}

func GetEnv() constants.Environment {
	return constants.Production
}
