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

package testinglog

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// lastMinuteFileWriter caches writes until ActuallyWrite is called. It only creates a file at the
// specified path when it begins to actually write. It also creates the directory containing the
// specified path at that point.
type lastMinuteFileWriter struct {
	path string

	// w is a *os.File if actuallyWrite has been called; otherwise it is a *bytes.Buffer.
	w    io.Writer
	b    *bytes.Buffer // This is the buffer that is w until actuallyWrite is called.
	f    *os.File      // This is the file   that is w after actuallyWrite is called.
	lock sync.Mutex
}

func newLastMinuteFileWriter(file string) *lastMinuteFileWriter {
	b := bytes.Buffer{}
	return &lastMinuteFileWriter{path: file, w: &b, b: &b}
}

// Write caches writes for later unless ActuallyWrite has been called.
func (fw *lastMinuteFileWriter) Write(p []byte) (n int, err error) {
	fw.lock.Lock()
	defer fw.lock.Unlock()

	return fw.w.Write(p)
}

func (fw *lastMinuteFileWriter) actuallyWrite() error {
	// No one should write/read the cache or check doWrite while we're doing this.
	fw.lock.Lock()
	defer fw.lock.Unlock()

	if fw.f != nil {
		// We already called actuallyWrite.
		return nil
	}

	err := os.MkdirAll(filepath.Dir(fw.path), os.ModePerm)
	if err != nil {
		return err
	}

	fw.f, err = os.Create(fw.path)
	if err != nil {
		return err
	}

	// Write the cache out.
	_, err = fw.f.Write(fw.b.Bytes())
	if err != nil {
		return err
	}

	// Replace the *bytes.Buffer with the *os.File.
	fw.w = fw.f
	fw.b = nil

	return nil
}

func (fw *lastMinuteFileWriter) Close() error {
	if fw.f == nil {
		// We never opened the file.
		return nil
	}

	return fw.f.Close()
}
