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

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Options passed to querier declaration.
type QuerierOptions struct {
	Target             option.Option[QueryTarget]        // The Queryables that should be target of the querier queries.
	Consolidataion     option.Option[QueryConsolidation] // The replies consolidation strategy to apply on replies to the querier queries.
	CongestionControl  option.Option[CongestionControl]  // The congestion control to apply when routing the query.
	Priority           option.Option[Priority]           // The priority of the querier queries.
	IsExpress          bool                              // If set to ``true``, the querier queries will not be batched. This usually has a positive impact on latency but negative impact on throughput.
	TimeoutMs          uint64                            // The timeout for the querier queries in milliseconds. 0 means default query timeout from zenoh configuration.
	AllowedDestination option.Option[Locality]           // Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release. The allowed destination for the querier queries.
	AcceptReplies      option.Option[ReplyKeyexpr]       // This API has been marked as unstable: it works as advertised, but it may be changed in a future release. The accepted replies for the querier queries.
}

func (opts *QuerierOptions) toCOpts(_pinner *runtime.Pinner) C.z_querier_options_t {
	var cOpts C.z_querier_options_t
	C.z_querier_options_default(&cOpts)
	if opts.Priority.IsSome() {
		cOpts.priority = uint32(C.z_priority_t(opts.Priority.Unwrap()))
	}
	if opts.CongestionControl.IsSome() {
		cOpts.congestion_control = uint32(opts.CongestionControl.Unwrap())
	}
	cOpts.is_express = C.bool(opts.IsExpress)
	if opts.Target.IsSome() {
		cOpts.target = uint32(opts.Target.Unwrap())
	}
	if opts.Consolidataion.IsSome() {
		cOpts.consolidation.mode = int32(opts.Consolidataion.Unwrap().mode)
	}
	cOpts.timeout_ms = C.uint64_t(opts.TimeoutMs)
	if opts.AllowedDestination.IsSome() {
		cOpts.allowed_destination = uint32(opts.AllowedDestination.Unwrap())
	}
	if opts.AcceptReplies.IsSome() {
		cOpts.accept_replies = uint32(opts.AcceptReplies.Unwrap())
	}
	return cOpts
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A Zenoh querier.
//
// Sends queries to matching queryables.
type Querier struct {
	querier *C.z_owned_querier_t
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Options passed to [Querier.Get] operation.
type QuerierGetOptions struct {
	Payload     option.Option[ZBytes]   // An optional payload to attach to the query.
	Encoding    option.Option[Encoding] // An optional encoding of the query payload and or attachment.
	Attachement option.Option[ZBytes]   // The attachment to attach to the query.
}

func (opts *QuerierGetOptions) toCOpts(pinner *runtime.Pinner) (C.z_querier_get_options_t, *C.zc_cgo_bytes_data_t, *C.zc_internal_encoding_data_t, *C.zc_cgo_bytes_data_t) {
	var cOpts C.z_querier_get_options_t
	C.z_querier_get_options_default(&cOpts)
	payload := (*C.zc_cgo_bytes_data_t)(nil)
	encoding := (*C.zc_internal_encoding_data_t)(nil)
	attachment := (*C.zc_cgo_bytes_data_t)(nil)
	if opts.Payload.IsSome() {
		cPayloadData := opts.Payload.Unwrap().toCData(pinner)
		payload = &cPayloadData
	}
	if opts.Attachement.IsSome() {
		cAttachmentData := opts.Attachement.Unwrap().toCData(pinner)
		attachment = &cAttachmentData
	}
	if opts.Encoding.IsSome() {
		cEncoding := opts.Encoding.Unwrap().toCData(pinner)
		encoding = &cEncoding
	}

	return cOpts, payload, encoding, attachment
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Get the key expression of the querier.
func (querier *Querier) KeyExpr() KeyExpr {
	return newKeyExprFromC(C.zc_cgo_keyexpr_get_data(C.z_querier_keyexpr(C.z_querier_loan(querier.querier))))
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Construct a querier for the given key expression.
// Querier MUST be explicitly destroyed using [Querier.Drop] once it is no longer needed.
func (session *Session) DeclareQuerier(keyexpr KeyExpr, options *QuerierOptions) (Querier, error) {
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toC(&pinner)
	res := int8(0)
	var cQuerier C.z_owned_querier_t
	if options == nil {
		res = int8(C.z_declare_querier(C.z_session_loan(session.session), &cQuerier, C.z_view_keyexpr_loan(&cKeyexpr), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_declare_querier(C.z_session_loan(session.session), &cQuerier, C.z_view_keyexpr_loan(&cKeyexpr), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return Querier{querier: &cQuerier}, nil
	}
	return Querier{}, NewZError(res, "Failed to declare Querier")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Destroy the querier.
func (querier *Querier) Drop() {
	C.z_querier_drop(C.z_querier_move(querier.querier))
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Query data from the matching queryables in the system.
// Replies are provided through a callback function, if handler is a [Closure], through returned receiver if it is a [RingChannel] or a [FifoChannel].
func (querier *Querier) Get(parameters string, handler Handler[Reply], get_options *QuerierGetOptions) (<-chan Reply, error) {
	callback, drop, channel := handler.ToCbDropHandler()
	closure := newClosure(callback, drop)
	pinner := runtime.Pinner{}
	cParams := (*C.char)(nil)
	if len(parameters) != 0 {
		cParams = C.CString(parameters)
		defer C.free(unsafe.Pointer(cParams))
	}
	res := int8(0)
	if get_options == nil {
		res = int8(C.zc_cgo_querier_get(querier.querier, cParams, unsafe.Pointer(closure), nil, nil, nil, nil))
	} else {
		cOpts, payload, encoding, attachment := get_options.toCOpts(&pinner)
		res = int8(C.zc_cgo_querier_get(querier.querier, cParams, unsafe.Pointer(closure), &cOpts, payload, encoding, attachment))
	}
	pinner.Unpin()

	if res == 0 {
		return channel, nil
	}
	return nil, NewZError(res, "Failed to perform Querier Get operation")
}
