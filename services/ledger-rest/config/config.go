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

import "time"

// Configuration of application
type Configuration struct {
	// RootStorage gives where to store journals
	RootStorage string
	// ServerPort is port which server is bound to
	ServerPort int
	// SecretsPath directory where .key and .crt is stored
	SecretsPath string
	// LakeHostname represent hostname of openbank lake service
	LakeHostname string
	// LogLevel ignorecase log level
	LogLevel string
	// MetricsRefreshRate represents interval in which in memory metrics should be
	// persisted to disk
	MetricsRefreshRate time.Duration
	// MetricsOutput represents output file for metrics persistence
	MetricsOutput string
	// MinFreeDiskSpace respresents threshold for minimum disk free space to
	// be possible operating
	MinFreeDiskSpace uint64
	// MinFreeMemory respresents threshold for minimum available memory to
	// be possible operating
	MinFreeMemory uint64
}

// GetConfig loads application configuration
func GetConfig() Configuration {
	return loadConfFromEnv()
}
