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

// With returns a new contextual Logger with keyvals prepended to those passed
// to calls to the new logger.
func With(l Logger, keyvals ...interface{}) Logger {
	if len(keyvals) == 0 {
		return l
	}

	return &contextLogger{l, keyvals, ""}
}

// WithMsgPrefix returns a new contextual Logger that will prepend the log message with the given string.
func WithMsgPrefix(l Logger, prefix string) Logger {
	if prefix == "" {
		return l
	}

	return &contextLogger{l, nil, prefix}
}

type contextLogger struct {
	log       Logger
	keyvals   []interface{}
	msgPrefix string
}

func (c *contextLogger) Error(msg string, err error, keyvals ...interface{}) {
	kvs := append(c.keyvals, keyvals...)
	c.log.Error(c.msgPrefix+msg, err, kvs...)
}

func (c *contextLogger) Info(msg string, keyvals ...interface{}) {
	kvs := append(c.keyvals, keyvals...)
	c.log.Info(c.msgPrefix+msg, kvs...)
}

func (c *contextLogger) Debug(msg string, keyvals ...interface{}) {
	kvs := append(c.keyvals, keyvals...)
	c.log.Debug(c.msgPrefix+msg, kvs...)
}

func (c *contextLogger) Trace(msg string, keyvals ...interface{}) {
	kvs := append(c.keyvals, keyvals...)
	c.log.Trace(c.msgPrefix+msg, kvs...)
}
