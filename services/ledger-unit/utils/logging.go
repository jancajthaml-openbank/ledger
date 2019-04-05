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

package utils

import (
	"bytes"

	log "github.com/sirupsen/logrus"
)

// LogFormat is a custom logging formater for logrus
type LogFormat struct {
}

var (
	debugPrefix = []byte("DEBU ")
	infoPrefix  = []byte("INFO ")
	warnPrefix  = []byte("WARN ")
	fatalPrefix = []byte("FATA ")
	errorPrefix = []byte("ERRO ")
	panicPrefix = []byte("PANI ")
)

// Format processed each log entry and produces formatted log line
func (f *LogFormat) Format(entry *log.Entry) ([]byte, error) {
	if entry.Message == "" {
		return nil, nil
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = new(bytes.Buffer)
	}

	switch entry.Level {
	case log.DebugLevel:
		b.Write(debugPrefix)
	case log.InfoLevel:
		b.Write(infoPrefix)
	case log.WarnLevel:
		b.Write(warnPrefix)
	case log.ErrorLevel:
		b.Write(errorPrefix)
	case log.FatalLevel:
		b.Write(fatalPrefix)
	case log.PanicLevel:
		b.Write(panicPrefix)
	}

	b.WriteString(entry.Message)
	b.WriteByte('\n')

	return b.Bytes(), nil
}
