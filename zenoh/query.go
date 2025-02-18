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

	"github.com/BooleanCat/option"
)

// A Zenoh query received by a queryable.
type Query struct {
	keyexpr    KeyExpr
	payload    option.Option[ZBytes]
	encoding   option.Option[Encoding]
	attachment option.Option[ZBytes]
	parameters string
	query      *C.z_owned_query_t
}

// Finalizes and destroys the query. This MUST be always called by user once all replies are provided.
func (query *Query) Drop() {
	C.zc_cgo_query_drop(query.query)
}

// Return the key expression of the query.
func (query *Query) KeyExpr() KeyExpr {
	return query.keyexpr
}

// Return query payload data if there is any.
func (query *Query) Payload() option.Option[ZBytes] {
	return query.payload
}

// Return the encoding associated with the query data, if there is any.
func (query *Query) Encoding() option.Option[Encoding] {
	return query.encoding
}

// Return query attachment if there is any.
func (query *Query) Attachement() option.Option[ZBytes] {
	return query.attachment
}

// Get query value selector parameters.
func (query *Query) Parameters() string {
	return query.parameters
}

func newQueryFromC(cQueryData C.zc_cgo_query_data_t) Query {
	var q Query
	q.keyexpr = newKeyExprFromC(cQueryData.keyexpr)
	q.parameters = C.GoStringN(cQueryData.params.str_ptr, C.int(cQueryData.params.len))
	if cQueryData.has_payload {
		q.payload = option.Some(newZBytesFromC(cQueryData.payload))
	}
	if cQueryData.has_attachment {
		q.attachment = option.Some(newZBytesFromC(cQueryData.attachment))
	}
	if cQueryData.has_encoding {
		q.encoding = option.Some(newEncodingFromC(cQueryData.encoding))
	}
	q.query = &cQueryData.query
	return q
}

// Options passed to [Query.Reply] operation.
type QueryReplyOptions struct {
	Encoding          option.Option[Encoding]          // The encoding of the reply payload.
	Attachement       option.Option[ZBytes]            // The attachment to attach to this reply.
	TimeStamp         option.Option[TimeStamp]         // The timestamp of the reply.
	CongestionControl option.Option[CongestionControl] // The congestion control to apply when routing the reply.
	Priority          option.Option[Priority]          // The priority of the reply.
	IsExpress         bool                             // If set to ``true``, this reply message will not be batched. This usually has a positive impact on latency but negative impact on throughput.
}

func (opts *QueryReplyOptions) toCOpts(pinner *runtime.Pinner) (C.z_query_reply_options_t, *C.zc_internal_encoding_data_t, *C.zc_cgo_bytes_data_t) {
	var cOpts C.z_query_reply_options_t
	C.z_query_reply_options_default(&cOpts)
	encoding := (*C.zc_internal_encoding_data_t)(nil)
	attachment := (*C.zc_cgo_bytes_data_t)(nil)
	if opts.Attachement.IsSome() {
		cAttachment := opts.Attachement.Unwrap().toCData(pinner)
		attachment = &cAttachment
	}
	if opts.Encoding.IsSome() {
		cEncoding := opts.Encoding.Unwrap().toCData(pinner)
		encoding = &cEncoding
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
	return cOpts, encoding, attachment
}

// Options passed to [Query.ReplyDel] operation.
type QueryReplyDelOptions struct {
	Attachement       option.Option[ZBytes]            // The attachment to attach to this reply.
	TimeStamp         option.Option[TimeStamp]         // The timestamp of the reply.
	CongestionControl option.Option[CongestionControl] // The congestion control to apply when routing the reply.
	Priority          option.Option[Priority]          // The priority of the reply.
	IsExpress         bool                             // If set to ``true``, this reply message will not be batched. This usually has a positive impact on latency but negative impact on throughput.
}

func (opts *QueryReplyDelOptions) toCOpts(pinner *runtime.Pinner) (C.z_query_reply_del_options_t, *C.zc_cgo_bytes_data_t) {
	var cOpts C.z_query_reply_del_options_t
	C.z_query_reply_del_options_default(&cOpts)
	attachment := (*C.zc_cgo_bytes_data_t)(nil)
	if opts.Attachement.IsSome() {
		cAttachment := opts.Attachement.Unwrap().toCData(pinner)
		attachment = &cAttachment
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
	return cOpts, attachment
}

// Options passed to [Query.ReplyErr] operation.
type QueryReplyErrOptions struct {
	Encoding option.Option[Encoding] // The encoding of the reply payload.
}

func (opts *QueryReplyErrOptions) toCOpts(pinner *runtime.Pinner) (C.z_query_reply_err_options_t, *C.zc_internal_encoding_data_t) {
	var cOpts C.z_query_reply_err_options_t
	C.z_query_reply_err_options_default(&cOpts)
	encoding := (*C.zc_internal_encoding_data_t)(nil)
	if opts.Encoding.IsSome() {
		cEncoding := opts.Encoding.Unwrap().toCData(pinner)
		encoding = &cEncoding
	}
	return cOpts, encoding
}

// Send a reply to the query.
//
// This function can be called multiple times to send multiple replies to a query. The reply
// will be considered complete when Drop is called.
func (query *Query) Reply(keyexpr KeyExpr, payload ZBytes, options *QueryReplyOptions) error {
	res := int8(0)
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toCData(&pinner)
	cPayload := payload.toCData(&pinner)
	if options == nil {
		res = int8(C.zc_cgo_query_reply(query.query, cKeyexpr, cPayload, nil, nil, nil))
	} else {
		cOpts, encoding, attachment := options.toCOpts(&pinner)
		res = int8(C.zc_cgo_query_reply(query.query, cKeyexpr, cPayload, &cOpts, encoding, attachment))
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
	cKeyexpr := keyexpr.toCData(&pinner)
	if options == nil {
		res = int8(C.zc_cgo_query_reply_del(query.query, cKeyexpr, nil, nil))
	} else {
		cOpts, attachment := options.toCOpts(&pinner)
		res = int8(C.zc_cgo_query_reply_del(query.query, cKeyexpr, &cOpts, attachment))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to send reply del")
}

// Send an error reply to the query.
//
// This function can be called multiple times to send multiple replies to a query. The reply
// will be considered complete when Drop is called.
func (query *Query) ReplyErr(payload ZBytes, options *QueryReplyErrOptions) error {
	res := int8(0)
	pinner := runtime.Pinner{}
	cPayload := payload.toCData(&pinner)
	if options == nil {
		res = int8(C.zc_cgo_query_reply_err(query.query, cPayload, nil, nil))
	} else {
		cOpts, encoding := options.toCOpts(&pinner)
		res = int8(C.zc_cgo_query_reply_err(query.query, cPayload, &cOpts, encoding))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return NewZError(res, "Failed to send reply error")
}
