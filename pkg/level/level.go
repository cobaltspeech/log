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

import "strings"

// Level enumerates different logging levels.
type Level byte

const (
	Trace Level = 1 << iota
	Debug
	Info
	Error
)

const (
	None    Level = 0
	Default Level = Info | Error
	All     Level = Trace | Debug | Info | Error
)

// levelCodes provides a string representation of different supported levels.
var levelCodes = map[Level]string{
	Trace:   "trace",
	Debug:   "debug",
	Info:    "info",
	Error:   "error",
	All:     "all",
	Default: "default",
	None:    "none",
}

func (l Level) String() string {
	return levelCodes[l]
}

// Verbosity maps an integer verbosity level to appropriate Level.  This maybe
// helpful for cmdline applications that need to provide a `-verbosity <int>`
// flag to control logging verbosity.
func Verbosity(v int) Level {
	l := Error

	if v >= 1 {
		l |= Info
	}

	if v >= 2 { //nolint:gomnd
		l |= Debug
	}

	if v >= 3 { //nolint:gomnd
		l |= Trace
	}

	return l
}

// FromString converts the given string label to the appropriate Level. If the
// string does not map to a valid logging level, `None` is returned.
func FromString(s string) Level {
	s = strings.ToLower(strings.TrimSpace(s))

	for level, levelStr := range levelCodes {
		if levelStr == s {
			return level
		}
	}

	return None
}
