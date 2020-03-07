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

package logfunc

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/cobaltspeech/log/pkg/level"
)

type invalidValue1 struct{}

func (v *invalidValue1) MarshalText() ([]byte, error) {
	return nil, fmt.Errorf("invalid value")
}

type invalidValue2 struct{}

func (v *invalidValue2) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("invalid value")
}

func TestJSONLogger(t *testing.T) {
	var b bytes.Buffer

	jl := NewJSONLogger(&b, level.Error)
	jl.Log()
	jl.Log("msg")
	jl.Log("msg", "test this")
	jl.Log("msg", &invalidValue1{})
	jl.Log("msg", &invalidValue2{})

	want := []string{
		`{}`,
		`{"msg":"missing"}`,
		`{"msg":"test this"}`,
		`error writing log entry: json: error calling MarshalText for type *logfunc.invalidValue1: invalid value`,
		`error writing log entry: json: error calling MarshalJSON for type *logfunc.invalidValue2: invalid value`,
	}

	scanner := bufio.NewScanner(&b)

	for i := 0; scanner.Scan(); i++ {
		got := strings.SplitN(scanner.Text(), " ", 4)
		if len(got) != 4 {
			t.Errorf("Incorrect output line %d: %q", i, scanner.Text())
			continue
		}

		if got[3] != want[i] {
			t.Errorf("Incorrect output line %d: got %q, want %q", i, got[3], want[i])
		}
	}

}
