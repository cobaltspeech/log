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
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cobaltspeech/log/pkg/level"
)

// writeTemporaryFile is used by some of the examples to provide a short way to create a file
// containing a specific string and later delete it. The returned function need only be called if
// err == nil.
func writeTemporaryFile(data string) (filename string, remove func(), err error) {
	var f *os.File

	f, err = ioutil.TempFile("", "")
	if err != nil {
		return "", func() {}, err
	}

	_, err = f.WriteString(data)
	if err != nil {
		return "", func() {}, err
	}

	filename = f.Name()
	remove = func() {
		e := os.Remove(filename)
		if e != nil {
			fmt.Println(e)
		}
	}

	err = f.Close()

	if err != nil {
		remove()
	}

	return filename, remove, err
}

func ExampleWithTruthFile() {
	// Write an example file.
	hypFile, remove, err := writeTemporaryFile(strings.Join([]string{
		`error {"msg":"There was a problem.","data":"3.14"}`,
		`debug {"msg":"Here's the number of calls.","numCalls":"17"}`,
	}, "\n"))
	if err != nil {
		fmt.Println(err)

		return
	}

	defer remove()

	runner := fakeRunner{}

	logger, err := NewLogger(&runner, WithTruthFile(hypFile))
	if err != nil {
		fmt.Println(err)

		return
	}

	logger.Error("msg", "There was a problem.", "data", 3.14)
	logger.Debug("msg", "Here's some pertinent information.", "numCalls", 17) // doesn't match hyp
	logger.Trace("msg", "This trace message shouldn't be here.")

	logger.Done()

	fmt.Print(runner.b.String())
	fmt.Println(runner.failed)
	// Output:
	// error {"msg":"There was a problem.","data":"3.14"}
	// debug {"msg":"Here's some pertinent information.","numCalls":"17"}
	// unexpected log message (-want +got):
	//   string(
	// - 	`debug {"msg":"Here's the number of calls.","numCalls":"17"}`,
	// + 	`debug {"msg":"Here's some pertinent information.","numCalls":"17"}`,
	//   )
	// trace {"msg":"This trace message shouldn't be here."}
	// unexpected log message (-want +got):
	//   string(
	// - 	"",
	// + 	`trace {"msg":"This trace message shouldn't be here."}`,
	//   )
	// true
}

func ExampleWithActualOutputFile() {
	hypFile, remove, err := writeTemporaryFile(strings.Join([]string{
		`error {"msg":"There was a problem.","data":"3.14"}`,
		`debug {"msg":"Here's the number of calls.","numCalls":"18"}`,
	}, "\n"))
	if err != nil {
		fmt.Println(err)

		return
	}

	defer remove()

	// Get a file we can use for the actual log output.
	var actualFile string
	actualFile, remove, err = writeTemporaryFile("")

	if err != nil {
		fmt.Println(err)

		return
	}

	defer remove()

	runner := fakeRunner{}

	logger, err := NewLogger(&runner, WithTruthFile(hypFile), WithActualOutputFile(actualFile))
	if err != nil {
		fmt.Println(err)

		return
	}

	logger.Error("msg", "There was a problem.", "data", 3.14)
	logger.Debug("msg", "Here's some pertinent information.", "numCalls", 18) // doesn't match hyp
	logger.Trace("msg", "This trace message shouldn't be here.")

	logger.Done()

	// Get the actual log output.
	var actualBytes []byte
	actualBytes, err = ioutil.ReadFile(actualFile)

	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Print(runner.b.String())
	fmt.Print(string(actualBytes))
	fmt.Println(runner.failed)
	// Output:
	// unexpected log message (-want +got):
	//   string(
	// - 	`debug {"msg":"Here's the number of calls.","numCalls":"18"}`,
	// + 	`debug {"msg":"Here's some pertinent information.","numCalls":"18"}`,
	//   )
	// unexpected log message (-want +got):
	//   string(
	// - 	"",
	// + 	`trace {"msg":"This trace message shouldn't be here."}`,
	//   )
	// error {"msg":"There was a problem.","data":"3.14"}
	// debug {"msg":"Here's some pertinent information.","numCalls":"18"}
	// trace {"msg":"This trace message shouldn't be here."}
	// true
}

func ExampleWithoutFailure() {
	hypFile, remove, err := writeTemporaryFile(strings.Join([]string{
		`error {"msg":"There was a problem.","data":"3.14"}`,
		`debug {"msg":"Here's the number of calls.","numCalls":"19"}`,
	}, "\n"))
	if err != nil {
		fmt.Println(err)

		return
	}

	defer remove()

	runner := fakeRunner{}

	logger, err := NewLogger(&runner, WithTruthFile(hypFile), WithoutFailure())
	if err != nil {
		fmt.Println(err)

		return
	}

	logger.Error("msg", "There was a problem.", "data", 3.14)

	// This log message doesn't match, but we still want the test to pass.
	logger.Debug("msg", "Here's some pertinent information.", "numCalls", 19)

	logger.Done()

	fmt.Println(runner.failed)
	// Output:
	// false
}

func ExampleWithFieldIgnoreFunc() {
	hypFile, remove, err := writeTemporaryFile(strings.Join([]string{
		`trace {"msg":"An ID was generated.","id":"brh634381n1ts5ibr1eg"}`,
		`debug {"msg":"This ID is deterministic.","id":"42"}`,
	}, "\n"))
	if err != nil {
		fmt.Println(err)

		return
	}

	defer remove()

	runner := fakeRunner{}

	ignorer := func(fields map[string]string) []string {
		msg, ok := fields["msg"]

		// Ignore only the "id" field in the non-deterministic ID log line.
		if ok && msg == "An ID was generated." {
			return []string{"id"}
		}

		return nil
	}

	logger, err := NewLogger(&runner, WithTruthFile(hypFile), WithFieldIgnoreFunc(ignorer))
	if err != nil {
		fmt.Println(err)

		return
	}

	// This is not deterministic.
	id := strconv.Itoa(rand.Intn(10000000)) //nolint: gosec // We don't need security for this test.

	logger.Trace("msg", "An ID was generated.", "id", id)
	logger.Debug("msg", "This ID is deterministic.", "id", 12) // doesn't match hyp
	logger.Error("msg", "This is unexpected.")                 // extra

	logger.Done()

	fmt.Print(strings.Replace(runner.b.String(), id, "<id removed>", 1))
	// Output:
	// trace {"msg":"An ID was generated.","id":"<id removed>"}
	// debug {"msg":"This ID is deterministic.","id":"12"}
	// unexpected log message (-want +got):
	//   string(
	// - 	`debug {"msg":"This ID is deterministic.","id":"42"}`,
	// + 	`debug {"msg":"This ID is deterministic.","id":"12"}`,
	//   )
	// error {"msg":"This is unexpected."}
	// unexpected log message (-want +got):
	//   string(
	// - 	"",
	// + 	`error {"msg":"This is unexpected."}`,
	//   )
}

type testingLogMsg struct {
	lvl     level.Level
	keyvals []interface{}
}

func TestWithTruthFile(t *testing.T) { // nolint: funlen // Tests are just long.
	tests := map[string]struct {
		// in is the set of log messages that will actually be given to the Logger.
		in []testingLogMsg

		// expect is the name of a file in "testdata" containing the log messages that the Logger
		// should expect.
		expect string

		// hyp is the complete string that we should expect to be written to the test runner given
		// `expect` and `in`.
		hyp string

		// expectFail is whether to expect the Logger to call runner.Fail.
		expectFail bool
	}{
		"empty": {
			expect:     "empty.log",
			expectFail: false,
		},
		"correct": {
			in: []testingLogMsg{
				{level.Trace, []interface{}{"msg", "This is just a trace.", "data", 3.14}},
				{level.Debug, []interface{}{"msg", "This is a debug msg.", "data", []byte{0, 1, 2, 3}}},
				{level.Info, []interface{}{"msg", "This msg might be useful.", "data", 12}},
				{level.Error, []interface{}{"msg", "There's a problem.", "data"}}, // missing value
			},
			expect: "correct.log",
			hyp: strings.Join([]string{
				`trace {"msg":"This is just a trace.","data":"3.14"}`,
				`debug {"msg":"This is a debug msg.","data":"[0 1 2 3]"}`,
				`info  {"msg":"This msg might be useful.","data":"12"}`,
				`error {"msg":"There's a problem.","data":"missing"}`,
			}, "\n"),
			expectFail: false,
		},
		"bad_order": {
			in: []testingLogMsg{
				{level.Trace, []interface{}{"msg", "This is just a trace.", "data", 3.14}},
				{level.Info, []interface{}{"msg", "This msg might be useful.", "data", 12}},
				{level.Debug, []interface{}{"msg", "This is a debug msg.", "data", []byte{0, 1, 2, 3}}},
			},
			expect: "bad_order.log",
			hyp: strings.Join([]string{
				`trace {"msg":"This is just a trace.","data":"3.14"}`,
				`info  {"msg":"This msg might be useful.","data":"12"}`,
				"unexpected log message (-want +got):",
				"  string(",
				"- 	`" + `debug {"msg":"This is a debug msg.","data":"[0 1 2 3]"}` + "`,",
				"+ 	`" + `info  {"msg":"This msg might be useful.","data":"12"}` + "`,",
				"  )",
				`debug {"msg":"This is a debug msg.","data":"[0 1 2 3]"}`,
				"unexpected log message (-want +got):",
				"  string(",
				"- 	`" + `info  {"msg":"This msg might be useful.","data":"12"}` + "`,",
				"+ 	`" + `debug {"msg":"This is a debug msg.","data":"[0 1 2 3]"}` + "`,",
				"  )",
			}, "\n"),
			expectFail: true,
		},
		"missing_message": {
			expect: "missing_message.log",
			hyp: strings.Join([]string{
				"missing log message (-want +got):",
				"  string(",
				"- 	`" + `trace {"msg":"This message should be here."}` + "`,",
				`+ 	"",`,
				"  )",
			}, "\n"),
			expectFail: true,
		},
		"extra_message": {
			in: []testingLogMsg{
				{level.Trace, []interface{}{"msg", "This message should not be here."}},
			},
			expect: "empty.log",
			hyp: strings.Join([]string{
				`trace {"msg":"This message should not be here."}`,
				"unexpected log message (-want +got):",
				"  string(",
				`- 	"",`,
				"+ 	`" + `trace {"msg":"This message should not be here."}` + "`,",
				"  )",
			}, "\n"),
			expectFail: true,
		},
		"exotic_types": {
			in: []testingLogMsg{
				{level.Debug, []interface{}{"msg", "Some data.", "data", struct{}{}}},
			},
			expect: "empty.log",
			hyp: strings.Join([]string{
				`debug {"msg":"Some data.","data":"{}"}`,
				"unexpected log message (-want +got):",
				"  string(",
				`- 	"",`,
				"+ 	`" + `debug {"msg":"Some data.","data":"{}"}` + "`,",
				"  )",
			}, "\n"),
			expectFail: true,
		},
	}

	for name, tc := range tests {
		tc := tc

		t.Run(name, func(t *testing.T) {
			runner := fakeRunner{}
			logger, err := NewLogger(&runner, WithTruthFile(filepath.Join("testdata", tc.expect)))
			if err != nil {
				t.Fatal(err)
			}

			writeLogMsgs(logger, tc.in)

			// The logger shouldn't call Fail until Done is called.
			if runner.failed {
				t.Errorf("logger called Fail before Done was called")
			}

			logger.Done()

			runner.compareOutput(t, tc.hyp, tc.expectFail)
		})
	}
}

func writeLogMsgs(logger *Logger, msgs []testingLogMsg) {
	for _, msg := range msgs {
		switch msg.lvl {
		case level.Error:
			logger.Error(msg.keyvals...)
		case level.Info:
			logger.Info(msg.keyvals...)
		case level.Debug:
			logger.Debug(msg.keyvals...)
		case level.Trace:
			logger.Trace(msg.keyvals...)
		default:
			panic(fmt.Sprintf("test log message had unexpected level: %v", msg.lvl))
		}
	}
}

func (r *fakeRunner) compareOutput(t *testing.T, hyp string, expectFail bool) {
	t.Helper()

	exp := r.b.String()
	expLines := strings.Split(exp, "\n")
	hypLines := strings.Split(hyp, "\n")

	for len(expLines) < len(hypLines) {
		expLines = append(expLines, "")
	}

	for len(hypLines) < len(expLines) {
		hypLines = append(hypLines, "")
	}

	for i, hypLine := range hypLines {
		diff := cmp.Diff(hypLine, expLines[i])
		if diff != "" {
			t.Errorf("runner log mismatch on line %d (-want +got):\n%s", i+1, diff)
		}
	}

	if t.Failed() {
		t.Log(exp)
	}

	diff := cmp.Diff(expectFail, r.failed)
	if diff != "" {
		t.Errorf("failure mismatch (-want +got):\n%s", diff)
	}
}

func TestWithTruthFile_concurrent(t *testing.T) {
	runner := fakeRunner{}

	logger, err := NewLogger(&runner, WithTruthFile(filepath.Join("testdata", "concurrent.log")))
	if err != nil {
		t.Fatalf("create logger: %v", err)
	}

	var wg sync.WaitGroup

	wg.Add(3)

	for i := 0; i < 3; i++ {
		go func() {
			logger.Info("msg", "This is a concurrent message.")
			wg.Done()
		}()
	}

	wg.Wait()
	logger.Done()

	runner.compareOutput(t, `info  {"msg":"This is a concurrent message."}
info  {"msg":"This is a concurrent message."}
info  {"msg":"This is a concurrent message."}
`, false)
}

func TestWithTruthFile_panic(t *testing.T) {
	defer func() {
		r := recover()

		diff := cmp.Diff("multiple truth sources provided", r)
		if diff != "" {
			t.Errorf("panic value different than expected:\n%s", diff)
		}
	}()

	_, _ = NewLogger(
		nil,
		WithTruthFile(filepath.Join("testdata", "empty.log")),
		WithTruthFile(filepath.Join("testdata", "correct.log")),
	)
}

const nilStr = "<nil>"

func TestWithTruthFile_noexist(t *testing.T) {
	truthFile := filepath.Join("testdata", "nonexistent.log")
	logger, err := NewLogger(nil, WithTruthFile(truthFile), WithActualOutputFile(truthFile+".generated"))

	errStr := nilStr
	if err != nil {
		errStr = err.Error()
	}

	diff := cmp.Diff(nilStr, errStr)
	if diff != "" {
		t.Errorf("err differed from expected (-want +got):\n%s", diff)
	}

	logger.Done()

	b, err := ioutil.ReadFile(truthFile + ".generated")
	if err != nil {
		t.Errorf("error trying to read generated file: %v", err)
	}

	diff = cmp.Diff("", string(b))
	if diff != "" {
		t.Errorf("generated file differed from expected (-want +got):\n%s", diff)
	}

	err = os.Remove(truthFile + ".generated")
	if err != nil {
		t.Fatalf("failed to remove generated file: %v", err)
	}
}

func TestWithActualOutputFile_noTruth(t *testing.T) {
	// Get a file we can use for the actual log output.
	actualFile, remove, err := writeTemporaryFile("")
	if err != nil {
		t.Fatal(err)
	}

	defer remove()

	runner := fakeRunner{}

	logger, err := NewLogger(&runner, WithActualOutputFile(actualFile))
	if err != nil {
		t.Fatal(err)
	}

	logger.Error("msg", "This is an error we expect to see!")
	logger.Trace("msg", "This is a trace message.")
	logger.Done()

	// We don't expect anything to have happened to the test runner.
	runnerExpect := ""
	expectFail := false
	runner.compareOutput(t, runnerExpect, expectFail)

	// Check the file and make sure it's what we expect.
	hyp := strings.Join([]string{
		`error {"msg":"This is an error we expect to see!"}`,
		`trace {"msg":"This is a trace message."}`,
		``, // We expect a final newline.
	}, "\n")

	b, err := ioutil.ReadFile(actualFile)
	if err != nil {
		t.Fatal(err)
	}

	diff := cmp.Diff(hyp, string(b))
	if diff != "" {
		t.Errorf("unexpected output to actual file (-want, +got):\n%s", diff)
	}
}

func TestWithActualOutputFile_withTruthMatch(t *testing.T) {
	// Make up a filename for the actual output. Since the logs will match, we should expect this
	// file not to be created.
	actualFile := filepath.Join("testdata", "this_should_never_exist.log.generated")

	runner := fakeRunner{}

	logger, err := NewLogger(&runner,
		WithActualOutputFile(actualFile),
		WithTruthFile(filepath.Join("testdata", "missing_message.log")),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Send the right log message.
	logger.Trace("msg", "This message should be here.")
	logger.Done()

	// The runner should have been told nothing.
	runnerExpect := ""
	expectFail := false
	runner.compareOutput(t, runnerExpect, expectFail)

	// The actual output file shouldn't have been created.
	if _, err := os.Stat(actualFile); err == nil {
		t.Error("actual file was created but should not have been")
	}
}

func TestWithActualOutputFile_withTruthNoMatch(t *testing.T) {
	// Get a file we can use for the actual log output.
	actualFile, remove, err := writeTemporaryFile("")
	if err != nil {
		t.Fatal(err)
	}

	defer remove()

	runner := fakeRunner{}

	logger, err := NewLogger(&runner,
		WithActualOutputFile(actualFile),
		WithTruthFile(filepath.Join("testdata", "missing_message.log")),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Send no log messages.
	logger.Done()

	// The runner should have been told about the missing message.
	runnerExpect := strings.Join([]string{
		"missing log message (-want +got):",
		"  string(",
		"- 	`" + `trace {"msg":"This message should be here."}` + "`,",
		`+ 	"",`,
		"  )",
	}, "\n")
	expectFail := true
	runner.compareOutput(t, runnerExpect, expectFail)

	// The actual output should be nothing.
	b, err := ioutil.ReadFile(actualFile)
	if err != nil {
		t.Fatal(err)
	}

	diff := cmp.Diff("", string(b))
	if diff != "" {
		t.Errorf("unexpected output to actual file (-want, +got):\n%s", diff)
	}
}

func TestWithActualOutputFile_error(t *testing.T) {
	// Get a directory we can use as the path for writing the actual file, so it errors.
	actualFile, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		e := os.RemoveAll(actualFile)
		if e != nil {
			t.Fatal(e)
		}
	}()

	runner := fakeRunner{}

	_, err = NewLogger(&runner, WithActualOutputFile(actualFile))

	errStr := nilStr
	if err != nil {
		errStr = err.Error()
	}

	hypErr := fmt.Sprintf("open %s: is a directory", actualFile)

	diff := cmp.Diff(hypErr, errStr)
	if diff != "" {
		t.Errorf("unexpected error (-want +got):\n%s", diff)
	}
}

func TestWithActualOutputFile_errorOnDoneWithTruth(t *testing.T) {
	// Get a directory we can use as the path for writing the actual file, so it errors.
	actualFile, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		e := os.RemoveAll(actualFile)
		if e != nil {
			t.Fatal(e)
		}
	}()

	runner := fakeRunner{}

	logger, err := NewLogger(&runner,
		WithActualOutputFile(actualFile),
		WithTruthFile(filepath.Join("testdata", "two.log")),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Forget to send the last log, but do send the first one so we have to write the actual.
	logger.Info("msg", "This msg might be useful.", "data", 12)
	logger.Done()

	// We should have received logging errors from the call to Done.
	runnerExpect := strings.Join([]string{
		"missing log message (-want +got):",
		"  string(",
		"- 	`" + `trace {"msg":"This message should be here."}` + "`,",
		`+ 	"",`,
		"  )",
		fmt.Sprintf(
			`error {"msg":"logging failure","error":"error writing to actual file: open %s: is a directory"}`+"\n",
			actualFile,
		),
	}, "\n")
	expectFail := true
	runner.compareOutput(t, runnerExpect, expectFail)
}

func TestWithFieldIgnorer(t *testing.T) { // nolint: funlen // Tests are just long.
	t.Parallel()

	dataIgnorer := func(fields map[string]string) []string {
		msg, ok := fields["msg"]

		if ok && msg == "This message should be here." {
			return []string{"data"}
		}

		return nil
	}

	tests := map[string]struct {
		ignorer    FieldIgnoreFunc
		in         []testingLogMsg
		expect     string
		hyp        string
		expectFail bool
	}{
		"extra_key_not_ok": {
			ignorer: dataIgnorer,
			in: []testingLogMsg{
				{level.Info, []interface{}{"msg", "This msg might be useful.", "data", "13"}}, // wrong value
				{level.Trace, []interface{}{"msg", "This message should be here.", "data"}},   // extra key
			},
			expect: "two.log",
			hyp: strings.Join([]string{
				`info  {"msg":"This msg might be useful.","data":"13"}`,
				"unexpected log message (-want +got):",
				"  string(",
				"- 	`" + `info  {"msg":"This msg might be useful.","data":"12"}` + "`,",
				"+ 	`" + `info  {"msg":"This msg might be useful.","data":"13"}` + "`,",
				"  )",
				`trace {"msg":"This message should be here.","data":"missing"}`,
				"unexpected log message (-want +got):",
				"  string(",
				"- 	`" + `trace {"msg":"This message should be here."}` + "`,",
				"+ 	`" + `trace {"msg":"This message should be here.","data":"missing"}` + "`,",
				"  )",
			}, "\n"),
			expectFail: true,
		},
		"wrong_keys": {
			ignorer: dataIgnorer,
			in: []testingLogMsg{
				{level.Info, []interface{}{"message", "This msg might be useful.", "data", "12"}}, // wrong key
				{level.Trace, []interface{}{"msg", "This message should be here."}},
			},
			expect: "two.log",
			hyp: strings.Join([]string{
				`info  {"message":"This msg might be useful.","data":"12"}`,
				"unexpected log message (-want +got):",
				"  string(",
				"- 	`" + `info  {"msg":"This msg might be useful.","data":"12"}` + "`,",
				"+ 	`" + `info  {"message":"This msg might be useful.","data":"12"}` + "`,",
				"  )",
				`trace {"msg":"This message should be here."}`,
			}, "\n"),
			expectFail: true,
		},
		"wrong_level": {
			ignorer: dataIgnorer,
			in: []testingLogMsg{
				{level.Error, []interface{}{"msg", "This msg might be useful.", "data", "12"}}, // wrong level
				{level.Trace, []interface{}{"msg", "This message should be here."}},
			},
			expect: "two.log",
			hyp: strings.Join([]string{
				`error {"msg":"This msg might be useful.","data":"12"}`,
				"unexpected log message (-want +got):",
				"  string(",
				"- 	`" + `info  {"msg":"This msg might be useful.","data":"12"}` + "`,",
				"+ 	`" + `error {"msg":"This msg might be useful.","data":"12"}` + "`,",
				"  )",
				`trace {"msg":"This message should be here."}`,
			}, "\n"),
			expectFail: true,
		},
	}

	for name, tc := range tests {
		tc := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			runner := fakeRunner{}
			logger, err := NewLogger(
				&runner,
				WithTruthFile(filepath.Join("testdata", tc.expect)),
				WithFieldIgnoreFunc(tc.ignorer),
			)
			if err != nil {
				t.Fatal(err)
			}

			writeLogMsgs(logger, tc.in)
			logger.Done()

			runner.compareOutput(t, tc.hyp, tc.expectFail)
		})
	}
}

// TestWithIgnoreOrder checks whether tests:
// 	- can pass when WithIgnoreOrder option is enabled and logs are not received
//    in the same order as in the truth file.
//
// 	- fail when WithIgnoreOrder option is not enabled and logs are not received
//    in the same order as in the truth file.
//
// 	- can pass when WithIgnoreOrder option is enabled and but no truth file is
//    provided.
func TestWithIgnoreOrder(t *testing.T) {
	t.Parallel()

	truthLogs := []testingLogMsg{
		{level.Trace, []interface{}{"msg", "trace message", "order", "0"}},
		{level.Debug, []interface{}{"msg", "debug message", "order", "1"}},
		{level.Info, []interface{}{"msg", "info message", "order", "2"}},
		{level.Error, []interface{}{"msg", "error message", "order", "3"}},
	}

	truthLogsStr := []string{
		`trace {"msg":"trace message","order":"0"}`,
		`debug {"msg":"debug message","order":"1"}`,
		`info  {"msg":"info message","order":"2"}`,
		`error {"msg":"error message","order":"3"}`,
	}

	// Write truth file.
	truthFile, remove, err := writeTemporaryFile(strings.Join(truthLogsStr, "\n"))
	if err != nil {
		t.Fatalf("failed to create temporary truth file, err=%v", err)
	}

	t.Cleanup(remove)

	testCases := []struct {
		name              string
		order             []int
		overrideLog       map[int]testingLogMsg
		passOrdered       bool
		passIgnoringOrder bool
	}{
		{
			name: "original order", order: []int{0, 1, 2, 3}, passOrdered: true, passIgnoringOrder: true,
		},
		{
			name: "shift order by 1", order: []int{1, 2, 3, 0}, passIgnoringOrder: true,
		},
		{
			name: "shift order by 2", order: []int{2, 3, 0, 1}, passIgnoringOrder: true,
		},
		{
			name: "shift order by 3", order: []int{3, 0, 1, 2}, passIgnoringOrder: true,
		},
		{
			name: "missing logging message", order: []int{0, 1, 2},
		},
		{
			name: "shift order by 2 and missing logging message", order: []int{2, 3, 0},
		},
		{
			name:        "extra logging message",
			order:       []int{0, 1, 2, 3},
			overrideLog: map[int]testingLogMsg{4: {level.Info, []interface{}{"msg", "extra message", "order", "4"}}},
		},
		{
			name:        "different logging message",
			order:       []int{0, 1, 2, 3},
			overrideLog: map[int]testingLogMsg{0: {level.Info, []interface{}{"msg", "trace message", "order", "4"}}},
		},
	}

	loggerOptsToTest := map[string][]LoggerOption{
		"with no truth file":      {},
		"with not ignoring order": {WithTruthFile(truthFile)},
		"with ignoring order":     {WithTruthFile(truthFile), WithIgnoreOrder()},
	}

	for _, tc := range testCases {
		for optName := range loggerOptsToTest {
			opts := loggerOptsToTest[optName]

			t.Run(tc.name+"_"+optName, func(subT *testing.T) {
				subT.Parallel()

				msgs := make([]testingLogMsg, 0)
				for _, ldx := range tc.order {
					msgs = append(msgs, truthLogs[ldx])
				}

				for ldx, override := range tc.overrideLog {
					if ldx < len(msgs) {
						msgs[ldx] = override
					} else {
						msgs = append(msgs, override)
					}
				}

				subTestWithIgnoreOrder(subT, opts, msgs, tc.passIgnoringOrder, tc.passOrdered)
			})
		}
	}
}

func subTestWithIgnoreOrder(
	t *testing.T, opts []LoggerOption, msgs []testingLogMsg, passIgnoringOrder, passOrdered bool,
) {
	t.Helper()

	runner := fakeRunner{}

	logger, err := NewLogger(&runner, opts...)
	if err != nil {
		t.Fatalf("failed to create logger using WithIgnoreOrder option, err=%v", err)
	}

	writeLogMsgs(logger, msgs)

	logger.Done()

	shouldPass := !logger.truthProvided || (logger.ignoreOrder && passIgnoringOrder) || (!logger.ignoreOrder && passOrdered)

	if !shouldPass && !runner.failed {
		t.Errorf(
			"expected to fail but didn't, diff:\n%s", runner.b.String(),
		)
	}

	if shouldPass && runner.failed {
		t.Errorf(
			"expected to pass  but didn't, diff:\n%s", runner.b.String(),
		)
	}
}
