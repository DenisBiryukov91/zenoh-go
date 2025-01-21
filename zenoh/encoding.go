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
// #include "zenoh_cgo.h"
import "C"
import (
	"runtime"
	"unsafe"
)

// The [encoding] of Zenoh data.
//
// [encoding]: https://zenoh.io/docs/manual/abstractions/#encoding
type Encoding struct {
	id     uint16
	schema []byte
}

// Construct default encoding.
func NewEncodingDefault() Encoding {
	return newEncodingFromC(C.zc_internal_encoding_get_data(C.z_encoding_loan_default()))
}

// Construct encoding from string.
func NewEncodinFromString(encoding string) Encoding {
	data, len := toDataLen(encoding)
	var cEncoding C.z_owned_encoding_t
	C.z_encoding_from_substr(&cEncoding, (*C.char)(unsafe.Pointer(&data[0])), C.size_t(len))
	return newEncodingFromOwnedC(&(cEncoding))
}

// Get string representation of encoding.
func (encoding *Encoding) String() string {
	cEncoding := encoding.toC()
	var s C.z_owned_string_t
	C.z_encoding_to_string(C.z_encoding_loan(&cEncoding), &s)
	cStringData := C.zc_cgo_string_get_data(C.z_string_loan(&s))
	out := C.GoStringN(cStringData.str_ptr, C.int(cStringData.len))
	C.zc_cgo_string_drop(&s)
	C.zc_cgo_encoding_drop(&cEncoding)
	return out
}

// Set schema to this encoding from a string.
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
	*encoding = newEncodingFromOwnedC(&cEncoding)
}

func (encoding Encoding) toCData(pinner *runtime.Pinner) C.zc_internal_encoding_data_t {
	var encodingData C.zc_internal_encoding_data_t
	encodingData.id = C.uint16_t(encoding.id)
	if len(encoding.schema) > 0 {
		pinner.Pin(&encoding.schema[0])
		encodingData.schema_ptr = (*C.uint8_t)(unsafe.Pointer(&encoding.schema[0]))
		encodingData.schema_len = C.size_t(len(encoding.schema))
	}
	return encodingData
}

func (encoding Encoding) toC() C.z_owned_encoding_t {
	var out C.z_owned_encoding_t
	pinner := runtime.Pinner{}
	encodingData := encoding.toCData(&pinner)
	C.zc_internal_encoding_from_data(&out, encodingData)
	pinner.Unpin()
	return out
}

func newEncodingFromC(cEncoding C.zc_internal_encoding_data_t) Encoding {
	return Encoding{id: uint16(cEncoding.id), schema: C.GoBytes(unsafe.Pointer(cEncoding.schema_ptr), C.int(cEncoding.schema_len))}
}

func newEncodingFromOwnedC(cEncoding *C.z_owned_encoding_t) Encoding {
	e := newEncodingFromC(C.zc_internal_encoding_get_data(C.z_encoding_loan(cEncoding)))
	C.zc_cgo_encoding_drop(cEncoding)
	return e
}
