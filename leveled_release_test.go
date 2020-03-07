// +build !cobalt_debug

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
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/cobaltspeech/log/pkg/level"
)

func TestLeveledLogger(t *testing.T) {
	var b bytes.Buffer

	l := NewLeveledLogger(&b, level.Trace)

	writelogs := func() {
		l.Trace("msg", "trace message")
		l.Debug("msg", "debug message")
		l.Info("msg", "info message")
		l.Error("msg", "error message")
	}

	writelogs()

	l.SetLevel(level.Error)
	writelogs()

	wantLvl := []string{"debug", "info", "error", "error"}
	wantLines := []string{
		`{"msg":"debug message"}`,
		`{"msg":"info message"}`,
		`{"msg":"error message"}`,
		`{"msg":"error message"}`,
	}

	scanner := bufio.NewScanner(&b)

	for i := 0; scanner.Scan(); i++ {
		// consecutive spaces should be treated as a single space
		txt := strings.Replace(scanner.Text(), "  ", " ", -1)

		got := strings.SplitN(txt, " ", 4)
		if len(got) != 4 {
			t.Errorf("Incorrect output line %d: %q", i, scanner.Text())
			continue
		}

		if got[0] != wantLvl[i] {
			t.Errorf("Incorrect level on line %d. got %s, want %s", i, got[0], wantLvl[i])
		}

		if got[3] != wantLines[i] {
			t.Errorf("Incorrect output line %d: got %q, want %q", i, got[3], wantLines[i])
		}
	}

}
