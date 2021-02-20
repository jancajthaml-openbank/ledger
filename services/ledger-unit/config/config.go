// Copyright (c) 2016-2021, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"github.com/jancajthaml-openbank/ledger-unit/support/env"
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
	// MetricsStastdEndpoint represents statsd daemon hostname
	MetricsStastdEndpoint string
	// TransactionIntegrityScanInterval represents backoff between scan for
	// non terminal transactions
	TransactionIntegrityScanInterval time.Duration
}

// LoadConfig loads application configuration
func LoadConfig() Configuration {
	return Configuration{
		Tenant:                           env.String("LEDGER_TENANT", ""),
		LakeHostname:                     env.String("LEDGER_LAKE_HOSTNAME", "127.0.0.1"),
		RootStorage:                      env.String("LEDGER_STORAGE", "/data") + "/" + "t_" + env.String("LEDGER_TENANT", ""),
		LogLevel:                         strings.ToUpper(env.String("LEDGER_LOG_LEVEL", "INFO")),
		TransactionIntegrityScanInterval: env.Duration("LEDGER_TRANSACTION_INTEGRITY_SCANINTERVAL", 5*time.Minute),
		MetricsStastdEndpoint:            env.String("LEDGER_STATSD_ENDPOINT", "127.0.0.1:8125"),
	}
}
