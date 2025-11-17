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
	"testing"
	"time"

	"zenoh-go/zenoh"

	"github.com/BooleanCat/option"
	"github.com/stretchr/testify/assert"
)

func TestCancellationGet(t *testing.T) {
	ke, _ := zenoh.NewKeyExpr("zenoh-go/query/cancellation_test")
	session1, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer session1.Drop()
	session2, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer session2.Drop()

	queryable, _ := session1.DeclareQueryable(ke, zenoh.NewFifoChannel[zenoh.Query](16), nil)
	defer queryable.Drop()

	time.Sleep(1 * time.Second)
	{
		fmt.Println("Check cancel removes callbacks")
		ct := zenoh.NewCancellationToken()
		opts := zenoh.GetOptions{CancellationToken: option.Some(ct)}
		dropped := false

		session2.Get(ke, "",
			zenoh.Closure[zenoh.Reply]{
				Call: func(reply zenoh.Reply) {},
				Drop: func() { dropped = true },
			},
			&opts)
		time.Sleep(1 * time.Second)

		assert.False(t, ct.IsCancelled())
		ct.Cancel()
		assert.True(t, ct.IsCancelled())
		assert.True(t, dropped)
		q := <-queryable.Handler()
		q.Drop()
	}
	{
		fmt.Println("Check cancel blocks until callback is finished")
		ct := zenoh.NewCancellationToken()
		opts := zenoh.GetOptions{CancellationToken: option.Some(ct)}
		val := 0

		session2.Get(ke, "",
			zenoh.Closure[zenoh.Reply]{
				Call: func(reply zenoh.Reply) {
					time.Sleep(1 * time.Second)
					val += 1
					time.Sleep(1 * time.Second)
				},
				Drop: func() {
					time.Sleep(1 * time.Second)
					val += 1
					time.Sleep(1 * time.Second)
				},
			},
			&opts)
		q := <-queryable.Handler()
		q.Reply(ke, zenoh.NewZBytesFromString("ok"), nil)
		q.Drop()
		time.Sleep(1 * time.Second)

		ct.Cancel()
		assert.True(t, ct.IsCancelled())
		assert.Equal(t, val, 2)
	}
	{
		fmt.Println("Check cancelled token does not send a query")
		ct := zenoh.NewCancellationToken()
		ct.Cancel()
		opts := zenoh.GetOptions{CancellationToken: option.Some(ct)}
		dropped := false
		session2.Get(ke, "",
			zenoh.Closure[zenoh.Reply]{
				Call: func(reply zenoh.Reply) {},
				Drop: func() {
					dropped = true
				},
			},
			&opts)

		ct.Cancel()
		assert.True(t, dropped)
	}
}

func TestCancellationQuerierGet(t *testing.T) {
	ke, _ := zenoh.NewKeyExpr("zenoh-go/querier/cancellation_test")
	session1, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer session1.Drop()
	session2, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer session2.Drop()

	queryable, _ := session1.DeclareQueryable(ke, zenoh.NewFifoChannel[zenoh.Query](16), nil)
	defer queryable.Drop()
	querier, _ := session2.DeclareQuerier(ke, nil)
	defer querier.Drop()

	time.Sleep(1 * time.Second)
	{
		fmt.Println("Check cancel removes callbacks")
		ct := zenoh.NewCancellationToken()
		opts := zenoh.QuerierGetOptions{CancellationToken: option.Some(ct)}
		dropped := false

		querier.Get("",
			zenoh.Closure[zenoh.Reply]{
				Call: func(reply zenoh.Reply) {},
				Drop: func() { dropped = true },
			},
			&opts)
		time.Sleep(1 * time.Second)

		assert.False(t, ct.IsCancelled())
		ct.Cancel()
		assert.True(t, ct.IsCancelled())
		assert.True(t, dropped)
		q := <-queryable.Handler()
		q.Drop()
	}
	{
		fmt.Println("Check cancel blocks until callback is finished")
		ct := zenoh.NewCancellationToken()
		opts := zenoh.QuerierGetOptions{CancellationToken: option.Some(ct)}
		val := 0

		querier.Get("",
			zenoh.Closure[zenoh.Reply]{
				Call: func(reply zenoh.Reply) {
					time.Sleep(1 * time.Second)
					val += 1
					time.Sleep(1 * time.Second)
				},
				Drop: func() {
					time.Sleep(1 * time.Second)
					val += 1
					time.Sleep(1 * time.Second)
				},
			},
			&opts)
		q := <-queryable.Handler()
		q.Reply(ke, zenoh.NewZBytesFromString("ok"), nil)
		q.Drop()
		time.Sleep(1 * time.Second)

		ct.Cancel()
		assert.True(t, ct.IsCancelled())
		assert.Equal(t, val, 2)
	}
	{
		fmt.Println("Check cancelled token does not send a query")
		ct := zenoh.NewCancellationToken()
		ct.Cancel()
		opts := zenoh.QuerierGetOptions{CancellationToken: option.Some(ct)}
		dropped := false
		querier.Get("",
			zenoh.Closure[zenoh.Reply]{
				Call: func(reply zenoh.Reply) {},
				Drop: func() {
					dropped = true
				},
			},
			&opts)

		ct.Cancel()
		assert.True(t, dropped)
	}
}

func TestLivelinessCancellationGet(t *testing.T) {
	ke, _ := zenoh.NewKeyExpr("zenoh-go/liveliness/cancellation_test")
	session1, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer session1.Drop()
	session2, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer session2.Drop()

	token, _ := session1.Liveliness().DeclareToken(ke, nil)
	defer token.Drop()

	time.Sleep(1 * time.Second)
	{
		fmt.Println("Check cancel blocks until callback is finished")
		ct := zenoh.NewCancellationToken()
		opts := zenoh.LivelinessGetOptions{CancellationToken: option.Some(ct)}
		val := 0

		session2.Liveliness().Get(ke,
			zenoh.Closure[zenoh.Reply]{
				Call: func(reply zenoh.Reply) {
					time.Sleep(1 * time.Second)
					val += 1
					time.Sleep(1 * time.Second)
				},
				Drop: func() {
					time.Sleep(1 * time.Second)
					val += 1
					time.Sleep(1 * time.Second)
				},
			},
			&opts)
		time.Sleep(1 * time.Second)

		ct.Cancel()
		assert.True(t, ct.IsCancelled())
		assert.Equal(t, val, 2)
	}
	{
		fmt.Println("Check cancelled token does not send a query")
		ct := zenoh.NewCancellationToken()
		ct.Cancel()
		opts := zenoh.LivelinessGetOptions{CancellationToken: option.Some(ct)}
		dropped := false
		session2.Liveliness().Get(ke,
			zenoh.Closure[zenoh.Reply]{
				Call: func(reply zenoh.Reply) {},
				Drop: func() {
					dropped = true
				},
			},
			&opts)

		ct.Cancel()
		assert.True(t, dropped)
	}
}
