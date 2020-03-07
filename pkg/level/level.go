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

// Package level defines the supported logging levels
package level

// Level enumerates different logging levels.
type Level byte

const (
	Trace Level = 1 << iota
	Debug
	Info
	Error
)

// levelCodes provides a string representation of different supported levels.
var levelCodes = map[Level]string{
	Trace: "trace",
	Debug: "debug",
	Info:  "info",
	Error: "error",
}

func (l Level) String() string {
	return levelCodes[l]
}
