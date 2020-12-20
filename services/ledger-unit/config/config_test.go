package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestGetConfig(t *testing.T) {
	for _, v := range os.Environ() {
		k := strings.Split(v, "=")[0]
		if strings.HasPrefix(k, "LEDGER") {
			os.Unsetenv(k)
		}
	}

	t.Log("has defaults for all values")
	{
		config := LoadConfig()

		if config.Tenant != "" {
			t.Errorf("Tenant default value is not empty")
		}
		if config.LakeHostname != "127.0.0.1" {
			t.Errorf("LakeHostname default value is not 127.0.0.1")
		}
		if config.RootStorage != "/data/t_" {
			t.Errorf("RootStorage default value is not /data/t_")
		}
		if config.LogLevel != "INFO" {
			t.Errorf("LogLevel default value is not INFO")
		}
		if config.TransactionIntegrityScanInterval != 5*time.Minute {
			t.Errorf("TransactionIntegrityScanInterval default value is not 5m")
		}
		if config.MetricsStastdEndpoint != "127.0.0.1:8125" {
			t.Errorf("MetricsStastdEndpoint default value is not 127.0.0.1:8125")
		}
	}
}
