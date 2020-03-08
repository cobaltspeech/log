// +build cobalt_log_trace

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

import "github.com/cobaltspeech/log/pkg/level"

// Trace sends the given key value pairs to the trace logger
func (l *LeveledLogger) Trace(keyvals ...interface{}) {
	if l.filterLevel&level.Trace > 0 {
		l.log(level.Trace, keyvals...)
	}
}
