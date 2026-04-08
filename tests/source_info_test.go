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

package zenoh_test

import (
	"sync"
	"testing"
	"time"

	"github.com/eclipse-zenoh/zenoh-go/zenoh"

	"github.com/BooleanCat/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertSourceInfoEqual checks that the received SourceInfo matches the expected
// entity global ID (by zid and eid) and sequence number.
func assertSourceInfoEqual(t *testing.T, expected zenoh.SourceInfo, received option.Option[zenoh.SourceInfo]) {
	t.Helper()
	require.True(t, received.IsSome(), "expected source info to be present in received message")
	got := received.Unwrap()
	assert.Equal(t, expected.Id().ZId().String(), got.Id().ZId().String(), "source info ZId mismatch")
	assert.Equal(t, expected.Id().EntityId(), got.Id().EntityId(), "source info EntityId mismatch")
	assert.Equal(t, expected.Sn(), got.Sn(), "source info sequence number mismatch")
}

// TestPutSourceInfo verifies that a SourceInfo set in PutOptions is propagated
// to the received Sample on the subscriber side.
func TestPutSourceInfo(t *testing.T) {
	sessionPub, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	require.NoError(t, err)
	defer sessionPub.Drop()

	sessionSub, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	require.NoError(t, err)
	defer sessionSub.Drop()

	keyexpr, err := zenoh.NewKeyExpr("zenoh/test/source_info/put")
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	var receivedSample zenoh.Sample

	_, err = sessionSub.DeclareSubscriber(keyexpr, zenoh.Closure[zenoh.Sample]{
		Call: func(sample zenoh.Sample) {
			receivedSample = sample
			wg.Done()
		},
	}, nil)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	// Build a SourceInfo using the publisher session's entity global ID.
	pubId := sessionPub.Id()
	const sn = uint32(42)
	sourceInfo := zenoh.NewSourceInfo(pubId, sn)

	err = sessionPub.Put(keyexpr, zenoh.NewZBytesFromString("hello"), &zenoh.PutOptions{
		SourceInfo: option.Some(sourceInfo),
	})
	require.NoError(t, err)

	wg.Wait()

	assertSourceInfoEqual(t, sourceInfo, receivedSample.SourceInfo())
}

// TestPublisherPutSourceInfo verifies that a SourceInfo set in PublisherPutOptions
// is propagated to the received Sample on the subscriber side.
func TestPublisherPutSourceInfo(t *testing.T) {
	sessionPub, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	require.NoError(t, err)
	defer sessionPub.Drop()

	sessionSub, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	require.NoError(t, err)
	defer sessionSub.Drop()

	keyexpr, err := zenoh.NewKeyExpr("zenoh/test/source_info/publisher_put")
	require.NoError(t, err)

	pub, err := sessionPub.DeclarePublisher(keyexpr, nil)
	require.NoError(t, err)
	defer pub.Drop()

	var wg sync.WaitGroup
	wg.Add(1)
	var receivedSample zenoh.Sample

	_, err = sessionSub.DeclareSubscriber(keyexpr, zenoh.Closure[zenoh.Sample]{
		Call: func(sample zenoh.Sample) {
			receivedSample = sample
			wg.Done()
		},
	}, nil)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	pubId := pub.Id()
	const sn = uint32(7)
	sourceInfo := zenoh.NewSourceInfo(pubId, sn)

	err = pub.Put(zenoh.NewZBytesFromString("hello"), &zenoh.PublisherPutOptions{
		SourceInfo: option.Some(sourceInfo),
	})
	require.NoError(t, err)

	wg.Wait()

	assertSourceInfoEqual(t, sourceInfo, receivedSample.SourceInfo())
}

// TestGetSourceInfo verifies that a SourceInfo set in GetOptions is propagated
// to the Query received by the queryable.
func TestGetSourceInfo(t *testing.T) {
	sessionQueryable, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	require.NoError(t, err)
	defer sessionQueryable.Drop()

	sessionGet, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	require.NoError(t, err)
	defer sessionGet.Drop()

	keyexpr, err := zenoh.NewKeyExpr("zenoh/test/source_info/get")
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	var receivedQuery zenoh.Query

	queryable, err := sessionQueryable.DeclareQueryable(keyexpr, zenoh.Closure[zenoh.Query]{
		Call: func(q zenoh.Query) {
			receivedQuery = q
			q.Reply(q.KeyExpr(), zenoh.NewZBytesFromString("ok"), nil)
			wg.Done()
		},
	}, nil)
	require.NoError(t, err)
	defer queryable.Drop()

	time.Sleep(500 * time.Millisecond)

	getId := sessionGet.Id()
	const sn = uint32(99)
	sourceInfo := zenoh.NewSourceInfo(getId, sn)

	replyCh, err := sessionGet.Get(keyexpr, "", zenoh.NewFifoChannel[zenoh.Reply](1), &zenoh.GetOptions{
		TimeoutMs:  1000,
		SourceInfo: option.Some(sourceInfo),
	})
	require.NoError(t, err)
	// Drain the reply channel.
	<-replyCh

	wg.Wait()

	assertSourceInfoEqual(t, sourceInfo, receivedQuery.SourceInfo())
	receivedQuery.Drop()
}

// TestQuerierGetSourceInfo verifies that a SourceInfo set in QuerierGetOptions
// is propagated to the Query received by the queryable.
func TestQuerierGetSourceInfo(t *testing.T) {
	sessionQueryable, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	require.NoError(t, err)
	defer sessionQueryable.Drop()

	sessionQuerier, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	require.NoError(t, err)
	defer sessionQuerier.Drop()

	keyexpr, err := zenoh.NewKeyExpr("zenoh/test/source_info/querier_get")
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	var receivedQuery zenoh.Query

	queryable, err := sessionQueryable.DeclareQueryable(keyexpr, zenoh.Closure[zenoh.Query]{
		Call: func(q zenoh.Query) {
			receivedQuery = q
			q.Reply(q.KeyExpr(), zenoh.NewZBytesFromString("ok"), nil)
			wg.Done()
		},
	}, nil)
	require.NoError(t, err)
	defer queryable.Drop()

	time.Sleep(500 * time.Millisecond)

	querier, err := sessionQuerier.DeclareQuerier(keyexpr, &zenoh.QuerierOptions{TimeoutMs: 1000})
	require.NoError(t, err)
	defer querier.Drop()

	querierId := querier.Id()
	const sn = uint32(55)
	sourceInfo := zenoh.NewSourceInfo(querierId, sn)

	replyCh, err := querier.Get("", zenoh.NewFifoChannel[zenoh.Reply](1), &zenoh.QuerierGetOptions{
		SourceInfo: option.Some(sourceInfo),
	})
	require.NoError(t, err)
	// Drain the reply channel.
	<-replyCh

	wg.Wait()

	assertSourceInfoEqual(t, sourceInfo, receivedQuery.SourceInfo())
	receivedQuery.Drop()
}

// TestReplySourceInfo verifies that a SourceInfo set in QueryReplyOptions
// is propagated to the Reply received by the getter.
func TestReplySourceInfo(t *testing.T) {
	sessionQueryable, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	require.NoError(t, err)
	defer sessionQueryable.Drop()

	sessionGet, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	require.NoError(t, err)
	defer sessionGet.Drop()

	keyexpr, err := zenoh.NewKeyExpr("zenoh/test/source_info/reply")
	require.NoError(t, err)

	queryableId := sessionQueryable.Id()
	const sn = uint32(13)
	sourceInfo := zenoh.NewSourceInfo(queryableId, sn)

	var queryableWg sync.WaitGroup
	queryableWg.Add(1)

	queryable, err := sessionQueryable.DeclareQueryable(keyexpr, zenoh.Closure[zenoh.Query]{
		Call: func(q zenoh.Query) {
			defer q.Drop()
			q.Reply(q.KeyExpr(), zenoh.NewZBytesFromString("reply_payload"), &zenoh.QueryReplyOptions{
				SourceInfo: option.Some(sourceInfo),
			})
			queryableWg.Done()
		},
	}, nil)
	require.NoError(t, err)
	defer queryable.Drop()

	time.Sleep(500 * time.Millisecond)

	var receivedReply zenoh.Reply
	var replyWg sync.WaitGroup
	replyWg.Add(1)

	_, err = sessionGet.Get(keyexpr, "", zenoh.Closure[zenoh.Reply]{
		Call: func(r zenoh.Reply) {
			receivedReply = r
			replyWg.Done()
		},
	}, &zenoh.GetOptions{TimeoutMs: 1000})
	require.NoError(t, err)

	queryableWg.Wait()
	replyWg.Wait()

	require.True(t, receivedReply.IsOk(), "expected an Ok reply")
	sample := receivedReply.Ok().Unwrap()
	assertSourceInfoEqual(t, sourceInfo, sample.SourceInfo())
}
