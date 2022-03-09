// +build release

package config

import (
	_ "embed"
)

//go:embed smartassistant.default.release.yaml
var DefaultSmartassistantConfigData []byte
