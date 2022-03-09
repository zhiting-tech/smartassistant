package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigTest(t *testing.T) {
	assert.True(t, GetConf().Debug)
	assert.True(t, alreadyInitConfig)
	assert.Equal(t, options, *GetConf())
	assert.Equal(t, "zt.registry.zhitingtech.com", GetConf().SmartAssistant.DockerRegistry)
}

func TestMain(m *testing.M) {
	TestSetup()
	code := m.Run()
	TestTeardown()
	os.Exit(code)
}
