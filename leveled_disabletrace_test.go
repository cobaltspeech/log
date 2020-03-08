// +build !cobalt_log_trace

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

func TestLeveledLogger_SetFilterLevel(t *testing.T) {
	writelogs := func(l Logger) {
		l.Trace("msg", "trace_message")
		l.Debug("msg", "debug_message")
		l.Info("msg", "info_message")
		l.Error("msg", "error_message")
	}

	var b bytes.Buffer
	logger := log.New(&b, "", 0)
	l := NewLeveledLogger(WithLogger(logger))

	// first use the default filter level
	writelogs(l)

	// now create a new logger with filter level set
	l = NewLeveledLogger(WithLogger(logger), WithFilterlevel(level.Debug))
	writelogs(l)

	l.SetFilterLevel(level.Info)
	writelogs(l)

	l.SetFilterLevel(level.Info | level.Debug)
	writelogs(l)

	l.SetFilterLevel(level.None)
	writelogs(l)

	l.SetFilterLevel(level.All)
	writelogs(l)

	want := `
info  {"msg":"info_message"}
error {"msg":"error_message"}
debug {"msg":"debug_message"}
info  {"msg":"info_message"}
debug {"msg":"debug_message"}
info  {"msg":"info_message"}
debug {"msg":"debug_message"}
info  {"msg":"info_message"}
error {"msg":"error_message"}
`
	if got := b.String(); strings.TrimSpace(got) != strings.TrimSpace(want) {
		t.Errorf("default filter level: got %q, want %q", got, want)
	}
}
