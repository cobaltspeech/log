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

package testinglog

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestLastMinuteFileWriter_concurrent(t *testing.T) {
	tempFile := makeTempFile(t)

	defer func() {
		e := os.Remove(tempFile)
		if e != nil {
			t.Fatalf("failed to remove temp file: %v", e)
		}
	}()

	fw := newLastMinuteFileWriter(tempFile)

	var wg sync.WaitGroup

	numLines := 500

	wg.Add(numLines)

	for i := 0; i < numLines; i++ {
		go func(i int) {
			defer wg.Done()

			fmt.Fprintln(fw, "this is line", i)
		}(i)
	}

	err := fw.actuallyWrite()
	if err != nil {
		t.Fatalf("error in call to actuallyWrite: %v", err)
	}

	wg.Wait()

	err = fw.Close()
	if err != nil {
		t.Fatalf("error closing lastMinuteFileWriter: %v", err)
	}

	// Make sure the lines were all written to the file.
	found := make(map[int]struct{})

	for _, line := range readLines(t, tempFile) {
		if !strings.HasPrefix(line, "this is line ") {
			t.Errorf(`found unexpected line "%s"`, line)
			continue
		}

		var val int

		val, err = strconv.Atoi(line[len("this is line "):])
		if err != nil {
			t.Errorf(`found unexpected line "%s": %v`, line, err)
			continue
		}

		if _, ok := found[val]; ok {
			t.Errorf(`found duplicate line "%s"`, line)
			continue
		}

		found[val] = struct{}{}
	}

	for i := 0; i < numLines; i++ {
		if _, ok := found[i]; !ok {
			t.Errorf(`did not find expected line "this is line %d"`, i)
		}
	}
}

func makeTempFile(t *testing.T) string {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tempFile := f.Name()

	err = f.Close()
	if err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	return tempFile
}

func readLines(t *testing.T, file string) []string {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}

	if b[len(b)-1] != '\n' {
		t.Fatalf("file should end with newline but ends with '%v'", b[len(b)-1])
	}

	b = b[:len(b)-1]

	return strings.Split(string(b), "\n")
}
