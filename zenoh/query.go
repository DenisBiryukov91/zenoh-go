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

// A Zenoh query received by a queryable.
type Query struct {
	query *C.z_owned_query_t
}

// Finalizes and destroys the query. This MUST be always called by user once all replies are provided.
func (query *Query) Drop() {
	C.z_query_drop(C.z_query_move(query.query))
}

// Return the key expression of the query.
func (query *Query) KeyExpr() KeyExpr {
	return newKeyExprFromC(C.z_query_keyexpr(C.z_query_loan(query.query)))
}

// Return query payload data if there is any.
func (query *Query) Payload() option.Option[ZBytes] {
	cPayload := C.z_query_payload(C.z_query_loan(query.query))
	if cPayload == nil {
		return option.None[ZBytes]()
	} else {
		return option.Some(newZBytesFromC(cPayload))
	}
}

// Return the encoding associated with the query data, if there is any.
func (query *Query) Encoding() option.Option[Encoding] {
	cEncoding := C.z_query_encoding(C.z_query_loan(query.query))
	if cEncoding == nil {
		return option.None[Encoding]()
	} else {
		return option.Some(newEncodingFromC(cEncoding))
	}
}

// Return query attachment if there is any.
func (query *Query) Attachement() option.Option[ZBytes] {
	cAttachment := C.z_query_attachment(C.z_query_loan(query.query))
	if cAttachment == nil {
		return option.None[ZBytes]()
	} else {
		return option.Some(newZBytesFromC(cAttachment))
	}
}

// Get query value selector parameters.
func (query *Query) Parameters() string {
	var c_string C.z_view_string_t
	C.z_query_parameters(C.z_query_loan(query.query), &c_string)
	loanedString := C.z_view_string_loan(&c_string)
	return C.GoStringN(C.z_string_data(loanedString), C.int(C.z_string_len(loanedString)))
}

func newQueryFromC(c_query *C.z_loaned_query_t) Query {
	var cq C.z_owned_query_t
	C.z_query_clone(&cq, c_query)
	return Query{query: &cq}
}

// Options passed to Query Reply operation.
type QueryReplyOptions struct {
	Encoding          option.Option[Encoding]          // The encoding of the reply payload.
	Attachement       option.Option[ZBytes]            // The attachment to attach to this reply.
	TimeStamp         option.Option[TimeStamp]         // The timestamp of the reply.
	CongestionControl option.Option[CongestionControl] // The congestion control to apply when routing the reply.
	Priority          option.Option[Priority]          // The priority of the reply.
	IsExpress         bool                             // If set to ``true``, this reply message will not be batched. This usually has a positive impact on latency but negative impact on throughput.
}

func (opts *QueryReplyOptions) toCOpts(pinner *runtime.Pinner) C.z_query_reply_options_t {
	var cOpts C.z_query_reply_options_t
	C.z_query_reply_options_default(&cOpts)
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

// Options passed to Query ReplyDel operation.
type QueryReplyDelOptions struct {
	Attachement       option.Option[ZBytes]            // The attachment to attach to this reply.
	TimeStamp         option.Option[TimeStamp]         // The timestamp of the reply.
	CongestionControl option.Option[CongestionControl] // The congestion control to apply when routing the reply.
	Priority          option.Option[Priority]          // The priority of the reply.
	IsExpress         bool                             // If set to ``true``, this reply message will not be batched. This usually has a positive impact on latency but negative impact on throughput.
}

func (opts *QueryReplyDelOptions) toCOpts(pinner *runtime.Pinner) C.z_query_reply_del_options_t {
	var cOpts C.z_query_reply_del_options_t
	C.z_query_reply_del_options_default(&cOpts)
	if opts.Attachement.IsSome() {
		cAttachment := opts.Attachement.Unwrap().toC()
		pinner.Pin(&cAttachment)
		cOpts.attachment = C.z_bytes_move(&cAttachment)
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

// Options passed to Query Reply operation.
type QueryReplyErrOptions struct {
	Encoding option.Option[Encoding] // The encoding of the reply payload.
}

func (opts *QueryReplyErrOptions) toCOpts(pinner *runtime.Pinner) C.z_query_reply_err_options_t {
	var cOpts C.z_query_reply_err_options_t
	C.z_query_reply_err_options_default(&cOpts)
	if opts.Encoding.IsSome() {
		cEncoding := opts.Encoding.Unwrap().toC()
		pinner.Pin(&cEncoding)
		cOpts.encoding = C.z_encoding_move(&cEncoding)
	}
	return cOpts
}

// Send a reply to the query.
//
// This function can be called multiple times to send multiple replies to a query. The reply
// will be considered complete when Drop is called.
func (query *Query) Reply(keyexpr KeyExpr, payload ZBytes, options *QueryReplyOptions) error {
	res := int8(0)
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toC(&pinner)
	cPayload := payload.toC()
	if options == nil {
		res = int8(C.z_query_reply(C.z_query_loan(query.query), C.z_view_keyexpr_loan(&cKeyexpr), C.z_bytes_move(&cPayload), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_query_reply(C.z_query_loan(query.query), C.z_view_keyexpr_loan(&cKeyexpr), C.z_bytes_move(&cPayload), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to send reply")
}

// Send a delete reply to the query.
//
// This function can be called multiple times to send multiple replies to a query. The reply
// will be considered complete when Drop is called.
func (query *Query) ReplyDel(keyexpr KeyExpr, options *QueryReplyDelOptions) error {
	res := int8(0)
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toC(&pinner)
	if options == nil {
		res = int8(C.z_query_reply_del(C.z_query_loan(query.query), C.z_view_keyexpr_loan(&cKeyexpr), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_query_reply_del(C.z_query_loan(query.query), C.z_view_keyexpr_loan(&cKeyexpr), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to send reply del")
}

// Send a error reply to the query.
//
// This function can be called multiple times to send multiple replies to a query. The reply
// will be considered complete when Drop is called.
func (query *Query) ReplyErr(payload ZBytes, options *QueryReplyErrOptions) error {
	res := int8(0)
	pinner := runtime.Pinner{}
	cPayload := payload.toC()
	if options == nil {
		res = int8(C.z_query_reply_err(C.z_query_loan(query.query), C.z_bytes_move(&cPayload), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_query_reply_err(C.z_query_loan(query.query), C.z_bytes_move(&cPayload), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to send reply error")
}
