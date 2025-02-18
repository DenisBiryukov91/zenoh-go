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

// [Session] liveliness functionality interface.
type Liveliness struct {
	session *Session
}

// Get access to Liveliness functionality.
func (session *Session) Liveliness() *Liveliness {
	return &Liveliness{session: session}
}

// Options to pass to [Liveliness.DeclareToken] operation.
type LivelinessTokenOptions struct {
}

func (opts *LivelinessTokenOptions) toCOpts(_ *runtime.Pinner) C.z_liveliness_token_options_t {
	var cOpts C.z_liveliness_token_options_t
	C.z_liveliness_token_options_default(&cOpts)
	return cOpts
}

// A liveliness token that can be used to provide the network with information about connectivity to its
// declarer: when constructed, a PUT sample will be received by liveliness subscribers on intersecting key
// expressions.
//
// A DELETE on the token's key expression will be received by subscribers if the token is destroyed, or if connectivity between the subscriber and the token's creator is lost.
type LivelinessToken struct {
	token *C.z_owned_liveliness_token_t
}

// Destroy a liveliness token and notify subscribers of its destruction.
func (token *LivelinessToken) Undeclare() error {
	res := int8(C.z_liveliness_undeclare_token(C.z_liveliness_token_move(token.token)))
	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to undeclare LivelinessToken")
}

// Destroy a liveliness token and notify subscribers of its destruction.
func (token *LivelinessToken) Drop() {
	C.z_liveliness_token_drop(C.z_liveliness_token_move(token.token))
}

// Declare a liveliness token on the network.
// Liveliness token subscribers on an intersecting key expression will receive a PUT sample when connectivity
// is achieved, and a DELETE sample if it's lost.
func (liveliness *Liveliness) DeclareToken(keyexpr KeyExpr, options *LivelinessTokenOptions) (LivelinessToken, error) {
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toC(&pinner)

	var cToken C.z_owned_liveliness_token_t
	res := int8(0)
	if options == nil {
		C.z_liveliness_declare_token(C.z_session_loan(liveliness.session.session), &cToken, C.z_view_keyexpr_loan(&cKeyexpr), nil)
	} else {
		cOpts := options.toCOpts(&pinner)
		C.z_liveliness_declare_token(C.z_session_loan(liveliness.session.session), &cToken, C.z_view_keyexpr_loan(&cKeyexpr), &cOpts)
	}
	pinner.Unpin()
	if res == 0 {
		return LivelinessToken{token: &cToken}, nil
	}
	return LivelinessToken{}, NewZError(res, "Failed to declare LivelinessToken")
}

// Options to pass to [Liveliness.DeclareSubscriber] operation.
type LivelinessSubscriberOptions struct {
	History bool // If ``true``, subscriber will receive the state change notifications for liveliness tokens that were declared before its declaration.
}

func (opts *LivelinessSubscriberOptions) toCOpts(_ *runtime.Pinner) C.z_liveliness_subscriber_options_t {
	var cOpts C.z_liveliness_subscriber_options_t
	C.z_liveliness_subscriber_options_default(&cOpts)
	cOpts.history = C.bool(opts.History)
	return cOpts
}

// Declares a subscriber on liveliness tokens that intersect `keyexpr`.
func (liveliness *Liveliness) DeclareSubscriber(keyexpr KeyExpr, handler Handler[Sample], options *LivelinessSubscriberOptions) (Subscriber, error) {
	callback, drop, channel := handler.ToCbDropHandler()
	closure := newClosure(callback, drop)
	var cClosure C.z_owned_closure_sample_t
	C.z_closure_sample(&cClosure, (*[0]byte)(C.zenohSubscriberCallback), (*[0]byte)(C.zenohSubscriberDrop), unsafe.Pointer(closure))
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toC(&pinner)
	res := int8(0)
	var cSubscriber C.z_owned_subscriber_t
	if options == nil {
		res = int8(C.z_liveliness_declare_subscriber(C.z_session_loan(liveliness.session.session), &cSubscriber, C.z_view_keyexpr_loan(&cKeyexpr), C.z_closure_sample_move(&cClosure), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_liveliness_declare_subscriber(C.z_session_loan(liveliness.session.session), &cSubscriber, C.z_view_keyexpr_loan(&cKeyexpr), C.z_closure_sample_move(&cClosure), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return Subscriber{subscriber: &cSubscriber, receiver: channel}, nil
	}
	return Subscriber{}, NewZError(res, "Failed to declare LivelinessSubscriber")
}

// Options to pass to [Liveliness.Get] operation.
type LivelinessGetOptions struct {
	TimeoutMs uint64 // The timeout for the liveliness query in milliseconds. 0 means default query timeout from zenoh configuration.
}

func (opts *LivelinessGetOptions) toCOpts(_ *runtime.Pinner) C.z_liveliness_get_options_t {
	var cOpts C.z_liveliness_get_options_t
	C.z_liveliness_get_options_default(&cOpts)
	cOpts.timeout_ms = C.uint64_t(opts.TimeoutMs)
	return cOpts
}

// Query liveliness tokens currently on the network with a key expression intersecting with `keyexpr`.
func (liveliness *Liveliness) Get(keyexpr KeyExpr, handler Handler[Reply], options *LivelinessGetOptions) (<-chan Reply, error) {
	callback, drop, channel := handler.ToCbDropHandler()
	closure := newClosure(callback, drop)
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toCData(&pinner)
	res := int8(0)
	if options == nil {
		res = int8(C.zc_cgo_liveliness_get(liveliness.session.session, cKeyexpr, unsafe.Pointer(closure), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.zc_cgo_liveliness_get(liveliness.session.session, cKeyexpr, unsafe.Pointer(closure), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return channel, nil
	}
	return nil, NewZError(res, "Failed to perform LivelinessGet operation")
}
