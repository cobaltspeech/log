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
	"io"

	"github.com/cobaltspeech/log/internal/logfunc"
	"github.com/cobaltspeech/log/pkg/level"
)

// LeveledLogger writes messages to the provided io.Writer using the go stdlib
// log package.  It implements the Logger interface and thus has four logging
// levels.  Level TRACE is only available in debug builds (build tag
// cobalt_debug)
type LeveledLogger struct {
	errorLogger logfunc.Logger
	infoLogger  logfunc.Logger
	debugLogger logfunc.Logger
	traceLogger logfunc.Logger

	level level.Level
}

// NewLeveledLogger returns a new Leveledlogger with logs filtered at the given level.
func NewLeveledLogger(w io.Writer, lvl level.Level) *LeveledLogger {
	return &LeveledLogger{
		level:       lvl,
		errorLogger: logfunc.NewJSONLogger(w, level.Error),
		infoLogger:  logfunc.NewJSONLogger(w, level.Info),
		debugLogger: logfunc.NewJSONLogger(w, level.Debug),
		traceLogger: logfunc.NewJSONLogger(w, level.Trace),
	}
}

// SetLevel changes the level of the given logger to the provided level.  This
// method should not be called concurrently with other methods.
func (l *LeveledLogger) SetLevel(lvl level.Level) {
	l.level = lvl
}

// Error sends the given key value pairs to the error logger
func (l *LeveledLogger) Error(keyvals ...interface{}) {
	// errors are always logged
	l.errorLogger.Log(keyvals...)
}

// Info sends the given key value pairs to the info logger
func (l *LeveledLogger) Info(keyvals ...interface{}) {
	if l.level <= level.Info {
		l.infoLogger.Log(keyvals...)
	}
}

// Debug sends the given key value pairs to the debug logger
func (l *LeveledLogger) Debug(keyvals ...interface{}) {
	if l.level <= level.Debug {
		l.debugLogger.Log(keyvals...)
	}
}

// Trace is defined in leveled_release.go and leveled_debug.go
