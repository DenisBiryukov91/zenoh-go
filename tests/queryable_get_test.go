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

	"github.com/BooleanCat/option"
)

type QueryData struct {
	key     string
	params  string
	payload string
}

func (q QueryData) Equals(other QueryData) bool {
	return q.key == other.key && q.params == other.params && q.payload == other.payload
}

func TestQueryableGet(t *testing.T) {
	ke, _ := zenoh.NewKeyExpr("zenoh/test/*")
	selector, _ := zenoh.NewKeyExpr("zenoh/test/1")
	var queries []QueryData
	var replies []string
	var errors []string

	session1, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer session1.Drop()
	session2, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer session2.Drop()

	var wg sync.WaitGroup
	wg.Add(3)

	queryHandler := func(q zenoh.Query) {
		defer q.Drop()
		payload := q.Payload().Unwrap().String()
		qd := QueryData{
			key:     q.KeyExpr().String(),
			params:  q.Parameters(),
			payload: payload,
		}
		queries = append(queries, qd)

		if q.Parameters() == "ok" {
			q.Reply(q.KeyExpr(), zenoh.NewZBytesFromString(payload), nil)
		} else {
			q.ReplyErr(zenoh.NewZBytesFromString("err"), nil)
		}
	}

	queryable, _ := session1.DeclareQueryable(ke, zenoh.Closure[zenoh.Query]{Call: queryHandler}, nil)
	defer queryable.Drop()

	time.Sleep(1 * time.Second)

	sendQuery := func(payload, params string) {
		opts := zenoh.GetOptions{
			Payload:   option.Some(zenoh.NewZBytesFromString(payload)),
			TimeoutMs: 1000,
		}
		session2.Get(selector, params,
			zenoh.Closure[zenoh.Reply]{Call: func(reply zenoh.Reply) {
				if reply.IsOk() {
					sample := reply.Ok().Unwrap()
					replies = append(replies, sample.Payload().String())
				} else {
					err := reply.Err().Unwrap()
					errors = append(errors, err.Payload().String())
				}
				wg.Done()
			}}, &opts)
	}

	sendQuery("1", "ok")
	sendQuery("2", "ok")
	sendQuery("3", "err")

	wg.Wait()

	if len(queries) != 3 {
		t.Fatalf("Expected 3 queries, got %d", len(queries))
	}
	expectedQueries := []QueryData{
		{"zenoh/test/1", "ok", "1"},
		{"zenoh/test/1", "ok", "2"},
		{"zenoh/test/1", "err", "3"},
	}
	for i, qd := range expectedQueries {
		if !queries[i].Equals(qd) {
			t.Errorf("Query %d does not match expected. Got %+v, expected %+v", i, queries[i], qd)
		}
	}

	if len(replies) != 2 || replies[0] != "1" || replies[1] != "2" {
		t.Errorf("Unexpected replies: %+v", replies)
	}

	if len(errors) != 1 || errors[0] != "err" {
		t.Errorf("Unexpected errors: %+v", errors)
	}
}
