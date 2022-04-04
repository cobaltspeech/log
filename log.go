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

// Package log provides a Logger interface that may be used in libaries to allow
// four logging levels: Error, Info, Debug, and Trace. It also provides two
// implementations of that interface. The LeveledLogger and DiscardLogger. The
// LevelLogger is meant to be initialized by main packages and registered with
// imported packages. The DiscardLogger is meant to be initialized in libraries
// as a no-op logger in case the main package never provides any logger.
package log

// Logger allows imported Cobalt packages to write logs of the standard four
// levels to a logger provided by the main package. This interface means that
// the logger may be any implementation of logger so long as it provides--or is
// wrapped by a struct that provides--these four functions.
type Logger interface {
	Error(msg string, err error, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Debug(msg string, keyvals ...interface{})
	Trace(msg string, keyvals ...interface{})
}
