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
	"bufio"
	"bytes"
	"errors"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/cobaltspeech/log/pkg/level"
)

func TestLeveledLogger(t *testing.T) {
	// LeveledLogger writes to os.Stderr by default; patch it to retrieve output
	var b bytes.Buffer

	osStderr = &b

	defer func() {
		// reset osStderr
		osStderr = os.Stderr
	}()

	l := NewLeveledLogger()
	logAndTest(t, l, &b)
}

func TestLeveledLogger_WithOutput(t *testing.T) {
	var b bytes.Buffer
	l := NewLeveledLogger(WithOutput(&b))
	logAndTest(t, l, &b)
}

func logAndTest(t *testing.T, l Logger, b *bytes.Buffer) {
	t.Helper()

	l.Error("test_message", errors.New("the_error"))

	wantJSON := `{"msg":"test_message","error":"the_error"}`
	rDate := `[0-9][0-9][0-9][0-9]/[0-9][0-9]/[0-9][0-9]`
	rTime := `[0-9][0-9]:[0-9][0-9]:[0-9][0-9]`
	pattern := "^" + rDate + " " + rTime + " error " + wantJSON + "\n"

	matched, err := regexp.Match(pattern, b.Bytes())
	if err != nil {
		t.Fatalf("pattern %q did not compile: %s", pattern, err)
	}

	if !matched {
		t.Error("message did not match pattern")
	}
}

func TestLeveledLogger_SetFilterLevel(t *testing.T) {
	writelogs := func(l Logger, label string) {
		l.Trace("trace_message", "label", label)
		l.Debug("debug_message", "label", label)
		l.Info("info_message", "label", label)
		l.Error("error_message", errors.New("the_error"), "label", label)
	}

	var b bytes.Buffer
	logger := log.New(&b, "", 0)
	l := NewLeveledLogger(WithLogger(logger))

	// first use the default filter level
	writelogs(l, "default filter level")

	// now create a new logger with filter level set
	l = NewLeveledLogger(WithLogger(logger), WithFilterLevel(level.Debug))
	writelogs(l, "WithFilterLevel(debug)")

	l.SetFilterLevel(level.Info)
	writelogs(l, "SetFilterLevel(Info)")

	l.SetFilterLevel(level.Info | level.Debug)
	writelogs(l, "SetFilterLevel(Info|Debug)")

	l.SetFilterLevel(level.None)
	writelogs(l, "SetFilterLevel(None)")

	l.SetFilterLevel(level.All)
	writelogs(l, "SetFilterLevel(All)")

	want := `info  {"msg":"info_message","label":"default filter level"}
error {"msg":"error_message","error":"the_error","label":"default filter level"}
debug {"msg":"debug_message","label":"WithFilterLevel(debug)"}
info  {"msg":"info_message","label":"SetFilterLevel(Info)"}
debug {"msg":"debug_message","label":"SetFilterLevel(Info|Debug)"}
info  {"msg":"info_message","label":"SetFilterLevel(Info|Debug)"}
trace {"msg":"trace_message","label":"SetFilterLevel(All)"}
debug {"msg":"debug_message","label":"SetFilterLevel(All)"}
info  {"msg":"info_message","label":"SetFilterLevel(All)"}
error {"msg":"error_message","error":"the_error","label":"SetFilterLevel(All)"}
`
	if got := b.String(); strings.TrimSpace(got) != strings.TrimSpace(want) {
		t.Log(got)
		t.Log(want)
		t.Errorf("default filter level: got %q, want %q", got, want)
	}
}

func TestLeveledLogger_log_jsonErrors(t *testing.T) {
	var b bytes.Buffer

	logger := log.New(&b, "", 0)
	l := NewLeveledLogger(WithLogger(logger))

	l.Info("t1")
	l.Info("t2", "msg")
	l.Info("t3", "test this")
	l.Info("t4", "failingTextMarshaller", &failingTextMarshaler{})
	l.Info("t5", "failingJSONMarshaler", &failingJSONMarshaler{})

	//nolint: lll // log truth can't be broken into multiple lines
	want := `info  {"msg":"t1"}
info  {"msg":"t2","msg":"missing"}
info  {"msg":"t3","test this":"missing"}
error {"msg":"logging failure","error":"json: error calling MarshalJSON for type logmap.MapSlice: json: error calling MarshalText for type *log.failingTextMarshaler: invalid value"}
error {"msg":"logging failure","error":"json: error calling MarshalJSON for type logmap.MapSlice: json: error calling MarshalJSON for type *log.failingJSONMarshaler: invalid value"}
`

	if got := b.String(); strings.TrimSpace(got) != strings.TrimSpace(want) {
		t.Log(got)
		t.Log(want)
		t.Errorf("log_jsonErrors: got %q, want %q", got, want)
	}
}

func TestLeveledLogger_Concurrent(t *testing.T) {
	// this test verifies that LeveledLogger can be called concurrently.
	// This is done by creating N concurrent logging calls and ensuring that
	// the output has exactly N log lines.
	var b bytes.Buffer
	logger := log.New(&b, "", 0)
	l := NewLeveledLogger(WithLogger(logger))

	var wg sync.WaitGroup

	N := 100
	wg.Add(N)

	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()

			l.Error("concurrent_logging_test", errors.New("the_error"), "i", i)
		}(i)
	}
	wg.Wait()

	scanner := bufio.NewScanner(&b)

	var i int

	for i = 0; scanner.Scan(); i++ {
		x := strings.Fields(scanner.Text())
		if len(x) != 2 {
			t.Fatalf("invalid line %q", scanner.Text())
		}
	}

	if i != N {
		t.Errorf("incorrect lines in concurrent output: got %d, want %d", i, N)
	}
}

// failingTextMarshaler implements encoding.TextMarshaler that fails
type failingTextMarshaler struct{}

func (v *failingTextMarshaler) MarshalText() ([]byte, error) {
	return nil, errInvalidValue
}

// failingJSONMarshaler implements json.Marshaler that fails
type failingJSONMarshaler struct{}

func (v *failingJSONMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errInvalidValue
}

var errInvalidValue = errors.New("invalid value")

type LError struct {
	msg     string
	keyvals []interface{}
}

func (err LError) Error() string {
	return err.msg
}

func (err LError) ErrorValues() []interface{} {
	return err.keyvals
}

func TestLeveledLogger_LoggableErrors(t *testing.T) {
	var b bytes.Buffer
	logger := log.New(&b, "", 0)
	l := NewLeveledLogger(WithLogger(logger))

	l.Error("error_message", errors.New("the_error"), "key1", "val1")
	le := LError{msg: "the_error", keyvals: []interface{}{"err.key1", "err.val1"}}
	l.Error("error_message", le, "key1", "val1")
	le2 := LError{msg: "the_error", keyvals: []interface{}{"err.key1", "err.val1", "err.key2", "err.val2"}}
	l.Error("error_message", le2, "key1", "val1")

	want := `
error {"msg":"error_message","error":"the_error","key1":"val1"}
error {"msg":"error_message","error":"the_error","err.key1":"err.val1","key1":"val1"}
error {"msg":"error_message","error":"the_error","err.key1":"err.val1","err.key2":"err.val2","key1":"val1"}
`
	if got := b.String(); strings.TrimSpace(got) != strings.TrimSpace(want) {
		t.Log(got)
		t.Log(want)
		t.Errorf("default filter level: got %q, want %q", got, want)
	}
}
