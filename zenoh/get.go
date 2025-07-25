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
// static const z_consolidation_mode_t CGO_Z_CONSOLIDATION_MODE_DEFAULT = Z_CONSOLIDATION_MODE_DEFAULT;
// static const z_query_target_t CGO_Z_QUERY_TARGET_DEFAULT = Z_QUERY_TARGET_DEFAULT;
import "C"
import (
	"runtime"
	"unsafe"

	"github.com/BooleanCat/option"
)

// The Queryables that should be target of a get.
type QueryTarget int

const (
	QueryTargetBestMatching QueryTarget = C.Z_QUERY_TARGET_BEST_MATCHING // The nearest complete queryable if any else all matching queryables.
	QueryTargetAll          QueryTarget = C.Z_QUERY_TARGET_ALL           // All matching queryables.
	QueryTargetAllComplete  QueryTarget = C.Z_QUERY_TARGET_ALL_COMPLETE  // All complete queryables.
	QueryTargetDefault      QueryTarget = C.CGO_Z_QUERY_TARGET_DEFAULT
)

// The Query consolidation mode.
type ConsolidationMode int

const (
	// Let Zenoh decide the best consolidation mode depending on the query selector.
	// If the selector contains time range properties, consolidation mode `NONE` is used.
	// Otherwise the `LATEST` consolidation mode is used.
	ConsolidationModeAuto ConsolidationMode = C.Z_CONSOLIDATION_MODE_AUTO
	// No consolidation is applied. Replies may come in any order and any number.
	ConsolidationModeNone ConsolidationMode = C.Z_CONSOLIDATION_MODE_NONE
	// It guarantees that any reply for a given key expression will be monotonic in time
	// w.r.t. the previous received replies for the same key expression. I.e., for the same key expression multiple
	// replies may be received. It is guaranteed that two replies received at t1 and t2 will have timestamp
	// ts2 > ts1. It optimizes latency.
	ConsolidationModeMonothonic ConsolidationMode = C.Z_CONSOLIDATION_MODE_MONOTONIC
	// It guarantees unicity of replies for the same key expression.
	// It optimizes bandwidth.
	ConsolidationModeLatest ConsolidationMode = C.Z_CONSOLIDATION_MODE_LATEST

	ConsolidationModeLatestDefault ConsolidationMode = C.CGO_Z_CONSOLIDATION_MODE_DEFAULT
)

// The replies consolidation strategy to apply on replies to a get.
type QueryConsolidation struct {
	mode ConsolidationMode
}

// Construct QueryConsolidation from [ConsolidationMode].
func NewQueryConsolidataion(mode ConsolidationMode) QueryConsolidation {
	return QueryConsolidation{mode: mode}
}

// Options passed to Session Get operation.
type GetOptions struct {
	Target             option.Option[QueryTarget]        // The Queryables that should be target of the query.
	Consolidataion     option.Option[QueryConsolidation] // The replies consolidation strategy to apply on replies to the query.
	Payload            option.Option[ZBytes]             // An optional payload to attach to the query.
	Encoding           option.Option[Encoding]           // An optional encoding of the query payload and/or attachment.
	Attachement        option.Option[ZBytes]             // The attachment to attach to the query.
	CongestionControl  option.Option[CongestionControl]  // The congestion control to apply when routing the query.
	Priority           option.Option[Priority]           // The priority of the query.
	IsExpress          bool                              // If set to ``true``, this query will not be batched. This usually has a positive impact on latency but negative impact on throughput.
	TimeoutMs          uint64                            // The timeout for the query reply in milliseconds. 0 means default query timeout from zenoh configuration.
	AllowedDestination option.Option[Locality]           // Restrict the queryables which receive the query to the ones with compatible AllowedOrigin.
	AcceptReplies      option.Option[ReplyKeyexpr]       // Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release. The accepted replies for the query.
}

func (opts *GetOptions) toCOpts(pinner *runtime.Pinner) (C.z_get_options_t, *C.zc_cgo_bytes_data_t, *C.zc_internal_encoding_data_t, *C.zc_cgo_bytes_data_t) {
	var cOpts C.z_get_options_t
	C.z_get_options_default(&cOpts)
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
	return cOpts, payload, encoding, attachment
}

//export zenohGetCallbackData
func zenohGetCallbackData(reply C.zc_cgo_reply_data_t, context unsafe.Pointer) {
	(*closureContext[Reply])(context).call(newReplyFromC(reply))
}

//export zenohGetDrop
func zenohGetDrop(context unsafe.Pointer) {
	(*closureContext[Reply])(context).drop()
}

// Query data from the matching queryables in the system.
// Replies are provided through a callback function, if handler is a [Closure], through returned receiver if it is a [RingChannel] or a [FifoChannel].
func (session *Session) Get(keyexpr KeyExpr, parameters string, handler Handler[Reply], get_options *GetOptions) (<-chan Reply, error) {
	callback, drop, channel := handler.ToCbDropHandler()
	closure := newClosure(callback, drop)
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toCData(&pinner)
	cParams := (*C.char)(nil)
	if len(parameters) != 0 {
		cParams = C.CString(parameters)
		defer C.free(unsafe.Pointer(cParams))
	}
	res := int8(0)
	if get_options == nil {
		res = int8(C.zc_cgo_get(session.session, cKeyexpr, cParams, unsafe.Pointer(closure), nil, nil, nil, nil))
	} else {
		cOpts, payload, encoding, attachment := get_options.toCOpts(&pinner)
		res = int8(C.zc_cgo_get(session.session, cKeyexpr, cParams, unsafe.Pointer(closure), &cOpts, payload, encoding, attachment))
	}
	pinner.Unpin()

	if res == 0 {
		return channel, nil
	}
	return nil, NewZError(res, "Failed to perform Get operation")
}
