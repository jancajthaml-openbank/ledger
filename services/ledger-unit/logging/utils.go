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
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

// NewLogger returns logger with preset field
func NewLogger(name string) logrus.FieldLogger {
	return logrus.WithField("src", name)
}

// SetupLogger properly sets up logging
func SetupLogger(level string) {
	if logLevel, err := logrus.ParseLevel(level); err == nil {
		logrus.Infof("Log level set to %v", strings.ToUpper(level))
		logrus.SetLevel(logLevel)
	} else {
		logrus.Warnf("Invalid log level %v, using level WARN", level)
		logrus.SetLevel(logrus.WarnLevel)
	}
	logrus.SetOutput(os.Stdout)
}
