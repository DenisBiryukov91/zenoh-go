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
	"testing"
	"time"
	"zenoh-go/zenoh"

	"github.com/BooleanCat/option"
	"github.com/stretchr/testify/assert"
)

func TestPublisherMatchingStatus(t *testing.T) {
	sessionPub, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionPub.Drop()

	keyexpr, _ := zenoh.NewKeyExpr("zenoh/test/matching")
	pub, _ := sessionPub.DeclarePublisher(keyexpr, nil)
	defer pub.Drop()
	res, err := pub.GetMatchingStatus()
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Matching, false)

	sessionSub, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionSub.Drop()
	sub, _ := sessionSub.DeclareSubscriber(keyexpr, zenoh.NewFifoChannel[zenoh.Sample](16), nil)
	time.Sleep(1 * time.Second)

	res, err = pub.GetMatchingStatus()
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Matching, true)

	sub.Drop()

	time.Sleep(1 * time.Second)

	res, err = pub.GetMatchingStatus()
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Matching, false)
}

func TestPublisherMatchingListener(t *testing.T) {
	sessionPub, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionPub.Drop()

	keyexpr, _ := zenoh.NewKeyExpr("zenoh/test/matching")
	pub, _ := sessionPub.DeclarePublisher(keyexpr, nil)
	defer pub.Drop()
	listener, _ := pub.DeclareMatchingListener(zenoh.NewFifoChannel[zenoh.MatchingStatus](16))
	defer listener.Drop()
	assert.Equal(t, len(listener.Handler()), 0)

	sessionSub, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionSub.Drop()
	sub, _ := sessionSub.DeclareSubscriber(keyexpr, zenoh.NewFifoChannel[zenoh.Sample](16), nil)
	time.Sleep(1 * time.Second)

	assert.Equal(t, len(listener.Handler()), 1)
	status := <-listener.Handler()
	assert.Equal(t, status.Matching, true)

	sub.Drop()
	time.Sleep(1 * time.Second)

	assert.Equal(t, len(listener.Handler()), 1)
	status = <-listener.Handler()
	assert.Equal(t, status.Matching, false)
}

func TestPublisherBackgroundMatchingListener(t *testing.T) {
	sessionPub, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionPub.Drop()

	keyexpr, _ := zenoh.NewKeyExpr("zenoh/test/matching")
	pub, _ := sessionPub.DeclarePublisher(keyexpr, nil)
	defer pub.Drop()

	var statuses []bool

	_ = pub.DeclareBackgroundMatchingListener(
		zenoh.Closure[zenoh.MatchingStatus]{Call: func(status zenoh.MatchingStatus) { statuses = append(statuses, status.Matching) }})

	sessionSub, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionSub.Drop()
	sub, _ := sessionSub.DeclareSubscriber(keyexpr, zenoh.NewFifoChannel[zenoh.Sample](16), nil)
	time.Sleep(1 * time.Second)

	assert.Equal(t, len(statuses), 1)
	assert.Equal(t, statuses[0], true)

	sub.Drop()
	time.Sleep(1 * time.Second)

	assert.Equal(t, len(statuses), 2)
	assert.Equal(t, statuses[1], false)
}

func TestQuerierMatchingStatus(t *testing.T) {
	sessionQuerier, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionQuerier.Drop()

	keyexpr1, _ := zenoh.NewKeyExpr("zenoh/test/matching")
	keyexpr2, _ := zenoh.NewKeyExpr("zenoh/test/**")
	querier1, _ := sessionQuerier.DeclareQuerier(keyexpr1, nil)
	defer querier1.Drop()
	querier2, _ := sessionQuerier.DeclareQuerier(keyexpr2, &zenoh.QuerierOptions{Target: option.Some(zenoh.QueryTargetAllComplete)})
	defer querier2.Drop()

	res, err := querier1.GetMatchingStatus()
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Matching, false)

	res, err = querier2.GetMatchingStatus()
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Matching, false)

	sessionQbl, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionQbl.Drop()
	queryable1, _ := sessionQbl.DeclareQueryable(keyexpr1, zenoh.NewFifoChannel[zenoh.Query](16), nil)
	time.Sleep(1 * time.Second)

	res, err = querier1.GetMatchingStatus()
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Matching, true)

	res, err = querier2.GetMatchingStatus()
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Matching, false)

	queryable2, _ := sessionQbl.DeclareQueryable(keyexpr2, zenoh.NewFifoChannel[zenoh.Query](16), &zenoh.QueryableOptions{Complete: true})
	time.Sleep(1 * time.Second)

	res, err = querier1.GetMatchingStatus()
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Matching, true)

	res, err = querier2.GetMatchingStatus()
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Matching, true)

	queryable2.Drop()
	time.Sleep(1 * time.Second)

	res, err = querier1.GetMatchingStatus()
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Matching, true)

	res, err = querier2.GetMatchingStatus()
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Matching, false)

	queryable1.Drop()
	time.Sleep(1 * time.Second)

	res, err = querier1.GetMatchingStatus()
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Matching, false)

	res, err = querier2.GetMatchingStatus()
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Matching, false)
}

func TestQuerierMatchingListener(t *testing.T) {
	sessionQuerier, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionQuerier.Drop()

	keyexpr1, _ := zenoh.NewKeyExpr("zenoh/test/matching")
	keyexpr2, _ := zenoh.NewKeyExpr("zenoh/test/**")
	querier1, _ := sessionQuerier.DeclareQuerier(keyexpr1, nil)
	defer querier1.Drop()
	querier2, _ := sessionQuerier.DeclareQuerier(keyexpr2, &zenoh.QuerierOptions{Target: option.Some(zenoh.QueryTargetAllComplete)})
	defer querier2.Drop()

	listener1, _ := querier1.DeclareMatchingListener(zenoh.NewFifoChannel[zenoh.MatchingStatus](16))
	defer listener1.Drop()
	listener2, _ := querier2.DeclareMatchingListener(zenoh.NewFifoChannel[zenoh.MatchingStatus](16))
	defer listener2.Drop()

	assert.Equal(t, len(listener1.Handler()), 0)
	assert.Equal(t, len(listener2.Handler()), 0)

	sessionQbl, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionQbl.Drop()
	queryable1, _ := sessionQbl.DeclareQueryable(keyexpr1, zenoh.NewFifoChannel[zenoh.Query](16), nil)
	time.Sleep(1 * time.Second)

	assert.Equal(t, len(listener1.Handler()), 1)
	status := <-listener1.Handler()
	assert.Equal(t, status.Matching, true)
	assert.Equal(t, len(listener2.Handler()), 0)

	queryable2, _ := sessionQbl.DeclareQueryable(keyexpr2, zenoh.NewFifoChannel[zenoh.Query](16), &zenoh.QueryableOptions{Complete: true})
	time.Sleep(1 * time.Second)

	assert.Equal(t, len(listener1.Handler()), 0)
	assert.Equal(t, len(listener2.Handler()), 1)
	status = <-listener2.Handler()
	assert.Equal(t, status.Matching, true)

	queryable2.Drop()
	time.Sleep(1 * time.Second)

	assert.Equal(t, len(listener1.Handler()), 0)
	assert.Equal(t, len(listener2.Handler()), 1)
	status = <-listener2.Handler()
	assert.Equal(t, status.Matching, false)

	queryable1.Drop()
	time.Sleep(1 * time.Second)

	assert.Equal(t, len(listener1.Handler()), 1)
	status = <-listener1.Handler()
	assert.Equal(t, status.Matching, false)
	assert.Equal(t, len(listener2.Handler()), 0)
}

func TestQuerierBackgroundMatchingListener(t *testing.T) {
	sessionQuerier, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionQuerier.Drop()

	keyexpr1, _ := zenoh.NewKeyExpr("zenoh/test/matching")
	keyexpr2, _ := zenoh.NewKeyExpr("zenoh/test/**")
	querier1, _ := sessionQuerier.DeclareQuerier(keyexpr1, nil)
	defer querier1.Drop()
	querier2, _ := sessionQuerier.DeclareQuerier(keyexpr2, &zenoh.QuerierOptions{Target: option.Some(zenoh.QueryTargetAllComplete)})
	defer querier2.Drop()

	var statuses1 []bool
	var statuses2 []bool

	_ = querier1.DeclareBackgroundMatchingListener(
		zenoh.Closure[zenoh.MatchingStatus]{Call: func(status zenoh.MatchingStatus) { statuses1 = append(statuses1, status.Matching) }})

	_ = querier2.DeclareBackgroundMatchingListener(
		zenoh.Closure[zenoh.MatchingStatus]{Call: func(status zenoh.MatchingStatus) { statuses2 = append(statuses2, status.Matching) }})

	sessionQbl, _ := zenoh.Open(zenoh.NewConfigDefault(), nil)
	defer sessionQbl.Drop()
	queryable1, _ := sessionQbl.DeclareQueryable(keyexpr1, zenoh.NewFifoChannel[zenoh.Query](16), nil)
	time.Sleep(1 * time.Second)

	assert.Equal(t, len(statuses1), 1)
	assert.Equal(t, statuses1[0], true)
	assert.Equal(t, len(statuses2), 0)

	queryable2, _ := sessionQbl.DeclareQueryable(keyexpr2, zenoh.NewFifoChannel[zenoh.Query](16), &zenoh.QueryableOptions{Complete: true})
	time.Sleep(1 * time.Second)

	assert.Equal(t, len(statuses1), 1)
	assert.Equal(t, len(statuses2), 1)
	assert.Equal(t, statuses2[0], true)

	queryable2.Drop()
	time.Sleep(1 * time.Second)

	assert.Equal(t, len(statuses1), 1)
	assert.Equal(t, len(statuses2), 2)
	assert.Equal(t, statuses2[1], false)

	queryable1.Drop()
	time.Sleep(1 * time.Second)

	assert.Equal(t, len(statuses1), 2)
	assert.Equal(t, len(statuses2), 2)
	assert.Equal(t, statuses1[1], false)
}
