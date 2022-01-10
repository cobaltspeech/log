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

package testinglog

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cobaltspeech/log/internal/logmap"
	"github.com/cobaltspeech/log/pkg/level"
)

// Logger logs messages to a test runner, and can optionally report differences between log messages
// received and those expected.
type Logger struct {
	// runner has its Log method called in order to report log messages if actualWriter is nil, and
	// its Fail method is called at every unexpected log if doFail is true.
	runner TestRunner

	// The LeveledLogger allows for concurrent logging by wrapping a log.Logger from the stdlib,
	// which in turn uses a sync.Mutex to synchronize writes. We provide similar behavior here.
	mu sync.Mutex

	// truth contains the expected log messages.
	truth         []string
	cur           int
	truthProvided bool
	ignoreOrder   bool

	// doFail is whether a discrepancy in log messages implies a call to runner.Fail.
	doFail bool

	// failed is whether the test has failed.
	failed bool

	// If non-nil, actualFile is where the actual log messages received are written.
	actualFile         *lastMinuteFileWriter
	actualFileOverride bool

	// If non-nil, ignorer is used to choose log message fields whose values should be ignored
	// during comparison.
	ignorer FieldIgnoreFunc
}

// TestRunner is an interface for an object that can receive reports of test failure and logging. It
// is the subset of testing.TB that the Logger requires.
type TestRunner interface {
	Log(args ...interface{})
	Fail()
}

// NewLogger creates a logger that reports all log messages to the provided test runner.
func NewLogger(runner TestRunner, opts ...LoggerOption) (*Logger, error) {
	out := Logger{
		runner: runner,
		doFail: true,
	}

	for _, opt := range opts {
		if err := opt(&out); err != nil {
			return &out, err
		}
	}

	// If the user provided WithActualOutputFile but did not provide WithTruthFile (or the truth
	// file did not exist), we want to write to the actual file even when no failures occur.
	if out.actualFile != nil && (!out.truthProvided || out.actualFileOverride) {
		err := out.actualFile.actuallyWrite()
		if err != nil {
			return &out, err
		}
	}

	// In order for WithIgnoreOrder() to work, we need an actual output writer.
	// If one hasn't been created, we create one that writes only to a memory
	// buffer.
	if out.actualFile == nil && out.ignoreOrder {
		if err := WithActualOutputFile("")(&out); err != nil {
			return &out, err
		}
	}

	return &out, nil
}

// NewConvenientLogger creates a Logger using tb as the TestRunner. Any error while creating the
// logger results in a call to tb.Fatalf. The logger will use testdata/t.Name()/test.log as its
// truth file, and testdata/t.Name()/test.log.generated will contain the actual output when it
// differs from the truth file.
func NewConvenientLogger(tb testing.TB, opts ...LoggerOption) *Logger {
	tb.Helper()

	logFile := filepath.Join("testdata", tb.Name(), "test.log")

	logger, err := NewLogger(tb, append(opts,
		WithTruthFile(logFile),
		WithActualOutputFile(logFile+".generated"),
	)...)
	if err != nil {
		tb.Fatalf("create testing logger: %v", err)
	}

	return logger
}

type LoggerOption func(*Logger) error

// WithTruthFile sets the Logger to compare each log message with the log text
// in the provided file. Each expected log line should be separated from the
// next by a newline. If there are any differences between expected and received
// log lines, they are reported to the test runner and the runner's Fail method
// will be called when Done is called (unless WithoutFailure is used in
// conjunction with this option).
//
// If the logging output order is non-deterministic (e.g. if the function being
// tested uses multiple goroutines), the WithIgnoreOrder() option can be used
// along with this option to handle such cases.
//
//
// If the provided file does not exist, the Logger will not expect any log
// lines.
func WithTruthFile(file string) LoggerOption {
	var lines []string

	var actualFileOverride bool

	truth, err := ioutil.ReadFile(file)

	if os.IsNotExist(err) {
		// We'll pretend it was an empty file, but then we'll be sure to write the actual file if
		// specified, even if it ends up being empty.
		err = nil
		actualFileOverride = true
	} else {
		lines = strings.Split(string(truth), "\n")

		if lines[len(lines)-1] == "" {
			// The file had a final newline or was empty, but we don't want to expect an empty line.
			lines = lines[:len(lines)-1]
		}
	}

	return func(l *Logger) error {
		if l.truthProvided {
			panic("multiple truth sources provided")
		}

		l.truth = lines
		l.truthProvided = true
		l.actualFileOverride = actualFileOverride

		return err
	}
}

// WithActualOutputFile specifies a file path where all actual log messages are written. If a truth
// file has been provided with WithTruthFile, the actual output file is only created if actual log
// messages differ from those in the truth file.
func WithActualOutputFile(file string) LoggerOption {
	return func(l *Logger) error {
		l.actualFile = newLastMinuteFileWriter(file)

		return nil
	}
}

// WithoutFailure sets the Logger to not call Fail() on the test runner even if log messages are
// different than expected.
func WithoutFailure() LoggerOption {
	return func(l *Logger) error {
		l.doFail = false

		return nil
	}
}

// WithIgnoreOrder sets the Logger to not compare each logging call as it is
// made. Instead it will compare the truth file with the actual logging outputs
// at the end when Done() is called. Both the actual logs and truth logs will be
// sorted to ignore any differences in the order the logs were obtained. This is
// useful for working around concurrent logging calls which may happen in a
// different order than in the truth file, but are nonetheless correct.
func WithIgnoreOrder() LoggerOption {
	return func(l *Logger) error {
		l.ignoreOrder = true

		return nil
	}
}

// FieldIgnoreFunc is a function that decides which fields' values should be ignored in a log
// message.
type FieldIgnoreFunc func(fields map[string]string) []string

// WithFieldIgnoreFunc provides the Logger with a function that returns a list of fields to ignore
// in each log message. The keys will still be checked, but not the values. The keys and values in
// the log message are formatted as strings before being provided to the FieldIgnorer.
//
// If ignorer is nil, this option has no effect.
func WithFieldIgnoreFunc(ignorer FieldIgnoreFunc) LoggerOption {
	return func(l *Logger) error {
		l.ignorer = ignorer

		return nil
	}
}

// WithIgnoredFields is a convenience function that uses WithFieldIgnoreFunc to cause the logger to
// ignore the fields in the map corresponding to the "msg" value in each log line. That is, the
// following line creates a LoggerOption that ignores the "devices" field whenever the "msg" field
// is "Devices found.":
//
// 	WithIgnoredFields(map[string][]string{"Devices found.": {"devices"}})
func WithIgnoredFields(ignoreList map[string][]string) LoggerOption {
	return WithFieldIgnoreFunc(func(fields map[string]string) []string {
		msg, ok := fields["msg"]
		if !ok {
			return nil
		}

		for message, list := range ignoreList {
			if msg == message {
				return list
			}
		}

		return nil
	})
}

// Error checks whether the Logger expected an error log line next. If not, it's reported to the
// test runner.
func (l *Logger) Error(keyvals ...interface{}) {
	l.compare(level.Error, keyvals...)
}

// Info checks whether the Logger expected an info log line next. If not, it's reported to the test
// runner.
func (l *Logger) Info(keyvals ...interface{}) {
	l.compare(level.Info, keyvals...)
}

// Debug checks whether the Logger expected a debug log line next. If not, it's reported to the test
// runner.
func (l *Logger) Debug(keyvals ...interface{}) {
	l.compare(level.Debug, keyvals...)
}

// Trace checks whether the Logger expected a trace log line next. If not, it's reported to the test
// runner.
func (l *Logger) Trace(keyvals ...interface{}) {
	l.compare(level.Trace, keyvals...)
}

// log forwards the log message to the actualFile if provided, otherwise to the runner.
func (l *Logger) log(args ...interface{}) {
	if l.actualFile == nil {
		l.runner.Log(args...)

		return
	}

	_, err := fmt.Fprint(l.actualFile, args...)
	if err != nil {
		l.error(fmt.Errorf("error writing to actual file: %w", err))
	}
}

// error reports an error to the test runner and calls runner.Fail immediately.
func (l *Logger) error(err error) {
	l.runner.Log(fmt.Sprintf(`%-5s {"msg":"logging failure","error":%q}`+"\n", level.Error, err))
	l.runner.Fail()
}

// compare checks whether the provided log data were expected, reporting any differences to
// l.runner. It also increments the internal log message counter.
func (l *Logger) compare(lvl level.Level, keyvals ...interface{}) {
	ms := logmap.FromKeyvals(keyvals...)

	exp, err := ms.JSONString()
	if err != nil {
		l.error(fmt.Errorf("error creating log string: %w", err))
	}

	exp = fmt.Sprintf("%-5s %s", lvl, exp)

	l.mu.Lock()
	defer l.mu.Unlock()

	// Log it to either the logWriter or the runner.
	l.log(exp)

	if !l.truthProvided || l.ignoreOrder {
		return
	}

	hyp := l.getCurrent()

	// exp comes with a newline at the end so we remove that for comparison.
	exp = exp[:len(exp)-1]

	// Check if the log lines are equivalent.
	if !l.cmp(hyp, exp, lvl, ms) {
		// Log the failure and the diff to the runner.
		l.runner.Log("unexpected log message (-want +got):\n", replaceNbsp(cmp.Diff(hyp, exp)))

		l.failed = true

		if l.actualFile != nil {
			// We're going to fail, so we can start writing the actual file.
			err = l.actualFile.actuallyWrite()
			if err != nil {
				l.error(err)
			}
		}
	}

	l.cur++
}

// compareFinalSortedLogs checks the given truth log file with the actual log
// file and reports any differences to l.runner. Both truth and actual logs are
// sorted to ignore any differences in the order the logs were obtained. This is
// useful for working around concurrent logging calls which may happen in a
// different order than in the truth file, but are nonetheless correct.
func (l *Logger) compareFinalSortedLogs() {
	// Return if missing truth logs.
	if !l.truthProvided {
		return
	}

	// Making a copy of truth logs since we sort them below.
	truthLogs := make([]string, len(l.truth))
	copy(truthLogs, l.truth)

	// Getting actual logs and filtering a final newline.
	actualLogs := strings.Split(l.actualFile.b.String(), "\n")
	if actualLogs[len(actualLogs)-1] == "" {
		actualLogs = actualLogs[:len(actualLogs)-1]
	}

	sort.Strings(truthLogs)
	sort.Strings(actualLogs)

	// We have missing or extra log lines. Fail and print diff
	if len(truthLogs) != len(actualLogs) {
		diff := replaceNbsp(cmp.Diff(truthLogs, actualLogs))
		l.runner.Log("unexpected number of log messages (-want +got):\n", diff)
		l.failed = true

		return
	}

	// Check if the log lines are equivalent.
	for i := 0; i < len(actualLogs); i++ {
		want := truthLogs[i]
		got := actualLogs[i]

		// Getting level for the actual log message.
		gotLvl := level.FromString(got[:6])

		// Getting key-val pairs in log message. Skipping level info.
		var gotMap logmap.MapSlice
		if err := gotMap.UnmarshalJSON([]byte(got[6:])); err != nil {
			panic(err)
		}

		if !l.cmp(want, got, gotLvl, gotMap) {
			// Log the failure and the diff to the runner.
			diff := replaceNbsp(cmp.Diff(want, got))
			l.runner.Log("unexpected log message (-want +got):\n", diff)
			l.failed = true
		}
	}
}

func (l *Logger) getCurrent() string {
	if l.cur < len(l.truth) {
		return l.truth[l.cur]
	}

	// We've moved beyond the list of truth log messages we received.
	return ""
}

// cmp compares the log lines using == or by checking each field individually if the ignorer is
// non-nil.
func (l *Logger) cmp(hyp, exp string, expLvl level.Level, expMap logmap.MapSlice) bool {
	if l.ignorer == nil {
		return hyp == exp
	}

	if hyp == "" {
		// We weren't expecting a log message, so this log line is definitely wrong.
		return false
	}

	var hypMap logmap.MapSlice

	err := hypMap.UnmarshalJSON([]byte(hyp[6:])) // Skip the log level.
	if err != nil {
		panic(err)
	}

	// If there aren't the same number of fields, there's no point in calling the ignorer.
	if len(hypMap) != len(expMap) {
		return false
	}

	// If they're not the same log level, there's no point in calling the ignorer.
	if !strings.HasPrefix(hyp, expLvl.String()) {
		return false
	}

	keysToIgnore := l.ignorer(hypMap.ToStringMap())

	for i := range hypMap {
		if hypMap[i].Key != expMap[i].Key {
			return false
		}

		if sliceContains(keysToIgnore, hypMap[i].Key) {
			// We're not going to compare the values.
			continue
		}

		// Values in hypMap were unmarshaled from JSON, so they're strings. This is not necessarily
		// the case with values in expMap, so we need to convert them to strings.
		if hypMap[i].Value != logmap.StringFromValue(expMap[i].Value) {
			return false
		}
	}

	return true
}

func sliceContains(slice []string, item string) bool {
	for i := range slice {
		if slice[i] == item {
			return true
		}
	}

	return false
}

// replaceNbsp replaces the non-breaking space character (unicode character 0xA0) with a regular
// space (unicode character 0x20) in certain places where cmp.Diff seems to randomly place them.
func replaceNbsp(s string) string {
	// We also need to catch NBSPs at the beginning of the first line.
	if len(s) >= 4 && s[:4] == "\u00a0\u00a0" {
		// NBSP is two chars long.
		s = "  " + s[4:]
	}

	return nbspReplacer.Replace(s)
}

var nbspReplacer = strings.NewReplacer(
	"\n\u00a0\u00a0", "\n  ", // cases like "  string("
	"\n-\u00a0", "\n- ", // cases like "- <value>"
	"\n+\u00a0", "\n+ ", // cases like "+ <value>"
)

// Done signals to the Logger that no more log methods will be called. It checks whether the Logger
// expected more log messages, and calls runner.Fail if any differences were reported.
func (l *Logger) Done() {
	if !l.doFail {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.ignoreOrder {
		l.compareFinalSortedLogs()
	} else if l.cur < len(l.truth) {
		// We're missing some log messages that we expected.
		l.failed = true
		for ; l.cur < len(l.truth); l.cur++ {
			l.runner.Log("missing log message (-want +got):\n", replaceNbsp(cmp.Diff(l.truth[l.cur], "")))
		}
	}

	if l.failed {
		if l.actualFile != nil {
			err := l.actualFile.actuallyWrite()
			if err != nil {
				l.error(fmt.Errorf("error writing to actual file: %w", err))
			}
		}

		l.runner.Fail()
	}

	if l.actualFile != nil {
		err := l.actualFile.Close()
		if err != nil {
			l.error(err)
		}
	}
}
