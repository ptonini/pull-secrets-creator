package kac

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const (
	configFile = "../config.yaml"
)

func Test_Config(t *testing.T) {

	ctx := context.Background()
	ctx = context.WithValue(ctx, keyFakeClientSet, true)

	t.Run("read invalid config", func(t *testing.T) {
		assert.Error(t, LoadConfig("invalid"))
	})

	t.Run("read config from file", func(t *testing.T) {
		_ = LoadConfig(configFile)
		_, err := getConfig()
		assert.NoError(t, err)
	})

	t.Run("read config from envs", func(t *testing.T) {
		_ = os.Setenv("IMAGE_PULL_SECRETS", `{"test-credentials": "eyJhdXRocyI6IHsiaHR0cHM6Ly9kb2NrZXIuaW8iOiB7ImF1dGgiOiAiZEdWemREcDBaWE4wWlFvPSJ9fX0K"}`)
		_ = LoadConfig(configFile)
		_, err := getConfig()
		assert.NoError(t, err)
	})

}
