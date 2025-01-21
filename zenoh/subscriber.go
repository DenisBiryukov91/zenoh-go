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

//export zenohSubscriberCallbackData
func zenohSubscriberCallbackData(sample C.zc_cgo_sample_data_t, context unsafe.Pointer) {
	(*closureContext[Sample])(context).call(newSampleFromC(sample))
}

//export zenohSubscriberDrop
func zenohSubscriberDrop(context unsafe.Pointer) {
	(*closureContext[Sample])(context).drop()
}

// A Zenoh [subscriber].
// Receives data from publication on intersecting key expressions.
// Destroying the subscriber cancels the subscription.
//
// [subscriber]: https://zenoh.io/docs/manual/abstractions/#subscriber.
type Subscriber struct {
	subscriber *C.z_owned_subscriber_t
}

// Undeclare and destroy the subscriber.
func (subscriber *Subscriber) Undeclare() error {
	res := int8(C.z_undeclare_subscriber(C.z_subscriber_move(subscriber.subscriber)))
	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to undeclare Subscriber")
}

// Destroy the subscriber.
// This is equivalent to calling [Subscriber.Undeclare] and discarding its return value.
func (subscriber *Subscriber) Drop() {
	C.z_subscriber_drop(C.z_subscriber_move(subscriber.subscriber))
}

func (subscriber *Subscriber) KeyExpr() KeyExpr {
	return newKeyExprFromC(C.zc_cgo_keyexpr_get_data(C.z_subscriber_keyexpr(C.z_subscriber_loan(subscriber.subscriber))))
}

// Options passed to subscriber declaration
type SubscriberOptions struct {
}

func (opts *SubscriberOptions) toCOpts(_ *runtime.Pinner) C.z_subscriber_options_t {
	var cOpts C.z_subscriber_options_t
	C.z_subscriber_options_default(&cOpts)
	return cOpts
}

// Construct a subscriber for the given key expression.
// Queryable MUST be explicitly destroyed using [Subscriber.Undeclare] or [Subscriber.Drop] once it is no longer needed.
func (session *Session) DeclareSubscriber(keyexpr KeyExpr, callback func(Sample), drop func(), options *SubscriberOptions) (Subscriber, error) {
	closure := newClosure(callback, drop)
	var cClosure C.z_owned_closure_sample_t
	C.z_closure_sample(&cClosure, (*[0]byte)(C.zenohSubscriberCallback), (*[0]byte)(C.zenohSubscriberDrop), unsafe.Pointer(closure))
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toC(&pinner)
	res := int8(0)
	var cSubscriber C.z_owned_subscriber_t
	if options == nil {
		res = int8(C.z_declare_subscriber(C.z_session_loan(session.session), &cSubscriber, C.z_view_keyexpr_loan(&cKeyexpr), C.z_closure_sample_move(&cClosure), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_declare_subscriber(C.z_session_loan(session.session), &cSubscriber, C.z_view_keyexpr_loan(&cKeyexpr), C.z_closure_sample_move(&cClosure), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return Subscriber{subscriber: &cSubscriber}, nil
	}
	return Subscriber{}, NewZError(res, "Failed to declare Subscriber")
}

// Construct and declare a background subscriber. Subscriber callback will be called to process the messages,
// until the corresponding session is closed or dropped.
func (session *Session) DeclareBackgroundSubscriber(keyexpr KeyExpr, callback func(Sample), drop func(), options *SubscriberOptions) error {
	closure := newClosure(callback, drop)
	var cClosure C.z_owned_closure_sample_t
	C.z_closure_sample(&cClosure, (*[0]byte)(C.zenohSubscriberCallback), (*[0]byte)(C.zenohSubscriberDrop), unsafe.Pointer(closure))
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toC(&pinner)
	res := int8(0)
	if options == nil {
		res = int8(C.z_declare_background_subscriber(C.z_session_loan(session.session), C.z_view_keyexpr_loan(&cKeyexpr), C.z_closure_sample_move(&cClosure), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_declare_background_subscriber(C.z_session_loan(session.session), C.z_view_keyexpr_loan(&cKeyexpr), C.z_closure_sample_move(&cClosure), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to declare background Subscriber")
}
