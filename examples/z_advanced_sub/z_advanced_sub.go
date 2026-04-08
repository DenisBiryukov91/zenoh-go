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

	"github.com/eclipse-zenoh/zenoh-go/examples/utils"

	"github.com/eclipse-zenoh/zenoh-go/zenoh"
	"github.com/eclipse-zenoh/zenoh-go/zenoh/zenohext"

	"github.com/BooleanCat/option"
	"github.com/spf13/pflag"
)

func dataHandler(sample zenoh.Sample) {
	fmt.Printf(">> [Subscriber] Received %s ('%s': '%s')\n",
		kindToStr(sample.Kind()),
		sample.KeyExpr().String(),
		sample.Payload().String())
}

func missHandler(miss zenohext.Miss) {
	fmt.Printf(">> [Subscriber] Missed %d samples from '{%s}/{%d}' !!!\n",
		miss.Nb,
		miss.Source.ZId(),
		miss.Source.EntityId())
}

func kindToStr(kind zenoh.SampleKind) string {
	switch kind {
	case zenoh.SampleKindPut:
		return "PUT"
	case zenoh.SampleKindDelete:
		return "DELETE"
	default:
		return "UNKNOWN"
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

	keyexpr, err := zenoh.NewKeyExpr(args.keyexpr)
	if err != nil {
		fmt.Printf("%s is not a valid key expression: %v\n", args.keyexpr, err)
		os.Exit(-1)
	}

	subOpts := zenohext.AdvancedSubscriberOptions{
		History: option.Some(zenohext.AdvancedSubscriberHistoryOptions{
			DetectLatePublishers: true,
		}),
		Recovery: option.Some(zenohext.AdvancedSubscriberRecoveryOptions{
			LastSampleMissDetection: zenohext.LastSampleMissDetectionModeHeartbeat(),
		}),
		SubscriberDetection: true,
	}

	fmt.Printf("Declaring AdvancedSubscriber on '%s'...\n", keyexpr)
	sub, err := zenohext.Ext(&session).DeclareAdvancedSubscriber(
		keyexpr,
		zenoh.Closure[zenoh.Sample]{Call: dataHandler},
		&subOpts,
	)
	if err != nil {
		fmt.Println("Unable to declare AdvancedSubscriber.")
		os.Exit(-1)
	}
	defer sub.Drop()

	if err := sub.DeclareBackgroundSampleMissListener(zenoh.Closure[zenohext.Miss]{Call: missHandler}); err != nil {
		fmt.Printf("Unable to declare background sample miss listener: %v\n", err)
		os.Exit(-1)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")
	<-stop
}

const defaultKeyexpr = "demo/example/**"

type Args struct {
	keyexpr string
	config  zenoh.Config
}

func parseArgs() Args {
	var keyexpr string

	pflag.StringVarP(&keyexpr, "key", "k", defaultKeyexpr, "The key expression to subscribe to.")

	return Args{
		keyexpr: keyexpr,
		config:  utils.ParseConfig(),
	}
}
