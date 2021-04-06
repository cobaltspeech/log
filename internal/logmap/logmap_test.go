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

package logmap

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestFromKeyvals(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in  []interface{}
		out MapSlice
	}{
		"empty": {},
		"simple": {
			[]interface{}{"msg", "Hi there."},
			MapSlice{MapItem{"msg", "Hi there."}},
		},
		"missing": {
			[]interface{}{"msg"},
			MapSlice{MapItem{"msg", "missing"}},
		},
		"json": {
			[]interface{}{"msg", "This is a JSONMarshaler.", "item", MapSlice{MapItem{"msg", "Meta!"}}},
			MapSlice{MapItem{"msg", "This is a JSONMarshaler."}, MapItem{"item", MapSlice{MapItem{"msg", "Meta!"}}}},
		},
		"sprintf": {
			[]interface{}{"msg", 4},
			MapSlice{MapItem{"msg", "4"}},
		},
		"text_marshaler": {
			[]interface{}{"msg", newTestMarshaler()},
			MapSlice{MapItem{"msg", newTestMarshaler()}},
		},
	}

	for name, tc := range tests {
		tc := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			exp := FromKeyvals(tc.in...)

			diff := cmp.Diff(tc.out, exp, cmpopts.EquateEmpty())
			if diff != "" {
				t.Errorf("unexpected output (-want +got):\n%s", diff)
			}
		})
	}
}

const nilStr = "<nil>"

func TestMapSlice_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		starting MapSlice
		in       string

		resulting MapSlice
		err       string
	}{
		"empty": {
			in:  "{}",
			err: nilStr,
		},
		"simple": {
			in:        `{"msg":"Hi there."}`,
			resulting: MapSlice{MapItem{"msg", "Hi there."}},
			err:       nilStr,
		},
		"already_present": {
			starting:  MapSlice{MapItem{"msg", "I won't be deleted."}, MapItem{"other", "Me neither."}},
			in:        `{"msg":"Hi there."}`,
			resulting: MapSlice{MapItem{"msg", "I won't be deleted."}, MapItem{"other", "Me neither."}, MapItem{"msg", "Hi there."}},
			err:       nilStr,
		},
		"escape_in_key": {
			in:        `{"\"hi": "Hi there."}`,
			resulting: MapSlice{MapItem{`"hi`, "Hi there."}},
			err:       nilStr,
		},
		"key_appears_in_value": {
			in:        `{"msg": "This \"data\": is for you.", "data": "2"}`,
			resulting: MapSlice{MapItem{"msg", `This "data": is for you.`}, MapItem{"data", "2"}},
			err:       nilStr,
		},
		"bad_json": {
			in:  "this ain't JSON",
			err: "invalid character 'h' in literal true (expecting 'r')",
		},
	}

	for name, tc := range tests {
		tc := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tc.starting.UnmarshalJSON([]byte(tc.in))

			errStr := nilStr
			if err != nil {
				errStr = err.Error()
			}

			diff := cmp.Diff(tc.err, errStr)
			if diff != "" {
				t.Errorf("unexpected error result (-want +got):\n%s", diff)
			}

			if tc.err != nilStr && diff == "" {
				// We don't care about the updates to tc.starting because we wanted an error.
				return
			}

			diff = cmp.Diff(tc.resulting, tc.starting, cmpopts.EquateEmpty())
			if diff != "" {
				t.Errorf("unexpected MapSlice result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMapSlice_String(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		starting MapSlice
		out      string
		err      string
	}{
		"empty": {
			out: "{}\n",
			err: nilStr,
		},
		"simple": {
			starting: MapSlice{MapItem{"msg", "Hello world."}},
			out:      `{"msg":"Hello world."}` + "\n",
			err:      nilStr,
		},
		"text_marshaler": {
			starting: MapSlice{MapItem{"msg", "Hi"}, MapItem{"data", newTestMarshaler()}},
			out:      `{"msg":"Hi","data":"my text: 5"}` + "\n",
			err:      nilStr,
		},
		"encoding_fail": {
			starting: MapSlice{MapItem{"msg", "Hi"}, MapItem{"data", newFailingJSONMarshaler()}},
			err: "json: error calling MarshalJSON for type logmap.MapSlice: json: error calling " +
				"MarshalJSON for type *logmap.failingJSONMarshaler: this error is on purpose",
		},
	}

	for name, tc := range tests {
		tc := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			exp, err := tc.starting.JSONString()

			errStr := nilStr
			if err != nil {
				errStr = err.Error()
			}

			diff := cmp.Diff(tc.err, errStr)
			if diff != "" {
				t.Errorf("unexpected error (-want +got):\n%s", diff)
			}

			if tc.err != nilStr && diff == "" {
				// We don't want to check the output because we got the error we wanted.
				return
			}

			diff = cmp.Diff(tc.out, exp)
			if diff != "" {
				t.Errorf("unexpected output (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMapSlice_ToStringMap(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in  MapSlice
		out map[string]string
	}{
		"empty": {},
		"json": {
			MapSlice{MapItem{"msg", "Hello."}, MapItem{"data", newTestJSONMarshaler()}},
			map[string]string{
				"msg":  "Hello.",
				"data": `{"fancy JSON":6}`,
			},
		},
		"text_marshaler": {
			MapSlice{MapItem{"msg", "Hello."}, MapItem{"data", newTestMarshaler()}},
			map[string]string{
				"msg":  "Hello.",
				"data": `"my text: 5"`,
			},
		},
	}

	for name, tc := range tests {
		tc := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			exp := tc.in.ToStringMap()

			diff := cmp.Diff(tc.out, exp, cmpopts.EquateEmpty())
			if diff != "" {
				t.Errorf("unexpected output (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMapSlice_StringFromValue_panic(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in    interface{}
		panic interface{}
	}{
		"json": {
			newFailingJSONMarshaler(),
			"json: error calling MarshalJSON for type *logmap.failingJSONMarshaler: this error is on purpose",
		},
		"text": {
			newFailingTextMarshaler(),
			"json: error calling MarshalText for type *logmap.failingTextMarshaler: this error is on purpose",
		},
	}

	for name, tc := range tests {
		tc := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()

				if r != nil {
					r = r.(error).Error()
				}

				diff := cmp.Diff(tc.panic, r)
				if diff != "" {
					t.Errorf("unexpected panic message (-want +got):\n%s", diff)
				}
			}()

			_ = StringFromValue(tc.in)
		})
	}
}

// testMarshaler is a type that implements the encoding.TextMarshaler interface, for testing.
type testMarshaler struct {
	A int
}

func newTestMarshaler() *testMarshaler {
	return &testMarshaler{A: 5}
}

func (t *testMarshaler) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("my text: %d", t.A)), nil
}

// testJSONMarshaler is a type that implements the json.JSONMarshaler interface, for testing.
type testJSONMarshaler testMarshaler

func newTestJSONMarshaler() *testJSONMarshaler {
	return &testJSONMarshaler{A: 6}
}

func (t *testJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"fancy JSON": %d}`, t.A)), nil
}

type failingTextMarshaler testMarshaler

func newFailingTextMarshaler() *failingTextMarshaler {
	return &failingTextMarshaler{A: 7}
}

func (f *failingTextMarshaler) MarshalText() ([]byte, error) {
	return nil, errOnPurpose
}

type failingJSONMarshaler testJSONMarshaler

func newFailingJSONMarshaler() *failingJSONMarshaler {
	return &failingJSONMarshaler{A: 8}
}

func (f *failingJSONMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errOnPurpose
}

var errOnPurpose = errors.New("this error is on purpose")
