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
// A struct that represents missed samples from an advanced publisher.
type Miss struct {
	// The source of the missed samples (entity global ID).
	Source zenoh.EntityGlobalId
	// The number of missed samples.
	Nb uint32
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A Zenoh sample miss listener.
// Sends notifications when an advanced subscriber detects missed samples.
// Dropping the corresponding subscriber also drops the listener.
type SampleMissListener struct {
	listener *C.ze_owned_sample_miss_listener_t
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Destroy the sample miss listener.
func (listener *SampleMissListener) Drop() {
	C.ze_sample_miss_listener_drop(C.ze_sample_miss_listener_move(listener.listener))
}

//export zenohMissListenerCallback
func zenohMissListenerCallback(miss *C.zc_cgo_const_miss_t, context unsafe.Pointer) {
	m := Miss{
		Source: newEntityGlobalIdFromUnsafeCPtr(unsafe.Pointer(&miss.source)),
		Nb:     uint32(miss.nb),
	}
	(*inner.ClosureContext[Miss])(context).Call(m)
}

//export zenohMissListenerDrop
func zenohMissListenerDrop(context unsafe.Pointer) {
	(*inner.ClosureContext[Miss])(context).Drop()
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A Zenoh advanced subscriber.
//
// In addition to receiving subscribed data, it can also detect missed samples and/or automatically recover them.
// Advanced subscribers are automatically undeclared when dropped.
type AdvancedSubscriber struct {
	subscriber *C.ze_owned_advanced_subscriber_t
	receiver   <-chan zenoh.Sample
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Settings for retrieving historical data for the advanced subscriber.
// Pass as [option.Some] in [AdvancedSubscriberOptions.History] to enable history recovery.
type AdvancedSubscriberHistoryOptions struct {
	// Enable detection of late joiner publishers and query for their historical data.
	// Late joiner detection requires publishers to have enabled PublisherDetection.
	// History can only be retransmitted by publishers with caching enabled.
	DetectLatePublishers bool
	// Number of samples to query for each resource. 0 means no limit.
	MaxSamples uint
	// Maximum age of samples to query, in milliseconds. 0 means no limit.
	MaxAgeMs uint64
}

type lastSampleMissDetectionModeHeartbeat struct{}
type lastSampleMissDetectionModePeriodicQueries struct {
	period uint64
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Setting for detecting the last sample miss by the advanced subscriber.
type LastSampleMissDetectionMode struct {
	mode interface{}
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Heartbeat-based last sample detection.
func LastSampleMissDetectionModeHeartbeat() LastSampleMissDetectionMode {
	return LastSampleMissDetectionMode{
		mode: lastSampleMissDetectionModeHeartbeat{},
	}
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Periodic query-based last sample detection with the given query period in milliseconds.
// These queries allow to retrieve the last Sample(s) if the last Sample(s) is/are lost.
// So it is useful for sporadic publications but useless for periodic publications
// with a period smaller or equal to this period.
func LastSampleMissDetectionModePeriodicQueries(periodMs uint64) LastSampleMissDetectionMode {
	return LastSampleMissDetectionMode{
		mode: lastSampleMissDetectionModePeriodicQueries{period: max(1, periodMs)},
	}
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Disable last sample miss detection.
func LastSampleMissDetectionModeNone() LastSampleMissDetectionMode {
	return LastSampleMissDetectionMode{}
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Settings for recovering lost samples for the advanced subscriber.
type AdvancedSubscriberRecoveryOptions struct {
	// Settings for detecting the last sample miss.
	// Note: this does not affect intermediate sample miss detection (performed automatically when recovery is enabled).
	// Can be set to either [LastSampleMissDetectionModeHeartbeat], [LastSampleMissDetectionModePeriodicQueries] or [LastSampleMissDetectionModeNone].
	LastSampleMissDetection LastSampleMissDetectionMode
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Options passed to [SessionExt.DeclareAdvancedSubscriber] and [SessionExt.DeclareBackgroundAdvancedSubscriber].
type AdvancedSubscriberOptions struct {
	// Restrict the matching publications to those with compatible AllowedDestination.
	AllowedOrigin option.Option[zenoh.Locality]
	// Settings for querying historical data. Set to option.Some to enable.
	// History can only be retransmitted by publishers with caching enabled.
	History option.Option[AdvancedSubscriberHistoryOptions]
	// Settings for retransmission of lost samples. Set to option.Some to enable.
	// Retransmission requires publishers with caching and sample_miss_detection enabled.
	Recovery option.Option[AdvancedSubscriberRecoveryOptions]
	// Timeout for history and recovery queries, in milliseconds. 0 uses the default.
	QueryTimeoutMs uint64
	// Allow this subscriber to be detected through liveliness.
	SubscriberDetection bool
	// An optional key expression to be added to the liveliness token key expression.
	// Can be used to convey metadata.
	SubscriberDetectionMetadata option.Option[zenoh.KeyExpr]
}

func (opts *AdvancedSubscriberOptions) toCOpts(pinner *runtime.Pinner) C.ze_advanced_subscriber_options_t {
	var cOpts C.ze_advanced_subscriber_options_t
	C.ze_advanced_subscriber_options_default(&cOpts)

	// Base subscriber options
	if opts.AllowedOrigin.IsSome() {
		cOpts.subscriber_options.allowed_origin = uint32(opts.AllowedOrigin.Unwrap())
	}

	// History options
	if opts.History.IsSome() {
		history := opts.History.Unwrap()
		cOpts.history.is_enabled = C.bool(true)
		cOpts.history.detect_late_publishers = C.bool(history.DetectLatePublishers)
		cOpts.history.max_samples = C.size_t(history.MaxSamples)
		cOpts.history.max_age_ms = C.uint64_t(history.MaxAgeMs)
	}

	// Recovery options
	if opts.Recovery.IsSome() {
		recovery := opts.Recovery.Unwrap()
		cOpts.recovery.is_enabled = C.bool(true)
		switch mode := recovery.LastSampleMissDetection.mode.(type) {
		case lastSampleMissDetectionModeHeartbeat:
			cOpts.recovery.last_sample_miss_detection.is_enabled = C.bool(true)
			cOpts.recovery.last_sample_miss_detection.periodic_queries_period_ms = 0
		case lastSampleMissDetectionModePeriodicQueries:
			cOpts.recovery.last_sample_miss_detection.is_enabled = C.bool(true)
			cOpts.recovery.last_sample_miss_detection.periodic_queries_period_ms = C.uint64_t(mode.period)
		default:
			cOpts.recovery.last_sample_miss_detection.is_enabled = C.bool(false)
		}
	}

	// Query timeout
	cOpts.query_timeout_ms = C.uint64_t(opts.QueryTimeoutMs)

	// Subscriber detection
	cOpts.subscriber_detection = C.bool(opts.SubscriberDetection)
	if opts.SubscriberDetectionMetadata.IsSome() {
		ke := opts.SubscriberDetectionMetadata.Unwrap()
		cKeyexpr := (*C.z_view_keyexpr_t)(keyExprToUnsafeCPtr(&ke, pinner))
		cOpts.subscriber_detection_metadata = C.z_view_keyexpr_loan(cKeyexpr)
	}

	return cOpts
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Undeclare and destroy the advanced subscriber.
func (subscriber *AdvancedSubscriber) Undeclare() error {
	res := int8(C.ze_undeclare_advanced_subscriber(C.ze_advanced_subscriber_move(subscriber.subscriber)))
	if res == 0 {
		return nil
	}
	return zenoh.NewZError(res, "Failed to undeclare AdvancedSubscriber")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Destroy the advanced subscriber.
// This is equivalent to calling [AdvancedSubscriber.Undeclare] and discarding its return value.
func (subscriber *AdvancedSubscriber) Drop() {
	C.ze_advanced_subscriber_drop(C.ze_advanced_subscriber_move(subscriber.subscriber))
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Return the advanced subscriber's receiver channel if it was constructed with a channel handler, nil otherwise.
func (subscriber *AdvancedSubscriber) Handler() <-chan zenoh.Sample {
	return subscriber.receiver
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Get the key expression of the advanced subscriber.
func (subscriber *AdvancedSubscriber) KeyExpr() zenoh.KeyExpr {
	ke := C.zc_cgo_keyexpr_get_data(C.ze_advanced_subscriber_keyexpr(C.ze_advanced_subscriber_loan(subscriber.subscriber)))
	return newKeyExprFromUnsafeCDataPtr(unsafe.Pointer(&ke))
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Returns the advanced subscriber's entity global ID.
func (subscriber *AdvancedSubscriber) Id() zenoh.EntityGlobalId {
	cId := C.ze_advanced_subscriber_id(C.ze_advanced_subscriber_loan(subscriber.subscriber))
	return newEntityGlobalIdFromUnsafeCPtr(unsafe.Pointer(&cId))
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Declare a sample miss listener for this advanced subscriber.
// The listener will send notifications when a sample is missed.
// SampleMissListener MUST be explicitly destroyed using [SampleMissListener.Drop] once it is no longer needed.
func (subscriber *AdvancedSubscriber) DeclareSampleMissListener(handler zenoh.Handler[Miss]) (SampleMissListener, error) {
	callback, drop, _ := handler.ToCbDropHandler()
	closure := inner.NewClosure(callback, drop)
	var cClosure C.ze_owned_closure_miss_t
	C.ze_closure_miss(&cClosure, (*[0]byte)(C.zenohMissListenerCCallback), (*[0]byte)(C.zenohMissListenerDrop), unsafe.Pointer(closure))

	var cListener C.ze_owned_sample_miss_listener_t
	res := int8(C.ze_advanced_subscriber_declare_sample_miss_listener(C.ze_advanced_subscriber_loan(subscriber.subscriber), &cListener, C.ze_closure_miss_move(&cClosure)))

	if res == 0 {
		return SampleMissListener{listener: &cListener}, nil
	}
	return SampleMissListener{}, zenoh.NewZError(res, "Failed to declare SampleMissListener")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Declare a background sample miss listener for this advanced subscriber.
// The callback will run in the background until the subscriber is dropped.
func (subscriber *AdvancedSubscriber) DeclareBackgroundSampleMissListener(closure zenoh.Closure[Miss]) error {
	missClosureCtx := inner.NewClosure(closure.Call, closure.Drop)
	var cClosure C.ze_owned_closure_miss_t
	C.ze_closure_miss(&cClosure, (*[0]byte)(C.zenohMissListenerCCallback), (*[0]byte)(C.zenohMissListenerDrop), unsafe.Pointer(missClosureCtx))

	res := int8(C.ze_advanced_subscriber_declare_background_sample_miss_listener(C.ze_advanced_subscriber_loan(subscriber.subscriber), C.ze_closure_miss_move(&cClosure)))

	if res == 0 {
		return nil
	}
	return zenoh.NewZError(res, "Failed to declare background SampleMissListener")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Declare a liveliness subscriber to detect matching advanced publishers.
// Only advanced publishers that have enabled PublisherDetection can be detected.
// The returned subscriber MUST be explicitly destroyed using [AdvancedSubscriber.Undeclare] or [AdvancedSubscriber.Drop].
func (subscriber *AdvancedSubscriber) DetectPublishers(handler zenoh.Handler[zenoh.Sample], history bool) (zenoh.Subscriber, error) {
	callback, drop, recv := handler.ToCbDropHandler()
	closure := inner.NewClosure(callback, drop)
	var cClosure C.z_owned_closure_sample_t
	C.z_closure_sample(&cClosure, (*[0]byte)(C.zenohSubscriberCallback), (*[0]byte)(C.zenohSubscriberDrop), unsafe.Pointer(closure))

	var cOpts C.z_liveliness_subscriber_options_t
	C.z_liveliness_subscriber_options_default(&cOpts)
	cOpts.history = C.bool(history)

	var cLivelinessSubscriber C.z_owned_subscriber_t
	res := int8(C.ze_advanced_subscriber_detect_publishers(C.ze_advanced_subscriber_loan(subscriber.subscriber), &cLivelinessSubscriber, C.z_closure_sample_move(&cClosure), &cOpts))

	if res == 0 {
		return subscriberFromUnsafeCPtrAndReceiver(unsafe.Pointer(&cLivelinessSubscriber), recv), nil
	}
	return zenoh.Subscriber{}, zenoh.NewZError(res, "Failed to detect publishers for AdvancedSubscriber")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Declare a background liveliness subscriber to detect matching advanced publishers.
// Only advanced publishers with PublisherDetection enabled can be detected.
// The callback will run in the background until the session is closed.
func (subscriber *AdvancedSubscriber) DetectPublishersBackground(closure zenoh.Closure[zenoh.Sample], history bool) error {
	subClosure := inner.NewClosure(closure.Call, closure.Drop)
	var cClosure C.z_owned_closure_sample_t
	C.z_closure_sample(&cClosure, (*[0]byte)(C.zenohSubscriberCallback), (*[0]byte)(C.zenohSubscriberDrop), unsafe.Pointer(subClosure))

	var cOpts C.z_liveliness_subscriber_options_t
	C.z_liveliness_subscriber_options_default(&cOpts)
	cOpts.history = C.bool(history)

	res := int8(C.ze_advanced_subscriber_detect_publishers_background(C.ze_advanced_subscriber_loan(subscriber.subscriber), C.z_closure_sample_move(&cClosure), &cOpts))

	if res == 0 {
		return nil
	}
	return zenoh.NewZError(res, "Failed to detect publishers (background) for AdvancedSubscriber")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Construct and declare an advanced subscriber for the given key expression.
// Advanced subscriber MUST be explicitly destroyed using [AdvancedSubscriber.Undeclare] or [AdvancedSubscriber.Drop] once it is no longer needed.
func (session *SessionExt) DeclareAdvancedSubscriber(keyexpr zenoh.KeyExpr, handler zenoh.Handler[zenoh.Sample], options *AdvancedSubscriberOptions) (AdvancedSubscriber, error) {
	callback, drop, recv := handler.ToCbDropHandler()
	closure := inner.NewClosure(callback, drop)
	var cClosure C.z_owned_closure_sample_t
	C.z_closure_sample(&cClosure, (*[0]byte)(C.zenohSubscriberCallback), (*[0]byte)(C.zenohSubscriberDrop), unsafe.Pointer(closure))

	pinner := runtime.Pinner{}
	cKeyexpr := (*C.z_view_keyexpr_t)(keyExprToUnsafeCPtr(&keyexpr, &pinner))
	var cSubscriber C.ze_owned_advanced_subscriber_t
	res := int8(0)

	if options == nil {
		res = int8(C.ze_declare_advanced_subscriber(C.z_session_loan(session.getInner()), &cSubscriber, C.z_view_keyexpr_loan(cKeyexpr), C.z_closure_sample_move(&cClosure), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.ze_declare_advanced_subscriber(C.z_session_loan(session.getInner()), &cSubscriber, C.z_view_keyexpr_loan(cKeyexpr), C.z_closure_sample_move(&cClosure), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return AdvancedSubscriber{subscriber: &cSubscriber, receiver: recv}, nil
	}
	return AdvancedSubscriber{}, zenoh.NewZError(res, "Failed to declare AdvancedSubscriber")
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Construct and declare a background advanced subscriber.
// The subscriber callback will be called to process messages until the corresponding session is closed or dropped.
func (session *SessionExt) DeclareBackgroundAdvancedSubscriber(keyexpr zenoh.KeyExpr, closure zenoh.Closure[zenoh.Sample], options *AdvancedSubscriberOptions) error {
	subClosure := inner.NewClosure(closure.Call, closure.Drop)
	var cClosure C.z_owned_closure_sample_t
	C.z_closure_sample(&cClosure, (*[0]byte)(C.zenohSubscriberCallback), (*[0]byte)(C.zenohSubscriberDrop), unsafe.Pointer(subClosure))

	pinner := runtime.Pinner{}
	cKeyexpr := (*C.z_view_keyexpr_t)(keyExprToUnsafeCPtr(&keyexpr, &pinner))
	res := int8(0)

	if options == nil {
		res = int8(C.ze_declare_background_advanced_subscriber(C.z_session_loan(session.getInner()), C.z_view_keyexpr_loan(cKeyexpr), C.z_closure_sample_move(&cClosure), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.ze_declare_background_advanced_subscriber(C.z_session_loan(session.getInner()), C.z_view_keyexpr_loan(cKeyexpr), C.z_closure_sample_move(&cClosure), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return zenoh.NewZError(res, "Failed to declare background AdvancedSubscriber")
}
