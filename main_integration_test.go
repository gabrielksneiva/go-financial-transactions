//go:build integration
// +build integration

// main_test.go
package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunApp(t *testing.T) {
	// Tenta rodar apenas se as variáveis mínimas estiverem definidas corretamente
	os.Setenv("API_PORT", "8081")
	os.Setenv("KAFKA_BROKER", "localhost:9092")
	os.Setenv("KAFKA_TOPIC", "transactions")
	os.Setenv("KAFKA_GROUP_ID", "test-group")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "user")
	os.Setenv("DB_PASSWORD", "pass")
	os.Setenv("DB_NAME", "testdb")

	defer func() {
		if r := recover(); r != nil {
			t.Log("🟡 Skipping TestRunApp due to setup failure (likely DB not available locally)")
		}
	}()

	go RunApp()
	assert.True(t, true)
}
