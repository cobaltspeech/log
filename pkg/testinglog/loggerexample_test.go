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
	"bytes"
	"fmt"
)

// fakeRunner implements TestRunner, for testing the Logger.
type fakeRunner struct {
	b      bytes.Buffer
	failed bool
}

func (r *fakeRunner) Fail() {
	r.failed = true
}

func (r *fakeRunner) Log(args ...interface{}) {
	fmt.Fprint(&r.b, args...)
}

func ExampleLogger() {
	runner := &fakeRunner{}
	logger, _ := NewLogger(runner)

	logger.Error("msg", "There was a problem.", "data", 3.14)
	logger.Debug("msg", "Here's some pertinent information.", "numCalls", 17)

	logger.Done()

	fmt.Println(runner.b.String())
	// Output:
	// error {"msg":"There was a problem.","data":"3.14"}
	// debug {"msg":"Here's some pertinent information.","numCalls":"17"}
}
