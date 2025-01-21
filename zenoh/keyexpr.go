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
// static const int8_t CGO_Z_EINVAL = Z_EINVAL;
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

// A Zenoh [Key expression].
//
// Key expressions can identify a single key or a set of keys.
//
// Examples :
//   - “"key/expression"“.
//   - “"key/ex*"“.
//
// Internally key expressiobn can be either:
//   - A plain string expression.
//   - A pure numerical id.
//   - The combination of a numerical prefix and a string suffix.
//
// [Key expression]: https://zenoh.io/docs/manual/abstractions/#key-expression
type KeyExpr struct {
	keyexpr []byte
}

func (keyexpr *KeyExpr) toC(pinner *runtime.Pinner) C.z_view_keyexpr_t {
	pinner.Pin(&keyexpr.keyexpr)
	var out C.z_view_keyexpr_t
	pinner.Pin(&keyexpr.keyexpr[0])
	C.z_view_keyexpr_from_substr_unchecked(&out, (*C.char)(unsafe.Pointer(&keyexpr.keyexpr[0])), C.size_t(len(keyexpr.keyexpr)))
	return out
}

func (keyexpr *KeyExpr) toCData(pinner *runtime.Pinner) C.zc_cgo_string_data_t {
	pinner.Pin(&keyexpr.keyexpr[0])
	return C.zc_cgo_string_data_t{str_ptr: (*C.char)(unsafe.Pointer(&keyexpr.keyexpr[0])), len: C.size_t(len(keyexpr.keyexpr))}
}

func newKeyExprFromC(keyexpr C.zc_cgo_string_data_t) KeyExpr {
	ke := KeyExpr{keyexpr: C.GoBytes(unsafe.Pointer(keyexpr.str_ptr), C.int(keyexpr.len))}
	return ke
}

// Construct key expression from string.
//   - `keyexpr` MUST be valid UTF8.
//   - `keyexpr` MUST follow the Key Expression specification, i.e.:
//     1. MUST NOT contain `//`, MUST NOT start nor end with `/`, MUST NOT contain any of the characters `?#$`.
//     2. any instance of `**` may only be lead or followed by `/`.
//     3. the key expression must have canon form.
func NewKeyExpr(keyexpr string) (KeyExpr, error) {
	if len(keyexpr) == 0 {
		return KeyExpr{}, NewZError(C.CGO_Z_EINVAL, "Empty string is not a valid key expression")
	}
	data, size := toDataLen(keyexpr)
	if C.z_keyexpr_is_canon((*C.char)(unsafe.Pointer(&data[0])), C.size_t(size)) == 0 {
		return KeyExpr{keyexpr: data}, nil
	}
	return KeyExpr{}, NewZError(C.CGO_Z_EINVAL, fmt.Sprintf("Failed to construct KeyExpr from: %s", keyexpr))
}

// Construct key expression from string, by first trying to canonize it.
func NewKeyExprAutocanonized(keyexpr string) (KeyExpr, error) {
	if len(keyexpr) == 0 {
		return KeyExpr{}, NewZError(C.CGO_Z_EINVAL, "Empty string is not a valid key expression")
	}
	data, size := toDataLen(keyexpr)
	c_size := C.size_t(size)
	res := int8(C.z_keyexpr_canonize((*C.char)(unsafe.Pointer(&data[0])), &c_size))
	if res == 0 {
		return KeyExpr{keyexpr: data[:int(c_size)]}, nil
	}
	return KeyExpr{}, NewZError(res, fmt.Sprintf("Failed to construct KeyExpr from: %s", keyexpr))
}

// Return a string representing given key expression.
func (keyexpr KeyExpr) String() string {
	return string(keyexpr.keyexpr)
}

// Return “true“ if the key expression intersect, i.e. there exists at least one key which is contained in both of the
// sets defined by “left“ and “right“, “false“ otherwise.
func (left KeyExpr) Intersects(right KeyExpr) bool {
	pinner := runtime.Pinner{}
	cLeft := left.toC(&pinner)
	cRight := right.toC(&pinner)
	res := C.z_keyexpr_intersects(C.z_view_keyexpr_loan(&cLeft), C.z_view_keyexpr_loan(&cRight))
	pinner.Unpin()
	return bool(res)
}

// Return “true“ if “left“ includes “right“, i.e. the set defined by “left“ contains every key belonging to the set
// defined by “right“, “false“ otherwise.
func (left KeyExpr) Includes(right KeyExpr) bool {
	pinner := runtime.Pinner{}
	cLeft := left.toC(&pinner)
	cRight := right.toC(&pinner)
	res := C.z_keyexpr_includes(C.z_view_keyexpr_loan(&cLeft), C.z_view_keyexpr_loan(&cRight))
	pinner.Unpin()
	return bool(res)
}

// Construct key expression by performing path-joining (automatically inserting '/' in-between) of `left` with `right`.
func (left KeyExpr) Join(right string) (KeyExpr, error) {
	return NewKeyExprAutocanonized(string(left.keyexpr) + "/" + right)
}

// Perform string concatenation and return the result as a `KeyExpr` if possible.
// You should probably prefer [KeyExpr.Join] as Zenoh may then take advantage of the hierarchical separation it inserts.
func (left KeyExpr) Concat(right string) (KeyExpr, error) {
	return NewKeyExpr(string(left.keyexpr) + right)
}
