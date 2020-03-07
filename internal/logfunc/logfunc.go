/*
   Copyright (2020) Cobalt Speech and Language Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package logfunc provides a basic interface for logging without levels and an
// implementation that uses the go stdlib log package as its backend.
package logfunc

import (
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/cobaltspeech/log/pkg/level"
)

// Logger defines the Log function that takes key-value pairs and builds a log
// message with them.
type Logger interface {
	Log(keyvals ...interface{})
}

// jsonLogger implements the Logger interface and writes json-formatted keyvalue
// pairs to stdlib log.
type jsonLogger struct {
	l *log.Logger
}

func NewJSONLogger(w io.Writer, l level.Level) *jsonLogger {
	return &jsonLogger{log.New(w, fmt.Sprintf("%-5s ", l), log.LstdFlags)}
}

func (j jsonLogger) Log(keyvals ...interface{}) {
	n := (len(keyvals) + 1) / 2 // +1 to handle case when len is odd
	m := make(map[string]interface{}, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		var v interface{} = "missing"
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}

		key := fmt.Sprint(k)

		var val interface{}
		switch v.(type) {
		case json.Marshaler:
			val = v
		case encoding.TextMarshaler:
			val = v
		default:
			val = fmt.Sprint(v)
		}

		m[key] = val
	}

	var sb strings.Builder
	enc := json.NewEncoder(&sb)
	enc.SetEscapeHTML(false)
	err := enc.Encode(m)
	if err != nil {
		j.l.Printf("error writing log entry: %v", err)
		return
	}
	j.l.Print(sb.String())
}
