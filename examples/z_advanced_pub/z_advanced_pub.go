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

package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/eclipse-zenoh/zenoh-go/examples/utils"

	"github.com/eclipse-zenoh/zenoh-go/zenoh"
	"github.com/eclipse-zenoh/zenoh-go/zenoh/zenohext"

	"github.com/BooleanCat/option"
	"github.com/spf13/pflag"
)

func main() {
	zenoh.InitLoggerFromEnvOr("error")
	args := parseArgs()

	if args.config.InsertJson5("timestamping/enabled", "true") != nil {
		fmt.Println("Failed to enable timestamping in the configuration")
		os.Exit(-1)
	}

	fmt.Println("Opening session...")
	session, err := zenoh.Open(args.config, nil)
	if err != nil {
		fmt.Printf("Failed to open Zenoh session: %v\n", err)
		os.Exit(-1)
	}
	defer session.Drop()

	keyexpr, err := zenoh.NewKeyExpr(args.keyexpr)
	if err != nil {
		fmt.Printf("%s is not a valid key expression: %v\n", args.keyexpr, err)
		os.Exit(-1)
	}

	fmt.Printf("Declaring AdvancedPublisher on '%s'...\n", keyexpr)
	pubOpts := zenohext.AdvancedPublisherOptions{
		Cache: option.Some(zenohext.AdvancedPublisherCacheOptions{
			MaxSamples: args.history,
		}),
		PublisherDetection: true,
		SampleMissDetection: option.Some(zenohext.AdvancedPublisherSampleMissDetectionOptions{
			HeartbeatMode: zenohext.HeartbeatModePeriodic(500),
		}),
	}
	pub, err := zenohext.Ext(&session).DeclareAdvancedPublisher(keyexpr, &pubOpts)
	if err != nil {
		fmt.Println("Unable to declare AdvancedPublisher for key expression!")
		os.Exit(-1)
	}
	defer pub.Drop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")

	idx := 0
	for {
		select {
		case <-stop:
			return
		default:
			time.Sleep(time.Second)
			message := fmt.Sprintf("[%4d] %s", idx, args.payload)
			fmt.Printf("Putting Data ('%s': '%s')...\n", keyexpr, message)
			if err := pub.Put(zenoh.NewZBytesFromString(message), nil); err != nil {
				fmt.Printf("Failed to put data: %v\n", err)
			}
			idx++
		}
	}
}

const (
	defaultKeyexpr = "demo/example/zenoh-go-pub"
	defaultValue   = "Pub from Go!"
	defaultHistory = 1
)

type Args struct {
	keyexpr string
	payload string
	history uint
	config  zenoh.Config
}

func parseArgs() Args {
	var keyexpr string
	var payload string
	var history uint

	pflag.StringVarP(&keyexpr, "key", "k", defaultKeyexpr, "The key expression to publish to.")
	pflag.StringVarP(&payload, "payload", "p", defaultValue, "The value to publish.")
	pflag.UintVarP(&history, "history", "i", defaultHistory, "The number of publications to keep in cache.")

	return Args{
		keyexpr: keyexpr,
		payload: payload,
		history: history,
		config:  utils.ParseConfig(),
	}
}
