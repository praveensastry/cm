package config_test

import (
	"testing"

	"github.com/praveensastry/cm/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigRead(t *testing.T) {
	getCfg := func() {
		config.ReadConfig()
	}

	assert.NotPanics(t, getCfg)
}
