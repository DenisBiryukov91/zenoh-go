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
	"unsafe"
)

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A struct that indicates if there exist Subscribers matching the Publisher's key expression or Queryables matching Querier's key expression and target.
type MatchingStatus struct {
	Matching bool // ``True`` if there exist matching Zenoh entities, ``false`` otherwise.
}

//export zenohMatchingListenerCallback
func zenohMatchingListenerCallback(status *C.zc_cgo_const_matching_status, context unsafe.Pointer) {
	(*closureContext[MatchingStatus])(context).call(MatchingStatus{Matching: bool(status.matching)})
}

//export zenohMatchingListenerDrop
func zenohMatchingListenerDrop(context unsafe.Pointer) {
	(*closureContext[MatchingStatus])(context).drop()
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A Zenoh matching listener.
//
// A listener that sends notifications when the [MatchingStatus] of a publisher or a querier changes.
type MatchingListener struct {
	listener *C.z_owned_matching_listener_t
	receiver <-chan MatchingStatus
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Return matching listener's receiver if it was constructed with channel, nil otherwise.
func (listener *MatchingListener) Handler() <-chan MatchingStatus {
	return listener.receiver
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Destroy the matching listner.
// This is equivalent to calling [MatchingListener.Undeclare] and discarding its return value.
func (listener *MatchingListener) Drop() {
	C.z_matching_listener_drop(C.z_matching_listener_move(listener.listener))
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Undeclare and destroy the matching listener.
func (listener *MatchingListener) Undeclare() error {
	res := int8(C.z_undeclare_matching_listener(C.z_matching_listener_move(listener.listener)))
	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to undeclare Matching Listener")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Get publisher matching status - i.e. if there are any subscribers matching its key expression.
func (publisher *Publisher) GetMatchingStatus() (MatchingStatus, error) {
	var status C.z_matching_status_t
	res := int8(C.z_publisher_get_matching_status(C.z_publisher_loan(publisher.publisher), &status))

	if res == 0 {
		return MatchingStatus{Matching: bool(status.matching)}, nil
	}
	return MatchingStatus{}, NewZError(res, "Failed to retrieve publisher Matching Status")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Construct matching listener, registering a handler for notifying subscribers matching with a given publisher.
// Matching listener MUST be explicitly destroyed using [MatchingListener.Undeclare] or [MatchingListener.Drop] once it is no longer needed.
func (publisher *Publisher) DeclareMatchingListener(handler Handler[MatchingStatus]) (MatchingListener, error) {
	callback, drop, recv := handler.ToCbDropHandler()
	closure := newClosure(callback, drop)
	var cClosure C.z_owned_closure_matching_status_t
	C.z_closure_matching_status(&cClosure, (*[0]byte)(C.zenohMatchingListenerCallback), (*[0]byte)(C.zenohMatchingListenerDrop), unsafe.Pointer(closure))

	var cMatchingListener C.z_owned_matching_listener_t
	res := int8(C.z_publisher_declare_matching_listener(C.z_publisher_loan(publisher.publisher), &cMatchingListener, C.z_closure_matching_status_move(&cClosure)))

	if res == 0 {
		return MatchingListener{listener: &cMatchingListener, receiver: recv}, nil
	}
	return MatchingListener{}, NewZError(res, "Failed to declare Matching Listener for publisher")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Declare a matching listener, registering a callback for notifying subscribers matching with a given publisher.
// The callback will be run in the background until the corresponding publisher is dropped.
func (publisher *Publisher) DeclareBackgroundMatchingListener(closure Closure[MatchingStatus]) error {
	listenerClosure := newClosure(closure.Call, closure.Drop)
	var cClosure C.z_owned_closure_matching_status_t
	C.z_closure_matching_status(&cClosure, (*[0]byte)(C.zenohMatchingListenerCallback), (*[0]byte)(C.zenohMatchingListenerDrop), unsafe.Pointer(listenerClosure))

	res := int8(C.z_publisher_declare_background_matching_listener(C.z_publisher_loan(publisher.publisher), C.z_closure_matching_status_move(&cClosure)))

	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to declare background Matching Listener for publisher")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Get querier matching status - i.e. if there are any queryables matching its key expression and target.
func (querier *Querier) GetMatchingStatus() (MatchingStatus, error) {
	var status C.z_matching_status_t
	res := int8(C.z_querier_get_matching_status(C.z_querier_loan(querier.querier), &status))

	if res == 0 {
		return MatchingStatus{Matching: bool(status.matching)}, nil
	}
	return MatchingStatus{}, NewZError(res, "Failed to retrieve querier Matching Status")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Construct matching listener, registering a handler for notifying queryables matching with a given querier.
// Matching listener MUST be explicitly destroyed using [MatchingListener.Undeclare] or [MatchingListener.Drop] once it is no longer needed.
func (querier *Querier) DeclareMatchingListener(handler Handler[MatchingStatus]) (MatchingListener, error) {
	callback, drop, recv := handler.ToCbDropHandler()
	closure := newClosure(callback, drop)
	var cClosure C.z_owned_closure_matching_status_t
	C.z_closure_matching_status(&cClosure, (*[0]byte)(C.zenohMatchingListenerCallback), (*[0]byte)(C.zenohMatchingListenerDrop), unsafe.Pointer(closure))

	var cMatchingListener C.z_owned_matching_listener_t
	res := int8(C.z_querier_declare_matching_listener(C.z_querier_loan(querier.querier), &cMatchingListener, C.z_closure_matching_status_move(&cClosure)))

	if res == 0 {
		return MatchingListener{listener: &cMatchingListener, receiver: recv}, nil
	}
	return MatchingListener{}, NewZError(res, "Failed to declare Matching Listener for querier")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Declare a matching listener, registering a callback for notifying queryables matching with a given querier.
// The callback will be run in the background until the corresponding publisher is dropped.
func (querier *Querier) DeclareBackgroundMatchingListener(closure Closure[MatchingStatus]) error {
	listenerClosure := newClosure(closure.Call, closure.Drop)
	var cClosure C.z_owned_closure_matching_status_t
	C.z_closure_matching_status(&cClosure, (*[0]byte)(C.zenohMatchingListenerCallback), (*[0]byte)(C.zenohMatchingListenerDrop), unsafe.Pointer(listenerClosure))

	res := int8(C.z_querier_declare_background_matching_listener(C.z_querier_loan(querier.querier), C.z_closure_matching_status_move(&cClosure)))

	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to declare background Matching Listener for querier")
}
