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

package logging

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

var log = New("global")

// New returns logger with preset field
func New(name string) zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	return zerolog.New(output).With().Timestamp().Str("src", name).Logger()
}

// SetupLogger properly sets up logging
func SetupLogger(level string) {
	switch level {
	case "DEBUG":
		log.Info().Msg("Log level set to DEBUG")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		log.Info().Msg("Log level set to INFO")
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "ERROR":
		log.Info().Msg("Log level set to ERROR")
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		log.Warn().Msgf("Invalid log level %v, using level INFO", level)
		log.Info().Msg("Log level set to INFO")
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
