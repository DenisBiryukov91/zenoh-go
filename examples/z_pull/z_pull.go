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

	"github.com/eclipse-zenoh/zenoh-go/examples/utils"

	"github.com/eclipse-zenoh/zenoh-go/zenoh"

	"github.com/spf13/pflag"
)

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

	fmt.Printf("Declaring Subscriber on '%s'...\n", keyexpr)
	sub, err := session.DeclareSubscriber(keyexpr, zenoh.NewRingChannel[zenoh.Sample](int(args.size)), nil)
	if err != nil {
		fmt.Printf("Unable to declare subscriber for key expression '%s': %v\n", keyexpr, err)
		os.Exit(-1)
	}
	defer sub.Drop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")

	for {
		select {
		case <-stop:
			{
				return
			}
		case sample := <-sub.Handler():
			{
				fmt.Printf(">> [Subscriber] Pulled %s ('%s': '%s')... performing a computation of %vs\n",
					kindToStr(sample.Kind()),
					sample.KeyExpr().String(),
					sample.Payload().String(),
					args.interval)
				time.Sleep(time.Duration(args.interval * float32(time.Second)))
			}
		}
	}
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

const defaultKeyexpr = "demo/example/**"
const defaultSize = 3
const defaultInterval = 5.0

type Args struct {
	keyexpr  string
	size     uint32
	interval float32
	config   zenoh.Config
}

func parseArgs() Args {
	var keyexpr string
	var size uint32
	var interval float32
	pflag.StringVarP(&keyexpr, "key", "k", defaultKeyexpr, "The key expression to subscribe to.")
	pflag.Uint32VarP(&size, "size", "s", defaultSize, "The size of the ringbuffer.")
	pflag.Float32VarP(&interval, "interval", "i", defaultInterval, "The interval for pulling the ringbuffer in seconds.")
	return Args{keyexpr: keyexpr, size: size, interval: interval, config: utils.ParseConfig()}
}
