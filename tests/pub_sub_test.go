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

package zenoh_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"zenoh-go/zenoh"

	"github.com/stretchr/testify/assert"
)

func TestPubSub(t *testing.T) {
	sessionPub, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionPub.Drop()
	sessionSub, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionSub.Drop()

	keyexpr, _ := zenoh.NewKeyExpr("zenoh/test")
	pub, _ := sessionPub.DeclarePublisher(keyexpr, nil)
	defer pub.Drop()

	var wg sync.WaitGroup
	receivedMessages := []string{}
	wg.Add(2)

	_, _ = sessionSub.DeclareSubscriber(keyexpr,
		zenoh.Closure[zenoh.Sample]{Call: func(sample zenoh.Sample) {
			receivedMessages = append(receivedMessages, sample.Payload().String())
			wg.Done()
		}}, nil)

	time.Sleep(1 * time.Second)

	pub.Put(zenoh.NewZBytesFromString("first"), nil)
	pub.Put(zenoh.NewZBytesFromString("second"), nil)

	wg.Wait()

	assert.Equal(t, 2, len(receivedMessages))
	assert.Equal(t, "first", receivedMessages[0])
	assert.Equal(t, "second", receivedMessages[1])
}

func TestPutSub(t *testing.T) {
	sessionPub, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionPub.Drop()
	sessionSub, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionSub.Drop()

	keyexpr, _ := zenoh.NewKeyExpr("zenoh/test")

	var wg sync.WaitGroup
	receivedMessages := []string{}
	wg.Add(2)

	_, _ = sessionSub.DeclareSubscriber(keyexpr, zenoh.Closure[zenoh.Sample]{Call: func(sample zenoh.Sample) {
		receivedMessages = append(receivedMessages, sample.Payload().String())
		wg.Done()
	}}, nil)

	time.Sleep(1 * time.Second)

	sessionPub.Put(keyexpr, zenoh.NewZBytesFromString("first"), nil)
	sessionPub.Put(keyexpr, zenoh.NewZBytesFromString("second"), nil)

	wg.Wait()

	assert.Equal(t, 2, len(receivedMessages))
	assert.Equal(t, "first", receivedMessages[0])
	assert.Equal(t, "second", receivedMessages[1])
}

func TestPutSubFifoChannel(t *testing.T) {
	sessionPub, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionPub.Drop()
	sessionSub, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionSub.Drop()

	keyexpr, _ := zenoh.NewKeyExpr("zenoh/test")

	sub, _ := sessionSub.DeclareSubscriber(keyexpr, zenoh.NewFifoChannel[zenoh.Sample](2), nil)
	time.Sleep(1 * time.Second)

	sessionPub.Put(keyexpr, zenoh.NewZBytesFromString("first"), nil)
	sessionPub.Put(keyexpr, zenoh.NewZBytesFromString("second"), nil)

	select {
	case sample := <-sub.Handler():
		assert.Equal(t, "first", sample.Payload().String())
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for first message")
	}

	select {
	case sample := <-sub.Handler():
		assert.Equal(t, "second", sample.Payload().String())
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for second message")
	}

	select {
	case <-sub.Handler():
		t.Fatal("Unexpected message received")
	case <-time.After(1 * time.Second):
		fmt.Println("No more data, as expected.")
	}
}

func TestPutSubRingChannel(t *testing.T) {
	sessionPub, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionPub.Drop()
	sessionSub, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionSub.Drop()

	keyexpr, _ := zenoh.NewKeyExpr("zenoh/test")

	sub, _ := sessionSub.DeclareSubscriber(keyexpr, zenoh.NewRingChannel[zenoh.Sample](2), nil)
	time.Sleep(1 * time.Second)

	sessionPub.Put(keyexpr, zenoh.NewZBytesFromString("first"), nil)
	sessionPub.Put(keyexpr, zenoh.NewZBytesFromString("second"), nil)
	sessionPub.Put(keyexpr, zenoh.NewZBytesFromString("third"), nil)
	time.Sleep(1 * time.Second)

	select {
	case sample := <-sub.Handler():
		assert.Equal(t, "second", sample.Payload().String())
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for first message")
	}

	select {
	case sample := <-sub.Handler():
		assert.Equal(t, "third", sample.Payload().String())
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for second message")
	}

	select {
	case <-sub.Handler():
		t.Fatal("Unexpected message received")
	case <-time.After(1 * time.Second):
		fmt.Println("No more data, as expected.")
	}
}
