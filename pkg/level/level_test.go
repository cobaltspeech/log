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

// Package level defines the supported logging levels
package level

import "testing"

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{Trace, "trace"},
		{Debug, "debug"},
		{Info, "info"},
		{Error, "error"},
	}

	for _, tc := range tests {
		if got := tc.level.String(); got != tc.want {
			t.Errorf("Level.String(%d) = %s; want %s", tc.level, got, tc.want)
		}
	}
}

func TestLevel_Verbosity(t *testing.T) {
	tests := []struct {
		verbosity int
		want      Level
	}{
		{-1, Error},
		{0, Error},
		{1, Error | Info},
		{2, Error | Info | Debug},
		{3, All},
		{4, All},
	}

	for _, tc := range tests {
		if got := Verbosity(tc.verbosity); got != tc.want {
			t.Errorf("incorrect level mapping for verbosity=%d", tc.verbosity)
		}
	}
}

func TestLevel_FromString(t *testing.T) {
	tests := []struct {
		str  string
		want Level
	}{
		{Trace.String(), Trace},
		{Debug.String(), Debug},
		{Info.String(), Info},
		{Error.String(), Error},
		{Default.String(), Default},
		{All.String(), All},
		{None.String(), None},
		{"unknown ", None},
		{"info  ", Info},
		{"  info  ", Info},
		{"INFO", Info},
		{"info level", None},
	}

	for _, tc := range tests {
		if got := FromString(tc.str); got != tc.want {
			t.Errorf("incorrect level mapping for string=%q", tc.str)
		}
	}
}
