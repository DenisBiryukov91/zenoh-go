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
	"sync"
	"testing"
	"time"
	"zenoh-go/zenoh"
)

func TestLivelinessGet(t *testing.T) {
	ke := "zenoh/liveliness/test/*"
	tokenKE := "zenoh/liveliness/test/1"

	session1, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	if err != nil {
		t.Fatalf("Failed to open session1: %v", err)
	}
	defer session1.Drop()

	session2, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	if err != nil {
		t.Fatalf("Failed to open session2: %v", err)
	}
	defer session2.Drop()

	tokenKeyExpr, _ := zenoh.NewKeyExpr(tokenKE)
	token, err := session1.Liveliness().DeclareToken(tokenKeyExpr, nil)
	if err != nil {
		t.Fatalf("Failed to declare liveliness token: %v", err)
	}
	defer token.Drop()

	time.Sleep(1 * time.Second)

	replies := make(chan zenoh.Reply, 3)
	keyExpr, _ := zenoh.NewKeyExpr(ke)
	session2.Liveliness().Get(
		keyExpr,
		func(reply zenoh.Reply) { replies <- reply },
		func() { close(replies) },
		nil,
	)

	found := false
	for reply := range replies {
		sample := reply.Ok().Unwrap()
		if reply.IsOk() && sample.KeyExpr().String() == tokenKE {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected liveliness token '%s' not found", tokenKE)
	}

	token.Drop()
	time.Sleep(1 * time.Second)

	replies = make(chan zenoh.Reply, 3)
	session2.Liveliness().Get(
		keyExpr,
		func(reply zenoh.Reply) { replies <- reply },
		func() { close(replies) },
		nil,
	)

	for reply := range replies {
		if reply.IsOk() {
			t.Errorf("Unexpected liveliness token found after undeclaration")
		}
	}
}

func TestLivelinessSubscriber(t *testing.T) {
	ke := "zenoh/liveliness/test/*"
	tokenKE1 := "zenoh/liveliness/test/1"
	tokenKE2 := "zenoh/liveliness/test/2"

	session1, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	if err != nil {
		t.Fatalf("Failed to open session1: %v", err)
	}
	defer session1.Drop()

	session2, err := zenoh.Open(zenoh.NewConfigDefault(), nil)
	if err != nil {
		t.Fatalf("Failed to open session2: %v", err)
	}
	defer session2.Drop()

	var wg sync.WaitGroup
	putTokens := make(map[string]bool)
	deleteTokens := make(map[string]bool)
	var mu sync.Mutex

	keyExpr, _ := zenoh.NewKeyExpr(ke)
	sub, err := session1.Liveliness().DeclareSubscriber(
		keyExpr,
		func(sample zenoh.Sample) {
			mu.Lock()
			defer mu.Unlock()
			if sample.Kind() == zenoh.SampleKindPut {
				putTokens[sample.KeyExpr().String()] = true
				wg.Done()
			} else if sample.Kind() == zenoh.SampleKindDelete {
				deleteTokens[sample.KeyExpr().String()] = true
				wg.Done()
			}
		}, nil, nil)
	if err != nil {
		t.Fatalf("Failed to declare liveliness subscriber: %v", err)
	}
	defer sub.Drop()

	wg.Add(2)
	tokenKeyExpr1, _ := zenoh.NewKeyExpr(tokenKE1)
	token1, _ := session2.Liveliness().DeclareToken(tokenKeyExpr1, nil)
	tokenKeyExpr2, _ := zenoh.NewKeyExpr(tokenKE2)
	token2, _ := session2.Liveliness().DeclareToken(tokenKeyExpr2, nil)
	defer token1.Drop()
	defer token2.Drop()

	wg.Wait()

	if !putTokens[tokenKE1] || !putTokens[tokenKE2] {
		t.Errorf("Expected liveliness tokens not received")
	}

	wg.Add(1)
	token1.Drop()
	wg.Wait()

	if !deleteTokens[tokenKE1] {
		t.Errorf("Liveliness token '%s' undeclaration not detected", tokenKE1)
	}
}
