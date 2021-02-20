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
	"strings"
	"github.com/jancajthaml-openbank/ledger-rest/support/env"
)

// Configuration of application
type Configuration struct {
	// RootStorage gives where to store journals
	RootStorage string
	// ServerPort is port which server is bound to
	ServerPort int
	// ServerKey path to server tls key file
	ServerKey string
	// ServerCert path to server tls cert file
	ServerCert string
	// LakeHostname represent hostname of openbank lake service
	LakeHostname string
	// LogLevel ignorecase log level
	LogLevel string
	// MinFreeDiskSpace respresents threshold for minimum disk free space to
	// be possible operating
	MinFreeDiskSpace uint64
	// MinFreeMemory respresents threshold for minimum available memory to
	// be possible operating
	MinFreeMemory uint64
}

// LoadConfig loads application configuration
func LoadConfig() Configuration {
	return Configuration{
		RootStorage:      env.String("LEDGER_STORAGE", "/data"),
		ServerPort:       env.Int("LEDGER_HTTP_PORT", 4401),
		ServerKey:        env.String("LEDGER_SERVER_KEY", ""),
		ServerCert:       env.String("LEDGER_SERVER_CERT", ""),
		LakeHostname:     env.String("LEDGER_LAKE_HOSTNAME", "127.0.0.1"),
		LogLevel:         strings.ToUpper(env.String("LEDGER_LOG_LEVEL", "INFO")),
		MinFreeDiskSpace: env.Uint64("VAULT_STORAGE_THRESHOLD", 0),
		MinFreeMemory:    env.Uint64("VAULT_MEMORY_THRESHOLD", 0),
	}
}
