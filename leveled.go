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

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cobaltspeech/log/internal/logmap"
	"github.com/cobaltspeech/log/pkg/level"
)

// LeveledLogger implements the Logger interface and uses the go stdlib log
// package to perform logging.  Each log message has a level prefix followed by
// JSON representation of the data being logged.
type LeveledLogger struct {
	logger      *log.Logger
	filterLevel level.Level
}

// we define osStdErr so that it can be changed for testing
var osStderr io.Writer = os.Stderr

// NewLeveledLogger returns a new Leveledlogger that writes Error and Info
// messages to stderr.  These defaults can be changed by providing Options.
func NewLeveledLogger(opts ...Option) *LeveledLogger {
	l := LeveledLogger{}
	l.filterLevel = level.Default

	for _, opt := range opts {
		opt(&l)
	}

	if l.logger == nil {
		l.logger = log.New(osStderr, "", log.LstdFlags)
	}

	return &l
}

type Option func(*LeveledLogger)

// WithOutput returns an Option that configures the LeveledLogger to write all
// log messages to the given Writer.  Do not combine with WithLogger.
func WithOutput(w io.Writer) Option {
	return func(l *LeveledLogger) {
		l.logger = log.New(w, "", log.LstdFlags)
	}
}

// WithLogger returns an Option that configures the LeveledLogger to use the
// provided log.Logger from go's stdlib package.  Do not combine with WithOutput.
func WithLogger(logger *log.Logger) Option {
	return func(l *LeveledLogger) {
		l.logger = logger
	}
}

// WithFilterLevel configures the new LeveledLogger being created to only log
// messages with the specified logging levels.
func WithFilterLevel(lvl level.Level) Option {
	return func(l *LeveledLogger) {
		l.filterLevel = lvl
	}
}

// SetFilterLevel changes the level of the given logger, at runtime, to the
// provided level.  An application may want to do this to enable debugging
// messages in production, without shutting down and reconfiguring the logger.
//
// This method is expected to be called rarely, and it does not use mutexes to
// lock the level change operations.  Applications may observe temporarily
// indeterminate filtering behavior when this method is called concurrently with
// other logging methods.
func (l *LeveledLogger) SetFilterLevel(lvl level.Level) {
	l.filterLevel = lvl
}

// Error sends the given key value pairs to the error logger.
func (l *LeveledLogger) Error(keyvals ...interface{}) {
	if l.filterLevel&level.Error > 0 {
		l.log(level.Error, keyvals...)
	}
}

// Info sends the given key value pairs to the info logger.
func (l *LeveledLogger) Info(keyvals ...interface{}) {
	if l.filterLevel&level.Info > 0 {
		l.log(level.Info, keyvals...)
	}
}

// Debug sends the given key value pairs to the debug logger.
func (l *LeveledLogger) Debug(keyvals ...interface{}) {
	if l.filterLevel&level.Debug > 0 {
		l.log(level.Debug, keyvals...)
	}
}

// Trace sends the given key value pairs to the trace logger.
func (l *LeveledLogger) Trace(keyvals ...interface{}) {
	if l.filterLevel&level.Trace > 0 {
		l.log(level.Trace, keyvals...)
	}
}

func (l *LeveledLogger) log(lvl level.Level, keyvals ...interface{}) {
	ms := logmap.FromKeyvals(keyvals...)

	line, err := ms.JSONString()
	if err != nil {
		l.logger.Printf(`%-5s {"msg":"logging failure","error":%q}`, level.Error, err)
		return
	}

	l.logger.Print(fmt.Sprintf("%-5s %s", lvl, line))
}
