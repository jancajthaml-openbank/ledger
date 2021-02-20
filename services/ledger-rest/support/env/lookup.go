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

package env

import (
	"strconv"
	"os"
)

// Get retrieves the string value of the environment variable named by the key
func Get(key string) (string, bool) {
	if v := os.Getenv(key); v != "" {
		return v, true
	}
	return "", false
}

// String retrieves the string value from environment named by the key.
func String(key string, fallback string) string {
	if str, exists := Get(key); exists {
		return str
	}
	return fallback
}

// Int retrieves integer value from the environment.
func Int(key string, fallback int) int {
	if str, exists := Get(key); exists {
		v, err := strconv.ParseInt(str, 10, 0)
		if err != nil {
			log.Warn().Msgf("invalid value in %s, using fallback", key)
			return fallback
		}
		return int(v)
	}
	return fallback
}

// Uint64 retrieves 64-bit unsigned integer value from the environment.
func Uint64(key string, fallback uint64) uint64 {
	if str, exists := Get(key); exists {
		v, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			log.Warn().Msgf("invalid value in %s, using fallback", key)
			return fallback
		}
		return v
	}
	return fallback
}
