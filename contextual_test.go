/*
   Copyright (2021) Cobalt Speech and Language Inc.

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
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/cobaltspeech/log/pkg/level"
)

func TestContextLogger(t *testing.T) {
	writelogs := func(l Logger, label string) {
		l.Trace("trace_message", "label", label)
		l.Debug("debug_message", "label", label)
		l.Info("info_message", "label", label)
		l.Error("error_message", errors.New("the_error"), "label", label)
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

	// create a new logger with empty keyvals
	l5 := With(l2)
	writelogs(l5, "With(key1,empty)")

	want := `trace {"msg":"trace_message","label":"default logger"}
debug {"msg":"debug_message","label":"default logger"}
info  {"msg":"info_message","label":"default logger"}
error {"msg":"error_message","error":"the_error","label":"default logger"}
trace {"msg":"trace_message","key1":"value1","label":"With(key1)"}
debug {"msg":"debug_message","key1":"value1","label":"With(key1)"}
info  {"msg":"info_message","key1":"value1","label":"With(key1)"}
error {"msg":"error_message","error":"the_error","key1":"value1","label":"With(key1)"}
trace {"msg":"trace_message","key2":"value2","label":"With(key2)"}
debug {"msg":"debug_message","key2":"value2","label":"With(key2)"}
info  {"msg":"info_message","key2":"value2","label":"With(key2)"}
error {"msg":"error_message","error":"the_error","key2":"value2","label":"With(key2)"}
trace {"msg":"trace_message","key1":"value1","key3":"value3","label":"With(key1,key3)"}
debug {"msg":"debug_message","key1":"value1","key3":"value3","label":"With(key1,key3)"}
info  {"msg":"info_message","key1":"value1","key3":"value3","label":"With(key1,key3)"}
error {"msg":"error_message","error":"the_error","key1":"value1","key3":"value3","label":"With(key1,key3)"}
trace {"msg":"trace_message","key1":"value1","label":"With(key1,empty)"}
debug {"msg":"debug_message","key1":"value1","label":"With(key1,empty)"}
info  {"msg":"info_message","key1":"value1","label":"With(key1,empty)"}
error {"msg":"error_message","error":"the_error","key1":"value1","label":"With(key1,empty)"}
`
	if got := b.String(); strings.TrimSpace(got) != strings.TrimSpace(want) {
		t.Log(got)
		t.Log(want)
		t.Errorf("default filter level: got %q, want %q", got, want)
	}
}
