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

package log

// DiscardLogger implements the Logger interface and ignores all logging
// messages.
type DiscardLogger struct{}

func NewDiscardLogger() *DiscardLogger {
	return &DiscardLogger{}
}

func (l *DiscardLogger) Error(keyvals ...interface{}) {}
func (l *DiscardLogger) Info(keyvals ...interface{})  {}
func (l *DiscardLogger) Debug(keyvals ...interface{}) {}
func (l *DiscardLogger) Trace(keyvals ...interface{}) {}
