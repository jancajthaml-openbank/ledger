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
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func loadConfFromEnv() Configuration {
	logLevel := strings.ToUpper(getEnvString("LEDGER_LOG_LEVEL", "DEBUG"))
	secrets := getEnvString("LEDGER_SECRETS", "")
	rootStorage := getEnvString("LEDGER_STORAGE", "/data")
	lakeHostname := getEnvString("LEDGER_LAKE_HOSTNAME", "")
	port := getEnvInteger("LEDGER_HTTP_PORT", 4401)
	minFreeDiskSpace := getEnvInteger("VAULT_STORAGE_THRESHOLD", 0)
	minFreeMemory := getEnvInteger("VAULT_MEMORY_THRESHOLD", 0)
	metricsOutput := getEnvFilename("LEDGER_METRICS_OUTPUT", "/tmp")
	metricsRefreshRate := getEnvDuration("LEDGER_METRICS_REFRESHRATE", time.Second)

	if lakeHostname == "" || secrets == "" || rootStorage == "" {
		log.Fatal("missing required parameter to run")
	}

	return Configuration{
		RootStorage:        rootStorage,
		ServerPort:         port,
		SecretsPath:        secrets,
		LakeHostname:       lakeHostname,
		LogLevel:           logLevel,
		MetricsRefreshRate: metricsRefreshRate,
		MetricsOutput:      metricsOutput,
		MinFreeDiskSpace:   uint64(minFreeDiskSpace),
		MinFreeMemory:      uint64(minFreeMemory),
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

func getEnvInteger(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	cast, err := strconv.Atoi(value)
	if err != nil {
		log.Panicf("invalid value of variable %s", key)
	}
	return cast
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
