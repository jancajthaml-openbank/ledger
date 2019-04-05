// Copyright (c) 2016-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	// Secrets represents cerificate .key
	SecretKey []byte
	// Secrets represents cerificate .crt
	SecretCert []byte
	// LakeHostname represent hostname of openbank lake service
	LakeHostname string
	// LogOutput represents log output
	LogOutput string
	// LogLevel ignorecase log level
	LogLevel string
	// MetricsRefreshRate represents interval in which in memory metrics should be
	// persisted to disk
	MetricsRefreshRate time.Duration
	// MetricsOutput represents output file for metrics persistence
	MetricsOutput string
}

// Resolver loads config
type Resolver interface {
	GetConfig() Configuration
}

type configResolver struct {
	cfg Configuration
}

// NewResolver provides config resolver
func NewResolver() Resolver {
	return configResolver{cfg: loadConfFromEnv()}
}

// GetConfig loads application configuration
func (c configResolver) GetConfig() Configuration {
	return c.cfg
}
