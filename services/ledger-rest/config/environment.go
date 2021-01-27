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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func envBoolean(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	cast, err := strconv.ParseBool(value)
	if err != nil {
		log.Warn().Msgf("invalid value in %s, using fallback", key)
		return fallback
	}
	return cast
}

func envFilename(key string, fallback string) string {
	var value = strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	value = filepath.Clean(value)
	if os.MkdirAll(value, os.ModePerm) != nil {
		log.Warn().Msgf("invalid value in %s, using fallback", key)
		return fallback
	}
	return value
}

func envString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInteger(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	cast, err := strconv.Atoi(value)
	if err != nil {
		log.Warn().Msgf("invalid value in %s, using fallback", key)
		return fallback
	}
	return cast
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	cast, err := time.ParseDuration(value)
	if err != nil {
		log.Warn().Msgf("invalid value in %s, using fallback", key)
		return fallback
	}
	return cast
}
