// +build !release

package config

import (
	_ "embed"
)

//go:embed smartassistant.default.debug.yaml
var DefaultSmartassistantConfigData []byte
