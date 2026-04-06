//
// Copyright (c) 2026 ZettaScale Technology
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

package zenohext

// #cgo CFLAGS: -I${SRCDIR}/..
// #include "zenoh.h"
import "C"

import (
	"runtime"
	"unsafe"
	"zenoh-go/zenoh"
)

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A Zenoh session extension, providing access to additional APIs that are not available in the main [zenoh.Session] type. To obtain an instance of this type, use the [Ext] function.
type SessionExt struct {
	session *zenoh.Session
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Obtain a session extension for the given session to access additional APIs that are not available in the main [zenoh.Session] type.
func Ext(session *zenoh.Session) *SessionExt {
	return &SessionExt{session: session}
}

//go:linkname sessionGetInner zenoh-go/zenoh.sessionGetInner
func sessionGetInner(*zenoh.Session) unsafe.Pointer

func (ext *SessionExt) getInner() *C.z_owned_session_t {
	return (*C.z_owned_session_t)(sessionGetInner(ext.session))
}

//go:linkname encodingToUnsafeCPtr zenoh-go/zenoh.encodingToUnsafeCPtr
func encodingToUnsafeCPtr(zenoh.Encoding) unsafe.Pointer

//go:linkname keyExprToUnsafeCPtr zenoh-go/zenoh.keyExprToUnsafeCPtr
func keyExprToUnsafeCPtr(*zenoh.KeyExpr, *runtime.Pinner) unsafe.Pointer

//go:linkname zbytesToUnsafeCData zenoh-go/zenoh.zbytesToUnsafeCData
func zbytesToUnsafeCData(zenoh.ZBytes, *runtime.Pinner, unsafe.Pointer)

//go:linkname newKeyExprFromUnsafeCDataPtr zenoh-go/zenoh.newKeyExprFromUnsafeCDataPtr
func newKeyExprFromUnsafeCDataPtr(unsafe.Pointer) zenoh.KeyExpr

//go:linkname encodingToUnsafeCData zenoh-go/zenoh.encodingToUnsafeCData
func encodingToUnsafeCData(zenoh.Encoding, *runtime.Pinner, unsafe.Pointer)

//go:linkname timeStampToUnsafeC zenoh-go/zenoh.timeStampToUnsafeC
func timeStampToUnsafeC(zenoh.TimeStamp, unsafe.Pointer)

//go:linkname sourceInfoToUnsafeC zenoh-go/zenoh.sourceInfoToUnsafeC
func sourceInfoToUnsafeC(zenoh.SourceInfo, unsafe.Pointer)

//go:linkname matchingListenerFromUnsafeCPtrAndReceiver zenoh-go/zenoh.matchingListenerFromUnsafeCPtrAndReceiver
func matchingListenerFromUnsafeCPtrAndReceiver(unsafe.Pointer, <-chan zenoh.MatchingStatus) zenoh.MatchingListener

//go:linkname subscriberFromUnsafeCPtrAndReceiver zenoh-go/zenoh.subscriberFromUnsafeCPtrAndReceiver
func subscriberFromUnsafeCPtrAndReceiver(unsafe.Pointer, <-chan zenoh.Sample) zenoh.Subscriber

//go:linkname newIdFromUnsafeCPtr zenoh-go/zenoh.newIdFromUnsafeCPtr
func newIdFromUnsafeCPtr(unsafe.Pointer) zenoh.Id

//go:linkname newEntityGlobalIdFromUnsafeCPtr zenoh-go/zenoh.newEntityGlobalIdFromUnsafeCPtr
func newEntityGlobalIdFromUnsafeCPtr(unsafe.Pointer) zenoh.EntityGlobalId

//go:linkname newZError zenoh-go/zenoh.newZError
func newZError(code int8) zenoh.ZError
