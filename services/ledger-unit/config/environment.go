// Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func loadConfFromEnv() Configuration {
	logLevel := strings.ToUpper(getEnvString("LEDGER_LOG_LEVEL", "DEBUG"))
	storage := getEnvString("LEDGER_STORAGE", "/data")
	tenant := getEnvString("LEDGER_TENANT", "")
	lakeHostname := getEnvString("LEDGER_LAKE_HOSTNAME", "")
	transactionIntegrityScanInterval := getEnvDuration("LEDGER_TRANSACTION_INTEGRITY_SCANINTERVAL", time.Minute)
	metricsOutput := getEnvFilename("LEDGER_METRICS_OUTPUT", "/tmp")
	metricsRefreshRate := getEnvDuration("LEDGER_METRICS_REFRESHRATE", time.Second)

	if tenant == "" || lakeHostname == "" || storage == "" {
		log.Fatal("missing required parameter to run")
	}

	return Configuration{
		Tenant:                           tenant,
		LakeHostname:                     lakeHostname,
		RootStorage:                      storage + "/" + "t_" + tenant,
		LogLevel:                         logLevel,
		MetricsRefreshRate:               metricsRefreshRate,
		MetricsOutput:                    metricsOutput,
		TransactionIntegrityScanInterval: transactionIntegrityScanInterval,
	}
}

func getEnvFilename(key, fallback string) string {
	var value = strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	value = filepath.Clean(value)
	if os.MkdirAll(value, os.ModePerm) != nil {
		return fallback
	}
	return value
}

func getEnvString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	cast, err := time.ParseDuration(value)
	if err != nil {
		log.Panicf("invalid value of variable %s", key)
	}
	return cast
}
