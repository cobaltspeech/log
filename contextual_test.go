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

package log

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/cobaltspeech/log/pkg/level"
)

func TestContextLogger(t *testing.T) {
	writelogs := func(l Logger, label string) {
		l.Trace("msg", "trace_message", "label", label)
		l.Debug("msg", "debug_message", "label", label)
		l.Info("msg", "info_message", "label", label)
		l.Error("msg", "error_message", "label", label)
	}

	var b bytes.Buffer
	logger := log.New(&b, "", 0)
	l1 := NewLeveledLogger(WithLogger(logger), WithFilterLevel(level.All))

	// first use the default logger
	writelogs(l1, "default logger")

	// now create a new logger with keyval
	l2 := With(l1, "key1", "value1")
	writelogs(l2, "With(key1)")

	l3 := With(l1, "key2", "value2")
	writelogs(l3, "With(key2)")

	l4 := With(l2, "key3", "value3")
	writelogs(l4, "With(key1,key3)")

	want := `
trace {"label":"default logger","msg":"trace_message"}
debug {"label":"default logger","msg":"debug_message"}
info  {"label":"default logger","msg":"info_message"}
error {"label":"default logger","msg":"error_message"}
trace {"key1":"value1","label":"With(key1)","msg":"trace_message"}
debug {"key1":"value1","label":"With(key1)","msg":"debug_message"}
info  {"key1":"value1","label":"With(key1)","msg":"info_message"}
error {"key1":"value1","label":"With(key1)","msg":"error_message"}
trace {"key2":"value2","label":"With(key2)","msg":"trace_message"}
debug {"key2":"value2","label":"With(key2)","msg":"debug_message"}
info  {"key2":"value2","label":"With(key2)","msg":"info_message"}
error {"key2":"value2","label":"With(key2)","msg":"error_message"}
trace {"key1":"value1","key3":"value3","label":"With(key1,key3)","msg":"trace_message"}
debug {"key1":"value1","key3":"value3","label":"With(key1,key3)","msg":"debug_message"}
info  {"key1":"value1","key3":"value3","label":"With(key1,key3)","msg":"info_message"}
error {"key1":"value1","key3":"value3","label":"With(key1,key3)","msg":"error_message"}
`
	if got := b.String(); strings.TrimSpace(got) != strings.TrimSpace(want) {
		t.Errorf("default filter level: got %q, want %q", got, want)
	}
}
