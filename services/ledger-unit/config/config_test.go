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
		if config.MetricsContinuous != true {
			t.Errorf("MetricsContinuous default value is not true")
		}
		if config.MetricsRefreshRate != time.Second {
			t.Errorf("MetricsRefreshRate default value is not 1s")
		}
		if config.MetricsOutput != "/tmp/ledger-unit-metrics" {
			t.Errorf("MetricsOutput default value is not /tmp/ledger-unit-metrics")
		}
		if config.TransactionIntegrityScanInterval != 5*time.Minute {
			t.Errorf("TransactionIntegrityScanInterval default value is not 5m")
		}
	}
}
