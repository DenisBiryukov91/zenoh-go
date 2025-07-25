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

	"github.com/BooleanCat/option"
)

//export zenohQueryableCallbackData
func zenohQueryableCallbackData(query C.zc_cgo_query_data_t, context unsafe.Pointer) {
	(*closureContext[Query])(context).call(newQueryFromC(query))
}

//export zenohQueryableDrop
func zenohQueryableDrop(context unsafe.Pointer) {
	(*closureContext[Query])(context).drop()
}

// A Zenoh [queryable]. Responds to queries sent via [Session.Get] with intersecting key expression.
//
// [queryable]: https://zenoh.io/docs/manual/abstractions/#queryable
type Queryable struct {
	queryable *C.z_owned_queryable_t
	receiver  <-chan Query
}

// Undeclare and destroy the queryable.
func (queryable *Queryable) Undeclare() error {
	res := int8(C.z_undeclare_queryable(C.z_queryable_move(queryable.queryable)))
	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to undeclare Queryable")
}

// Return Queryable receiver if it was constructed with channel, nil otherwise.
func (queryable *Queryable) Handler() <-chan Query {
	return queryable.receiver
}

// Destroy the queryable.
// This is equivalent to calling [Queryable.Undeclare] and discarding its return value.
func (queryable *Queryable) Drop() {
	C.z_queryable_drop(C.z_queryable_move(queryable.queryable))
}

// Options passed to queryable declaration.
type QueryableOptions struct {
	Complete      bool                    // The completeness of the Queryable
	AllowedOrigin option.Option[Locality] // Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release. Restrict the matching requests that will be received by this Queryable to the ones that have the compatible AllowedDestination.
}

func (opts *QueryableOptions) toCOpts(_ *runtime.Pinner) C.z_queryable_options_t {
	var cOpts C.z_queryable_options_t
	C.z_queryable_options_default(&cOpts)
	cOpts.complete = C.bool(opts.Complete)
	if opts.AllowedOrigin.IsSome() {
		cOpts.allowed_origin = uint32(opts.AllowedOrigin.Unwrap())
	}
	return cOpts
}

// Construct a queryable for the given key expression.
// Queryable MUST be explicitly destroyed using [Queryable.Undeclare] or [Queryable.Drop] once it is no longer needed.
func (session *Session) DeclareQueryable(keyexpr KeyExpr, handler Handler[Query], options *QueryableOptions) (Queryable, error) {
	callback, drop, channel := handler.ToCbDropHandler()
	closure := newClosure(callback, drop)
	var cClosure C.z_owned_closure_query_t
	C.z_closure_query(&cClosure, (*[0]byte)(C.zenohQueryableCallback), (*[0]byte)(C.zenohQueryableDrop), unsafe.Pointer(closure))
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toC(&pinner)
	res := int8(0)
	var cQueryable C.z_owned_queryable_t
	if options == nil {
		res = int8(C.z_declare_queryable(C.z_session_loan(session.session), &cQueryable, C.z_view_keyexpr_loan(&cKeyexpr), C.z_closure_query_move(&cClosure), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_declare_queryable(C.z_session_loan(session.session), &cQueryable, C.z_view_keyexpr_loan(&cKeyexpr), C.z_closure_query_move(&cClosure), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return Queryable{queryable: &cQueryable, receiver: channel}, nil
	}
	return Queryable{}, NewZError(res, "Failed to declare Queryable")
}

// Declare a background queryable for a given keyexpr. The queryable callback will be be called
// to proccess incoming queries until the corresponding session is closed or dropped.
func (session *Session) DeclareBackgroundQueryable(keyexpr KeyExpr, closure Closure[Query], options *QueryableOptions) error {
	qClosure := newClosure(closure.Call, closure.Drop)
	var cClosure C.z_owned_closure_query_t
	C.z_closure_query(&cClosure, (*[0]byte)(C.zenohQueryableCallback), (*[0]byte)(C.zenohQueryableDrop), unsafe.Pointer(qClosure))
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toC(&pinner)
	res := int8(0)
	if options == nil {
		res = int8(C.z_declare_background_queryable(C.z_session_loan(session.session), C.z_view_keyexpr_loan(&cKeyexpr), C.z_closure_query_move(&cClosure), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_declare_background_queryable(C.z_session_loan(session.session), C.z_view_keyexpr_loan(&cKeyexpr), C.z_closure_query_move(&cClosure), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to declare background Queryable")
}

// Get the key expression of the queryable.
func (queryable *Queryable) KeyExpr() KeyExpr {
	return newKeyExprFromC(C.zc_cgo_keyexpr_get_data(C.z_queryable_keyexpr(C.z_queryable_loan(queryable.queryable))))
}
