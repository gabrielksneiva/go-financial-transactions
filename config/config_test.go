package config_test

import (
	"os"
	"testing"

	"github.com/gabrielksneiva/go-financial-transactions/config"
	"github.com/stretchr/testify/assert"
)

func TestGetEnv_WithEnvVarSet(t *testing.T) {
	os.Setenv("MY_TEST_VAR", "custom_value")
	defer os.Unsetenv("MY_TEST_VAR")

	val := config.GetEnv("MY_TEST_VAR", "default")
	assert.Equal(t, "custom_value", val)
}

func TestGetEnv_WithEnvVarNotSet(t *testing.T) {
	val := config.GetEnv("UNSET_ENV_VAR", "default_value")
	assert.Equal(t, "default_value", val)
}

func TestLoadConfig_DefaultRedisDB(t *testing.T) {
	os.Setenv("REDIS_DB", "3") // simula um .env válido
	defer os.Unsetenv("REDIS_DB")

	cfg := config.LoadConfig()
	assert.Equal(t, 3, cfg.RedisDB)
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "real_value")
	defer os.Unsetenv("TEST_KEY")

	assert.Equal(t, "real_value", config.GetEnv("TEST_KEY", "default"))
	assert.Equal(t, "default", config.GetEnv("UNDEFINED_KEY", "default"))
}

func TestLoadConfig(t *testing.T) {
	os.Setenv("API_PORT", "9999")
	os.Setenv("REDIS_DB", "2")
	os.Setenv("DB_HOST", "localhost")
	// ... defina os principais

	cfg := config.LoadConfig()

	assert.Equal(t, "9999", cfg.APIPort)
	assert.Equal(t, 2, cfg.RedisDB)
	assert.Equal(t, "localhost", cfg.DBHost)
}
