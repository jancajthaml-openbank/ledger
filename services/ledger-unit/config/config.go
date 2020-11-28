// Copyright (c) 2016-2020, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"strings"
	"time"
)

// Configuration of application
type Configuration struct {
	// Tenant represent tenant of given vault
	Tenant string
	// LakeHostname represent hostname of openbank lake service
	LakeHostname string
	// RootStorage gives where to store journals
	RootStorage string
	// LogLevel ignorecase log level
	LogLevel string
	// MetricsRefreshRate represents interval in which in memory metrics should be
	// persisted to disk
	MetricsRefreshRate time.Duration
	// MetricsOutput represents output file for metrics persistence
	MetricsOutput string
	// TransactionIntegrityScanInterval represents backoff between scan for
	// non terminal transactions
	TransactionIntegrityScanInterval time.Duration
}

// LoadConfig loads application configuration
func LoadConfig() Configuration {
	return Configuration{
		Tenant:                           envString("LEDGER_TENANT", ""),
		LakeHostname:                     envString("LEDGER_LAKE_HOSTNAME", "127.0.0.1"),
		RootStorage:                      envString("LEDGER_STORAGE", "/data") + "/" + "t_" + envString("LEDGER_TENANT", ""),
		LogLevel:                         strings.ToUpper(envString("LEDGER_LOG_LEVEL", "INFO")),
		MetricsRefreshRate:               envDuration("LEDGER_METRICS_REFRESHRATE", time.Second),
		MetricsOutput:                    envFilename("LEDGER_METRICS_OUTPUT", "/tmp/ledger-unit-metrics"),
		TransactionIntegrityScanInterval: envDuration("LEDGER_TRANSACTION_INTEGRITY_SCANINTERVAL", 5*time.Minute),
	}
}
