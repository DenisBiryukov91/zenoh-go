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

package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
	"zenoh-go/examples/utils"
	"zenoh-go/zenoh"

	"github.com/spf13/pflag"
)

type Stats struct {
	count            uint64
	finishedRounds   uint64
	start            time.Time
	maxRounds        uint64
	messagesPerRound uint64
}

func newStats(maxRounds uint64, messagesPerRound uint64) Stats {
	return Stats{maxRounds: maxRounds, messagesPerRound: messagesPerRound}
}

func (stats *Stats) update(_ *zenoh.Sample) {
	if stats.count == 0 {
		stats.start = time.Now()
		stats.count++
	} else if stats.count < stats.messagesPerRound {
		stats.count++
	} else {
		stats.finishedRounds++
		msgsPerSecond := float64(stats.messagesPerRound) * 1000000.0 / float64(time.Since(stats.start).Microseconds())
		fmt.Printf("%v msgs/s\n", msgsPerSecond)
		stats.count = 0
		if stats.finishedRounds > stats.maxRounds {
			os.Exit(0)
		}
	}
}

func main() {
	zenoh.InitLoggerFromEnvOr("error")
	args := parseArgs()

	fmt.Println("Opening session...")
	session, err := zenoh.Open(args.config, nil)
	if err != nil {
		fmt.Printf("Failed to open Zenoh session: %v\n", err)
		os.Exit(-1)
	}
	defer session.Drop()

	stats := newStats(args.Samples, args.numMessages)
	keyexpr, _ := zenoh.NewKeyExpr("test/thr")
	fmt.Printf("Declaring Subscriber on '%s'...\n", keyexpr)
	sub, err := session.DeclareSubscriber(keyexpr, zenoh.Closure[zenoh.Sample]{Call: func(sample zenoh.Sample) { stats.update(&sample) }}, nil)
	if err != nil {
		fmt.Println("Unable to declare subscriber.")
		os.Exit(-1)
	}
	defer sub.Drop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")
	<-stop
}

const (
	defaultsamples = 10
	defaultNumber  = 1000000
)

type Args struct {
	Samples     uint64
	numMessages uint64
	config      zenoh.Config
}

func parseArgs() Args {
	var samples uint64
	var numMessages uint64
	pflag.Uint64VarP(&samples, "samples", "s", defaultsamples, "Number of throughput measurements.")
	pflag.Uint64VarP(&numMessages, "number", "n", defaultNumber, "Number of messages in each throughput measurements.")
	return Args{Samples: samples, numMessages: numMessages, config: utils.ParseConfig()}
}
