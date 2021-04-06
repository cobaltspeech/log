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

package logmap

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
)

// FromKeyvals creates an ordered map containing the keys and values of the log message. The keys
// are formatted to strings, as well as any values that do not implement json.Marshaler or
// encoding.TextMarshaler.
//
// A final value "missing" is inserted if an odd number of values are passed to FromKeyvals.
func FromKeyvals(keyvals ...interface{}) MapSlice {
	n := (len(keyvals) + 1) / 2 // +1 to handle case when len is odd
	m := make(MapSlice, 0, n)

	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]

		var v interface{} = "missing"
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}

		key := fmt.Sprint(k)

		// If v implements json.Marshaler or encoding.TextMarshaler, we
		// give that a priority over fmt.Sprint
		var val interface{}
		switch v.(type) {
		case json.Marshaler:
			val = v
		case encoding.TextMarshaler:
			val = v
		default:
			val = fmt.Sprint(v)
		}

		m = append(m, MapItem{Key: key, Value: val})
	}

	return m
}

type MapItem struct {
	Key   string
	Value interface{}
}

type IndexedMapValue struct {
	Value interface{}
	Index uint64
}

type MapSlice []MapItem

// UnmarshalJSON adds entries from the JSON object to the MapSlice.
func (ms *MapSlice) UnmarshalJSON(b []byte) error {
	m := map[string]IndexedMapValue{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	type indexedEntry struct {
		entry MapItem
		index uint64
	}

	var sorted []indexedEntry
	for k, v := range m {
		sorted = append(sorted, indexedEntry{MapItem{Key: k, Value: v.Value}, v.Index})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].index < sorted[j].index
	})

	for _, v := range sorted {
		*ms = append(*ms, v.entry)
	}

	return nil
}

var indexCounter uint64

func nextIndex() uint64 {
	return atomic.AddUint64(&indexCounter, 1)
}

// UnmarshalJSON for an indexed map item. Used for sorting the resulting MapSlice.
func (mi *IndexedMapValue) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	mi.Value = v
	mi.Index = nextIndex()

	return nil
}

// JSONString creates an ordered JSON representation of the keys and values in the MapSlice.
func (ms MapSlice) JSONString() (string, error) {
	var sb strings.Builder
	enc := json.NewEncoder(&sb)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(ms); err != nil {
		return "", err
	}

	return sb.String(), nil
}

func (ms MapSlice) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.Write([]byte{'{'})

	for i, mi := range ms {
		b, err := json.Marshal(&mi.Value)
		if err != nil {
			return nil, err
		}

		buf.WriteString(fmt.Sprintf("%q:", fmt.Sprintf("%v", mi.Key)))
		buf.Write(b)

		if i < len(ms)-1 {
			buf.Write([]byte{','})
		}
	}

	buf.Write([]byte{'}'})

	return buf.Bytes(), nil
}

func (ms MapSlice) ToStringMap() map[string]string {
	out := map[string]string{}
	for i := range ms {
		out[ms[i].Key] = StringFromValue(ms[i].Value)
	}

	return out
}

// StringFromValue creates a string from the value by calling the value's MarshalJSON or MarshalText
// methods if present, and otherwise calling fmt.Sprint.
func StringFromValue(v interface{}) string {
	switch v.(type) {
	case json.Marshaler:
		valBytes, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}

		return string(valBytes)

	case encoding.TextMarshaler:
		valBytes, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}

		return string(valBytes)

	default:
		return fmt.Sprint(v)
	}
}
