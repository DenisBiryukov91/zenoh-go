//
// Copyright (c) 2025 ZettaScale Technology
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
// which is available at https://www.apache.org/licenses/LICENSE-2.0.
//
// SPDX-License-Identifier: EPL-2.0 OR Apache-2.0
//
// Contributors:
//   ZettaScale Zenoh Team, <zenoh@zettascale.tech>
//

package zenoh

// #include "zenoh.h"
import "C"
import (
	"unsafe"
)

// The [encoding] of Zenoh data.
//
// [encoding]: https://zenoh.io/docs/manual/abstractions/#encoding
type Encoding struct {
	encoding string
}

// Construct default encoding.
func NewEncodingDefault() Encoding {
	var s C.z_owned_string_t
	var e Encoding
	C.z_encoding_to_string(C.z_encoding_zenoh_bytes(), &s)
	loanedString := C.z_string_loan(&s)
	e.encoding = C.GoStringN(C.z_string_data(loanedString), C.int(C.z_string_len(loanedString)))
	C.z_string_drop(C.z_string_move(&s))
	return e
}

// Construct encoding from string.
func NewEncodinFromString(encoding string) Encoding {
	return Encoding{encoding: encoding}
}

// Get string representation of encoding.
func (encoding *Encoding) String() string {
	return encoding.encoding
}

//	Set schema to this encoding from a string.
//
// Zenoh does not define what a schema is and its semantics is left to the implementer.
// E.g. a common schema for `text/plain` encoding is `utf-8`.
func (encoding *Encoding) SetSchema(schema string) {
	if len(schema) == 0 {
		return
	}
	cEncoding := encoding.toC()
	schemaData, schemaSize := toDataLen(schema)
	loanedEncoding := C.z_encoding_loan(&cEncoding)
	C.z_encoding_set_schema_from_substr(loanedEncoding, (*C.char)(unsafe.Pointer(&schemaData[0])), C.size_t(schemaSize))
	encoding.encoding = newEncodingFromC(loanedEncoding).encoding
	C.z_encoding_drop(C.z_encoding_move(&cEncoding))
}

func (encoding Encoding) toC() C.z_owned_encoding_t {
	var out C.z_owned_encoding_t
	data, size := toDataLen(encoding.encoding)
	if size > 0 {
		C.z_encoding_from_substr(&out, (*C.char)(unsafe.Pointer(&data[0])), C.size_t(size))
	} else {
		C.z_encoding_from_substr(&out, nil, C.size_t(size))
	}
	return out
}

func newEncodingFromC(cEncoding *C.z_loaned_encoding_t) Encoding {
	var s C.z_owned_string_t
	C.z_encoding_to_string(cEncoding, &s)
	loanedString := C.z_string_loan(&s)
	encoding := Encoding{encoding: C.GoStringN(C.z_string_data(loanedString), C.int(C.z_string_len(loanedString)))}
	C.z_string_drop(C.z_string_move(&s))
	return encoding
}
