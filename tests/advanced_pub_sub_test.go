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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/eclipse-zenoh/zenoh-go/zenoh"
	"github.com/eclipse-zenoh/zenoh-go/zenoh/zenohext"

	"github.com/BooleanCat/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var advancedPubSubValues = []string{
	"test_value_1",
	"test_value_2",
	"test_value_3",
	"test_value_4",
	"test_value_5",
	"test_value_6",
}

// TestAdvancedPubSubCache tests that an AdvancedSubscriber with history enabled
// can recover samples published before it was declared (cache recovery).
// The publisher publishes the first half of values before the subscriber is
// declared, and the second half after. The subscriber should receive all values.
func TestAdvancedPubSubCache(t *testing.T) {
	valuesCount := len(advancedPubSubValues)

	// Publisher session: enable timestamping (required for cache)
	pubConfig := zenoh.NewConfigDefault()
	err := pubConfig.InsertJson5("timestamping/enabled", "true")
	require.NoError(t, err, "Failed to enable timestamping in publisher config")

	sessionPub, err := zenoh.Open(pubConfig, nil)
	require.NoError(t, err)
	defer sessionPub.Drop()

	keyexpr, err := zenoh.NewKeyExpr("zenoh/test/advanced")
	require.NoError(t, err)

	// Declare AdvancedPublisher with cache large enough to hold all values,
	// publisher detection and sample miss detection enabled.
	pubOpts := zenohext.AdvancedPublisherOptions{
		Cache: option.Some(zenohext.AdvancedPublisherCacheOptions{
			MaxSamples: uint(valuesCount),
		}),
		PublisherDetection: true,
		SampleMissDetection: option.Some(zenohext.AdvancedPublisherSampleMissDetectionOptions{
			HeartbeatMode: zenohext.HeartbeatModeNone(),
		}),
	}
	pub, err := zenohext.Ext(&sessionPub).DeclareAdvancedPublisher(keyexpr, &pubOpts)
	require.NoError(t, err, "Failed to declare AdvancedPublisher")
	defer pub.Drop()

	// Publish the first half of values into the cache before the subscriber joins.
	for i := 0; i < valuesCount/2; i++ {
		err := pub.Put(zenoh.NewZBytesFromString(advancedPubSubValues[i]), nil)
		require.NoError(t, err)
	}

	// Give the router time to process.
	time.Sleep(500 * time.Millisecond)

	// Subscriber session (separate session to simulate a late joiner).
	sessionSub, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	require.NoError(t, err)
	defer sessionSub.Drop()

	var mu sync.Mutex
	receivedValues := []string{}
	var wg sync.WaitGroup
	wg.Add(valuesCount)

	// Declare AdvancedSubscriber with history (detect late publishers) and
	// recovery (periodic queries for last sample miss detection) enabled.
	subOpts := zenohext.AdvancedSubscriberOptions{
		History: option.Some(zenohext.AdvancedSubscriberHistoryOptions{
			DetectLatePublishers: true,
		}),
		Recovery: option.Some(zenohext.AdvancedSubscriberRecoveryOptions{
			LastSampleMissDetection: zenohext.LastSampleMissDetectionModePeriodicQueries(1000),
		}),
		SubscriberDetection: true,
	}

	sub, err := zenohext.Ext(&sessionSub).DeclareAdvancedSubscriber(
		keyexpr,
		zenoh.Closure[zenoh.Sample]{Call: func(sample zenoh.Sample) {
			mu.Lock()
			receivedValues = append(receivedValues, sample.Payload().String())
			mu.Unlock()
			wg.Done()
		}},
		&subOpts,
	)
	require.NoError(t, err, "Failed to declare AdvancedSubscriber")
	defer sub.Drop()

	// Give the subscriber time to query the publisher's cache.
	time.Sleep(1 * time.Second)

	// Publish the second half of values (live samples).
	for i := valuesCount / 2; i < valuesCount; i++ {
		err := pub.Put(zenoh.NewZBytesFromString(advancedPubSubValues[i]), nil)
		require.NoError(t, err)
	}

	// Wait for all values to be received with a timeout.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout: did not receive all expected samples")
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, valuesCount, len(receivedValues), "Expected to receive all %d values", valuesCount)
	for i, v := range advancedPubSubValues {
		assert.Equal(t, v, receivedValues[i], "Mismatch at index %d", i)
	}
}

func spawnRouter(locator string) (*zenoh.Session, error) {
	config := zenoh.NewConfigDefault()
	err := config.InsertJson5("listen/endpoints", fmt.Sprintf("[\"%s\"]", locator))
	if err != nil {
		return nil, fmt.Errorf("Failed to insert router endpoint into config: %w", err)
	}
	err = config.InsertJson5("scouting/multicast/enabled", "false")
	if err != nil {
		return nil, fmt.Errorf("Failed to insert multicast scouting config: %w", err)
	}
	err = config.InsertJson5("mode", "\"router\"")
	if err != nil {
		return nil, fmt.Errorf("Failed to insert router mode into config: %w", err)
	}
	session, err := zenoh.Open(config, nil)
	return &session, err
}

// TestAdvancedPubSubMissDetection tests that an AdvancedSubscriber with recovery
// enabled can detect and recover missed samples using periodic queries.
func TestAdvancedPubSubMissDetection(t *testing.T) {
	valuesCount := len(advancedPubSubValues)
	locator := "tcp/127.0.0.1:12001"
	router, err := spawnRouter(locator)
	require.NoError(t, err)
	defer router.Drop()

	pubConfig := zenoh.NewConfigDefault()
	err = pubConfig.InsertJson5("timestamping/enabled", "true")
	require.NoError(t, err)
	err = pubConfig.InsertJson5("connect/endpoints", fmt.Sprintf("[\"%s\"]", locator))
	require.NoError(t, err)
	err = pubConfig.InsertJson5("scouting/multicast/enabled", "false")
	require.NoError(t, err)
	err = pubConfig.InsertJson5("mode", "\"client\"")
	require.NoError(t, err)

	sessionPub, err := zenoh.Open(pubConfig, nil)
	require.NoError(t, err)
	defer sessionPub.Drop()

	subConfig := zenoh.NewConfigDefault()
	err = subConfig.InsertJson5("connect/endpoints", fmt.Sprintf("[\"%s\"]", locator))
	require.NoError(t, err)
	err = subConfig.InsertJson5("scouting/multicast/enabled", "false")
	require.NoError(t, err)
	err = subConfig.InsertJson5("mode", "\"client\"")
	require.NoError(t, err)
	sessionSub, err := zenoh.Open(subConfig, nil)
	require.NoError(t, err)
	defer sessionSub.Drop()

	keyexpr, err := zenoh.NewKeyExpr("zenoh/test/advanced/miss")
	require.NoError(t, err)

	pubOpts := zenohext.AdvancedPublisherOptions{
		SampleMissDetection: option.Some(zenohext.AdvancedPublisherSampleMissDetectionOptions{
			HeartbeatMode: zenohext.HeartbeatModeNone(),
		}),
	}
	pub, err := zenohext.Ext(&sessionPub).DeclareAdvancedPublisher(keyexpr, &pubOpts)
	require.NoError(t, err)
	defer pub.Drop()

	var mu sync.Mutex
	missedNbs := []uint32{}

	subOpts := zenohext.AdvancedSubscriberOptions{
		Recovery: option.Some(zenohext.AdvancedSubscriberRecoveryOptions{}),
	}

	var receivedValues []string
	var wg sync.WaitGroup
	wg.Add(valuesCount)

	sub, err := zenohext.Ext(&sessionSub).DeclareAdvancedSubscriber(
		keyexpr,
		zenoh.Closure[zenoh.Sample]{Call: func(sample zenoh.Sample) {
			mu.Lock()
			receivedValues = append(receivedValues, sample.Payload().String())
			mu.Unlock()
			wg.Done()
		}},
		&subOpts,
	)
	require.NoError(t, err)
	defer sub.Drop()

	err = sub.DeclareBackgroundSampleMissListener(zenoh.Closure[zenohext.Miss]{
		Call: func(miss zenohext.Miss) {
			mu.Lock()
			missedNbs = append(missedNbs, miss.Nb)
			mu.Unlock()
			for i := 0; i < int(miss.Nb); i++ {
				wg.Done()
			}
		},
	})
	require.NoError(t, err, "Failed to declare background sample miss listener")

	time.Sleep(500 * time.Millisecond)

	// Publish 2 values
	for i := 0; i < 2; i++ {
		err := pub.Put(zenoh.NewZBytesFromString(advancedPubSubValues[i]), nil)
		require.NoError(t, err)
	}
	time.Sleep(500 * time.Millisecond)

	// disable router to simulate a network failure that causes the subscriber to miss samples
	router.Drop()
	time.Sleep(500 * time.Millisecond)
	// Publish 2 more values
	for i := 2; i < 4; i++ {
		pub.Put(zenoh.NewZBytesFromString(advancedPubSubValues[i]), nil)
	}
	time.Sleep(500 * time.Millisecond)
	router, err = spawnRouter(locator)
	require.NoError(t, err)
	defer router.Drop()
	time.Sleep(5000 * time.Millisecond)
	// Publish the last 2 values
	for i := 4; i < 6; i++ {
		err := pub.Put(zenoh.NewZBytesFromString(advancedPubSubValues[i]), nil)
		require.NoError(t, err)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout: did not receive all expected samples")
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 4, len(receivedValues))
	// We should have missed 2 samples from a single source (the ones published while the router was down).
	assert.Equal(t, 1, len(missedNbs))
	assert.EqualValues(t, 2, missedNbs[0])
}
