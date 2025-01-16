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
	"runtime"

	"github.com/BooleanCat/option"
)

type Publisher struct {
	publisher *C.z_owned_publisher_t
}

// Options passed to Publisher Put operation.
type PublisherPutOptions struct {
	Encoding    option.Option[Encoding]  // The encoding of the publication.
	Attachement option.Option[ZBytes]    // The attachment to attach to the publication.
	TimeStamp   option.Option[TimeStamp] // The timestamp of the publication.
}

func (opts *PublisherPutOptions) toCOpts(pinner *runtime.Pinner) C.z_publisher_put_options_t {
	var cOpts C.z_publisher_put_options_t
	C.z_publisher_put_options_default(&cOpts)
	if opts.Attachement.IsSome() {
		cAttachment := opts.Attachement.Unwrap().toC()
		pinner.Pin(&cAttachment)
		cOpts.attachment = C.z_bytes_move(&cAttachment)
	}
	if opts.Encoding.IsSome() {
		cEncoding := opts.Encoding.Unwrap().toC()
		pinner.Pin(&cEncoding)
		cOpts.encoding = C.z_encoding_move(&cEncoding)
	}
	if opts.TimeStamp.IsSome() {
		var c_timestamp = opts.TimeStamp.Unwrap().timestamp
		cOpts.timestamp = &c_timestamp
	}
	return cOpts
}

// Options passed to Publisher Delete operation.
type PublisherDeleteOptions struct {
	TimeStamp option.Option[TimeStamp] // The timestamp of the publication.
}

func (opts *PublisherDeleteOptions) toCOpts(_ *runtime.Pinner) C.z_publisher_delete_options_t {
	var cOpts C.z_publisher_delete_options_t
	C.z_publisher_delete_options_default(&cOpts)
	if opts.TimeStamp.IsSome() {
		var c_timestamp = opts.TimeStamp.Unwrap().timestamp
		cOpts.timestamp = &c_timestamp
	}
	return cOpts
}

// Undeclare and destroy the publisher.
func (publisher *Publisher) Undeclare() error {
	res := int8(C.z_undeclare_publisher(C.z_publisher_move(publisher.publisher)))
	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to undeclare Publisher")
}

// Destroy the publisher.
// This is equivalent to calling [Publisher.Undeclare] and discarding its return value.
func (publisher *Publisher) Drop() {
	C.z_publisher_drop(C.z_publisher_move(publisher.publisher))
}

// Publish message onto the publisher's key expression.
func (publisher *Publisher) Put(payload ZBytes, options *PublisherPutOptions) error {
	cPayload := payload.toC()
	pinner := runtime.Pinner{}
	res := int8(0)
	if options == nil {
		res = int8(C.z_publisher_put(C.z_publisher_loan(publisher.publisher), C.z_bytes_move(&cPayload), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_publisher_put(C.z_publisher_loan(publisher.publisher), C.z_bytes_move(&cPayload), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to perform Publisher Put operation")
}

// Publish a `DELETE` message onto the publisher's key expression.
func (publisher *Publisher) Delete(options *PublisherDeleteOptions) error {
	pinner := runtime.Pinner{}
	res := int8(0)
	if options == nil {
		res = int8(C.z_publisher_delete(C.z_publisher_loan(publisher.publisher), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_publisher_delete(C.z_publisher_loan(publisher.publisher), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to perform Publisher Delete operation")
}

// Get the key expression of the publisher.
func (publisher *Publisher) KeyExpr() KeyExpr {
	return newKeyExprFromC(C.z_publisher_keyexpr(C.z_publisher_loan(publisher.publisher)))
}

// Options passed to publisher declaration.
type PublisherOptions struct {
	Encoding          option.Option[Encoding]          // Default encoding for messages put by this publisher.
	CongestionControl option.Option[CongestionControl] // The congestion control to apply when routing messages from this publisher.
	Priority          option.Option[Priority]          // The priority of messages from this publisher.
	IsExpress         bool                             // If set to ``true``, the messages of this publisher will not be batched. This usually has a positive impact on latency but negative impact on throughput.
}

func (opts *PublisherOptions) toCOpts(pinner *runtime.Pinner) C.z_publisher_options_t {
	var cOpts C.z_publisher_options_t
	C.z_publisher_options_default(&cOpts)
	if opts.Encoding.IsSome() {
		cEncoding := opts.Encoding.Unwrap().toC()
		pinner.Pin(cEncoding)
		cOpts.encoding = C.z_encoding_move(&cEncoding)
	}
	if opts.Priority.IsSome() {
		cOpts.priority = uint32(C.z_priority_t(opts.Priority.Unwrap()))
	}
	if opts.CongestionControl.IsSome() {
		cOpts.congestion_control = uint32(opts.CongestionControl.Unwrap())
	}
	cOpts.is_express = C.bool(opts.IsExpress)
	return cOpts
}

// Construct and declare a publisher for the given key expression.
//
// Data can be put and deleted with this publisher with the help of the
// [Publisher.Put] and [Publisher.Delete] functions.
// Publisher MUST be explicitly destroyed using [Publisher.Undeclare] or [Publisher.Drop] once it is no longer needed.
func (session *Session) DeclarePublisher(keyexpr KeyExpr, options *PublisherOptions) (Publisher, error) {
	res := int8(0)
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toC(&pinner)
	var cPublisher C.z_owned_publisher_t

	if options == nil {
		res = int8(C.z_declare_publisher(C.z_session_loan(session.session), &cPublisher, C.z_view_keyexpr_loan(&cKeyexpr), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_declare_publisher(C.z_session_loan(session.session), &cPublisher, C.z_view_keyexpr_loan(&cKeyexpr), &cOpts))
	}
	pinner.Unpin()
	if res == 0 {
		return Publisher{publisher: &cPublisher}, nil
	}
	return Publisher{}, NewZError(res, "Failed to declare Publisher")
}

// Options passed to Session Put operation.
type PutOptions struct {
	Encoding          option.Option[Encoding]          // The encoding of the publication.
	Attachement       option.Option[ZBytes]            // The attachment to attach to the publication.
	TimeStamp         option.Option[TimeStamp]         // The timestamp of the publication.
	CongestionControl option.Option[CongestionControl] // The congestion control to apply when routing the publication.
	Priority          option.Option[Priority]          // The priority of the publication.
	IsExpress         bool                             // If set to ``true``, the message will not be batched. This usually has a positive impact on latency but negative impact on throughput.
}

func (opts *PutOptions) toCOpts(pinner *runtime.Pinner) C.z_put_options_t {
	var cOpts C.z_put_options_t
	C.z_put_options_default(&cOpts)
	if opts.Attachement.IsSome() {
		cAttachment := opts.Attachement.Unwrap().toC()
		pinner.Pin(&cAttachment)
		cOpts.attachment = C.z_bytes_move(&cAttachment)
	}
	if opts.Encoding.IsSome() {
		cEncoding := opts.Encoding.Unwrap().toC()
		pinner.Pin(&cEncoding)
		cOpts.encoding = C.z_encoding_move(&cEncoding)
	}
	if opts.TimeStamp.IsSome() {
		var c_timestamp = opts.TimeStamp.Unwrap().timestamp
		cOpts.timestamp = &c_timestamp
	}
	if opts.Priority.IsSome() {
		cOpts.priority = uint32(C.z_priority_t(opts.Priority.Unwrap()))
	}
	if opts.CongestionControl.IsSome() {
		cOpts.congestion_control = uint32(opts.CongestionControl.Unwrap())
	}
	cOpts.is_express = C.bool(opts.IsExpress)
	return cOpts
}

// Options passed to Session Delete operation.
type DeleteOptions struct {
	TimeStamp         option.Option[TimeStamp]         // The timestamp of the delete message.
	CongestionControl option.Option[CongestionControl] // The congestion control to apply when routing the delete message.
	Priority          option.Option[Priority]          // The priority of the delete message.
	IsExpress         bool                             // If set to ``true``, the delete message will not be batched. This usually has a positive impact on latency but negative impact on throughput.
}

func (opts *DeleteOptions) toCOpts(_ *runtime.Pinner) C.z_delete_options_t {
	var cOpts C.z_delete_options_t
	C.z_delete_options_default(&cOpts)
	if opts.TimeStamp.IsSome() {
		var c_timestamp = opts.TimeStamp.Unwrap().timestamp
		cOpts.timestamp = &c_timestamp
	}
	if opts.Priority.IsSome() {
		cOpts.priority = uint32(C.z_priority_t(opts.Priority.Unwrap()))
	}
	if opts.CongestionControl.IsSome() {
		cOpts.congestion_control = uint32(opts.CongestionControl.Unwrap())
	}
	cOpts.is_express = C.bool(opts.IsExpress)
	return cOpts
}

// Publish data on specified key expression.
func (session *Session) Put(keyExpr KeyExpr, payload ZBytes, options *PutOptions) error {
	cPayload := payload.toC()
	pinner := runtime.Pinner{}
	cKeyexpr := keyExpr.toC(&pinner)
	res := int8(0)
	if options == nil {
		res = int8(C.z_put(C.z_session_loan(session.session), C.z_view_keyexpr_loan(&cKeyexpr), C.z_bytes_move(&cPayload), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_put(C.z_session_loan(session.session), C.z_view_keyexpr_loan(&cKeyexpr), C.z_bytes_move(&cPayload), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to perform Put operation")
}

// Send request to delete data on specified key expression (used when working with [Zenoh storages]).
//
// [Zenoh storages]: https://zenoh.io/docs/manual/abstractions/#storage
func (session *Session) Delete(keyExpr KeyExpr, options *DeleteOptions) error {
	pinner := runtime.Pinner{}
	cKeyexpr := keyExpr.toC(&pinner)
	res := int8(0)
	if options == nil {
		res = int8(C.z_delete(C.z_session_loan(session.session), C.z_view_keyexpr_loan(&cKeyexpr), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_delete(C.z_session_loan(session.session), C.z_view_keyexpr_loan(&cKeyexpr), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to perform Delete operation")
}
