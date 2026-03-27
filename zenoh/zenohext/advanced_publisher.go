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

// #include "zenoh.h"
// #include "zenoh_ext_cgo.h"
import "C"
import (
	"runtime"
	"unsafe"
	"zenoh-go/zenoh"
	"zenoh-go/zenoh/inner"

	"github.com/BooleanCat/option"
)

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A Zenoh advanced publisher.
//
// In addition to publishing the data, it also maintains a cache, allowing advanced subscribers to recover missed or historical samples.
// Advanced publishers are automatically undeclared when dropped.
type AdvancedPublisher struct {
	publisher *C.ze_owned_advanced_publisher_t
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Options for the advanced publisher cache.
// The cache allows advanced subscribers to recover history and/or lost samples.
// Pass as [option.Some] in [AdvancedPublisherOptions.Cache] to enable the cache.
type AdvancedPublisherCacheOptions struct {
	// Number of samples to keep for each resource.
	MaxSamples uint
	// The congestion control to apply to replies.
	CongestionControl option.Option[zenoh.CongestionControl]
	// The priority of replies.
	Priority option.Option[zenoh.Priority]
	// If true, replies will not be batched. Positive impact on latency but negative on throughput.
	IsExpress bool
}

type heartbeatPeriodic struct {
	period uint64
}
type heartbeatSporadic struct {
	PeriodMs uint64 // Period of heartbeats in milliseconds.
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Heartbeat mode for advanced publisher sample miss detection. Pass as [option.Some] in [AdvancedPublisherSampleMissDetectionOptions.HeartbeatMode] to enable miss detection.
type HeartbeatMode struct {
	mode interface{}
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Allow last sample miss detection through periodic heartbeat.
// Periodically send the last published Sample's sequence number to allow last sample miss detection.
func HeartbeatModePeriodic(periodMs uint64) HeartbeatMode {
	return HeartbeatMode{
		mode: heartbeatPeriodic{period: max(1, periodMs)},
	}
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Allow last sample miss detection through sporadic heartbeat.
// Each period, the last published Sample's sequence number is sent with [zenoh.CongestionControlBlock] congestion control,
// but only if it changed since last period.
func HeartbeatModeSporadic(periodMs uint64) HeartbeatMode {
	return HeartbeatMode{
		mode: heartbeatSporadic{PeriodMs: max(1, periodMs)},
	}
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Disable last sample miss detection through heartbeat.
func HeartbeatModeNone() HeartbeatMode {
	return HeartbeatMode{}
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Options for sample miss detection on the advanced publisher.
// Pass as [option.Some] in [AdvancedPublisherOptions.SampleMissDetection] to enable miss detection.
type AdvancedPublisherSampleMissDetectionOptions struct {
	// Can be set to either [HeartbeatModePeriodic], [HeartbeatModeSporadic] or [HeartbeatModeNone].
	HeartbeatMode HeartbeatMode
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Options passed to [SessionExt.DeclareAdvancedPublisher].
type AdvancedPublisherOptions struct {
	// Base publisher options (encoding, congestion control, priority, etc.).
	Encoding           option.Option[zenoh.Encoding]
	CongestionControl  option.Option[zenoh.CongestionControl]
	Priority           option.Option[zenoh.Priority]
	IsExpress          bool
	Reliability        option.Option[zenoh.Reliability]
	AllowedDestination option.Option[zenoh.Locality]
	// Publisher cache settings. Set to option.Some to enable the cache.
	Cache option.Option[AdvancedPublisherCacheOptions]
	// Settings for sample miss detection. Set to option.Some to enable miss detection.
	SampleMissDetection option.Option[AdvancedPublisherSampleMissDetectionOptions]
	// Allow this publisher to be detected through liveliness.
	PublisherDetection bool
	// An optional key expression to be added to the liveliness token key expression.
	// Can be used to convey metadata.
	PublisherDetectionMetadata option.Option[zenoh.KeyExpr]
}

func (opts *AdvancedPublisherOptions) toCOpts(pinner *runtime.Pinner) C.ze_advanced_publisher_options_t {
	var cOpts C.ze_advanced_publisher_options_t
	C.ze_advanced_publisher_options_default(&cOpts)

	// Base publisher options
	if opts.Encoding.IsSome() {
		cEncoding := (*C.z_owned_encoding_t)(encodingToUnsafeCPtr(opts.Encoding.Unwrap()))
		pinner.Pin(cEncoding)
		cOpts.publisher_options.encoding = C.z_encoding_move(cEncoding)
	}
	if opts.Priority.IsSome() {
		cOpts.publisher_options.priority = uint32(C.z_priority_t(opts.Priority.Unwrap()))
	}
	if opts.CongestionControl.IsSome() {
		cOpts.publisher_options.congestion_control = uint32(opts.CongestionControl.Unwrap())
	}
	if opts.Reliability.IsSome() {
		cOpts.publisher_options.reliability = uint32(opts.Reliability.Unwrap())
	}
	if opts.AllowedDestination.IsSome() {
		cOpts.publisher_options.allowed_destination = uint32(opts.AllowedDestination.Unwrap())
	}
	cOpts.publisher_options.is_express = C.bool(opts.IsExpress)

	// Cache options
	if opts.Cache.IsSome() {
		cache := opts.Cache.Unwrap()
		cOpts.cache.is_enabled = C.bool(true)
		cOpts.cache.max_samples = C.size_t(cache.MaxSamples)
		if cache.CongestionControl.IsSome() {
			cOpts.cache.congestion_control = uint32(cache.CongestionControl.Unwrap())
		}
		if cache.Priority.IsSome() {
			cOpts.cache.priority = uint32(C.z_priority_t(cache.Priority.Unwrap()))
		}
		cOpts.cache.is_express = C.bool(cache.IsExpress)
	}

	// Sample miss detection options
	if opts.SampleMissDetection.IsSome() {
		smd := opts.SampleMissDetection.Unwrap()
		cOpts.sample_miss_detection.is_enabled = C.bool(true)
		switch mode := smd.HeartbeatMode.mode.(type) {
		case heartbeatPeriodic:
			cOpts.sample_miss_detection.heartbeat_mode = C.ZE_ADVANCED_PUBLISHER_HEARTBEAT_MODE_PERIODIC
			cOpts.sample_miss_detection.heartbeat_period_ms = C.uint64_t(mode.period)
		case heartbeatSporadic:
			cOpts.sample_miss_detection.heartbeat_mode = C.ZE_ADVANCED_PUBLISHER_HEARTBEAT_MODE_SPORADIC
			cOpts.sample_miss_detection.heartbeat_period_ms = C.uint64_t(mode.PeriodMs)
		default:
			cOpts.sample_miss_detection.heartbeat_mode = C.ZE_ADVANCED_PUBLISHER_HEARTBEAT_MODE_NONE
		}
	}

	// Publisher detection
	cOpts.publisher_detection = C.bool(opts.PublisherDetection)
	if opts.PublisherDetectionMetadata.IsSome() {
		ke := opts.PublisherDetectionMetadata.Unwrap()
		cKeyexpr := (*C.z_view_keyexpr_t)(keyExprToUnsafeCPtr(&ke, pinner))
		cOpts.publisher_detection_metadata = C.z_view_keyexpr_loan(cKeyexpr)
	}

	return cOpts
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Options passed to [AdvancedPublisher.Put].
type AdvancedPublisherPutOptions struct {
	Encoding   option.Option[zenoh.Encoding]  // The encoding of the publication.
	Attachment option.Option[zenoh.ZBytes]    // The attachment to attach to the publication.
	TimeStamp  option.Option[zenoh.TimeStamp] // The timestamp of the publication.
}

func (opts *AdvancedPublisherPutOptions) toCOpts(pinner *runtime.Pinner) (C.ze_advanced_publisher_put_options_t, *C.zc_internal_encoding_data_t, *C.zc_cgo_bytes_data_t) {
	var cOpts C.ze_advanced_publisher_put_options_t
	C.ze_advanced_publisher_put_options_default(&cOpts)
	encoding := (*C.zc_internal_encoding_data_t)(nil)
	attachment := (*C.zc_cgo_bytes_data_t)(nil)
	if opts.Attachment.IsSome() {
		attachment = (*C.zc_cgo_bytes_data_t)(zbytesToUnsafeCDataPtr(opts.Attachment.Unwrap(), pinner))
	}
	if opts.Encoding.IsSome() {
		encoding = (*C.zc_internal_encoding_data_t)(encodingToUnsafeCDataPtr(opts.Encoding.Unwrap(), pinner))
	}
	if opts.TimeStamp.IsSome() {
		cOpts.put_options.timestamp = (*C.z_timestamp_t)(timeStampToUnsafeCPtr(opts.TimeStamp.Unwrap()))
	}
	return cOpts, encoding, attachment
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Options passed to [AdvancedPublisher.Delete].
type AdvancedPublisherDeleteOptions struct {
	TimeStamp option.Option[zenoh.TimeStamp] // The timestamp of the delete message.
}

func (opts *AdvancedPublisherDeleteOptions) toCOpts(_ *runtime.Pinner) C.ze_advanced_publisher_delete_options_t {
	var cOpts C.ze_advanced_publisher_delete_options_t
	C.ze_advanced_publisher_delete_options_default(&cOpts)
	if opts.TimeStamp.IsSome() {
		cOpts.delete_options.timestamp = (*C.z_timestamp_t)(timeStampToUnsafeCPtr(opts.TimeStamp.Unwrap()))
	}
	return cOpts
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Undeclare and destroy the advanced publisher.
func (publisher *AdvancedPublisher) Undeclare() error {
	res := int8(C.ze_undeclare_advanced_publisher(C.ze_advanced_publisher_move(publisher.publisher)))
	if res == 0 {
		return nil
	}
	return zenoh.NewZError(res, "Failed to undeclare AdvancedPublisher")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Destroy the advanced publisher.
// This is equivalent to calling [AdvancedPublisher.Undeclare] and discarding its return value.
func (publisher *AdvancedPublisher) Drop() {
	C.ze_advanced_publisher_drop(C.ze_advanced_publisher_move(publisher.publisher))
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Publish data onto the advanced publisher's key expression.
func (publisher *AdvancedPublisher) Put(payload zenoh.ZBytes, options *AdvancedPublisherPutOptions) error {
	pinner := runtime.Pinner{}
	cPayload := (*C.zc_cgo_bytes_data_t)(zbytesToUnsafeCDataPtr(payload, &pinner))
	res := int8(0)
	if options == nil {
		res = int8(C.zc_cgo_advanced_publisher_put(publisher.publisher, cPayload, nil, nil, nil))
	} else {
		cOpts, encoding, attachment := options.toCOpts(&pinner)
		res = int8(C.zc_cgo_advanced_publisher_put(publisher.publisher, cPayload, &cOpts, encoding, attachment))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return zenoh.NewZError(res, "Failed to perform AdvancedPublisher Put operation")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Publish a DELETE message onto the advanced publisher's key expression.
func (publisher *AdvancedPublisher) Delete(options *AdvancedPublisherDeleteOptions) error {
	pinner := runtime.Pinner{}
	res := int8(0)
	if options == nil {
		res = int8(C.zc_cgo_advanced_publisher_delete(publisher.publisher, nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.zc_cgo_advanced_publisher_delete(publisher.publisher, &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return zenoh.NewZError(res, "Failed to perform AdvancedPublisher Delete operation")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Get the key expression of the advanced publisher.
func (publisher *AdvancedPublisher) KeyExpr() zenoh.KeyExpr {
	ke := C.zc_cgo_keyexpr_get_data(C.ze_advanced_publisher_keyexpr(C.ze_advanced_publisher_loan(publisher.publisher)))
	return newKeyExprFromUnsafeCDataPtr((unsafe.Pointer)(&ke))
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Get the matching status of the advanced publisher - i.e. if there are any subscribers matching its key expression.
func (publisher *AdvancedPublisher) GetMatchingStatus() (zenoh.MatchingStatus, error) {
	var status C.z_matching_status_t
	res := int8(C.ze_advanced_publisher_get_matching_status(C.ze_advanced_publisher_loan(publisher.publisher), &status))
	if res == 0 {
		return zenoh.MatchingStatus{Matching: bool(status.matching)}, nil
	}
	return zenoh.MatchingStatus{}, zenoh.NewZError(res, "Failed to retrieve AdvancedPublisher Matching Status")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Declare a matching listener for notifying when subscribers matching with this advanced publisher connect or disconnect.
// Matching listener MUST be explicitly destroyed using [zenoh.MatchingListener.Undeclare] or [zenoh.MatchingListener.Drop] once it is no longer needed.
func (publisher *AdvancedPublisher) DeclareMatchingListener(handler zenoh.Handler[zenoh.MatchingStatus]) (zenoh.MatchingListener, error) {
	callback, drop, recv := handler.ToCbDropHandler()
	listenerClosure := inner.NewClosure(callback, drop)
	var cClosure C.z_owned_closure_matching_status_t
	C.z_closure_matching_status(&cClosure, (*[0]byte)(C.zenohMatchingListenerCallback), (*[0]byte)(C.zenohMatchingListenerDrop), unsafe.Pointer(listenerClosure))

	var cMatchingListener C.z_owned_matching_listener_t
	res := int8(C.ze_advanced_publisher_declare_matching_listener(C.ze_advanced_publisher_loan(publisher.publisher), &cMatchingListener, C.z_closure_matching_status_move(&cClosure)))

	if res == 0 {
		return matchingListenerFromUnsafeCPtrAndReceiver(unsafe.Pointer(&cMatchingListener), recv), nil
	}
	return zenoh.MatchingListener{}, zenoh.NewZError(res, "Failed to declare Matching Listener for AdvancedPublisher")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Declare a background matching listener for this advanced publisher.
// The callback will be run in the background until the corresponding publisher is dropped.
func (publisher *AdvancedPublisher) DeclareBackgroundMatchingListener(closure zenoh.Closure[zenoh.MatchingStatus]) error {
	listenerClosure := inner.NewClosure(closure.Call, closure.Drop)
	var cClosure C.z_owned_closure_matching_status_t
	C.z_closure_matching_status(&cClosure, (*[0]byte)(C.zenohMatchingListenerCallback), (*[0]byte)(C.zenohMatchingListenerDrop), unsafe.Pointer(listenerClosure))

	res := int8(C.ze_advanced_publisher_declare_background_matching_listener(C.ze_advanced_publisher_loan(publisher.publisher), C.z_closure_matching_status_move(&cClosure)))

	if res == 0 {
		return nil
	}
	return zenoh.NewZError(res, "Failed to declare background Matching Listener for AdvancedPublisher")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Construct and declare an advanced publisher for the given key expression.
//
// Data can be put and deleted via [AdvancedPublisher.Put] and [AdvancedPublisher.Delete].
// The publisher MUST be explicitly destroyed using [AdvancedPublisher.Undeclare] or [AdvancedPublisher.Drop] once it is no longer needed.
func (session *SessionExt) DeclareAdvancedPublisher(keyexpr zenoh.KeyExpr, options *AdvancedPublisherOptions) (AdvancedPublisher, error) {
	pinner := runtime.Pinner{}
	cKeyexpr := (*C.z_view_keyexpr_t)(keyExprToUnsafeCPtr(&keyexpr, &pinner))
	var cPublisher C.ze_owned_advanced_publisher_t
	res := int8(0)

	if options == nil {
		res = int8(C.ze_declare_advanced_publisher(C.z_session_loan(session.getInner()), &cPublisher, C.z_view_keyexpr_loan(cKeyexpr), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.ze_declare_advanced_publisher(C.z_session_loan(session.getInner()), &cPublisher, C.z_view_keyexpr_loan(cKeyexpr), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return AdvancedPublisher{publisher: &cPublisher}, nil
	}
	return AdvancedPublisher{}, zenoh.NewZError(res, "Failed to declare AdvancedPublisher")
}
