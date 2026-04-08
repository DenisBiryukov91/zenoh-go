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

	"github.com/eclipse-zenoh/zenoh-go/examples/utils"

	"github.com/eclipse-zenoh/zenoh-go/zenoh"

	"github.com/spf13/pflag"
)

func dataHandler(sample zenoh.Sample) {
	fmt.Printf(">> [Subscriber] Received %s ('%s': '%s')",
		kindToStr(sample.Kind()),
		sample.KeyExpr().String(),
		sample.Payload().String())

	// check if attachment exists
	if sample.Attachement().IsSome() {
		fmt.Printf(" (%s)", sample.Attachement().Unwrap().String())
	}
	fmt.Print("\n")
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

	fmt.Printf("Declaring Subscriber on '%s'...\n", keyexpr)
	sub, err := session.DeclareSubscriber(keyexpr, zenoh.Closure[zenoh.Sample]{Call: dataHandler}, nil)
	if err != nil {
		fmt.Printf("Unable to declare subscriber for key expression '%s': %v\n", keyexpr, err)
		os.Exit(-1)
	}
	defer sub.Drop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")
	<-stop
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

type Args struct {
	keyexpr string
	config  zenoh.Config
}

func parseArgs() Args {
	var keyexpr string
	pflag.StringVarP(&keyexpr, "key", "k", defaultKeyexpr, "The key expression to subscribe to.")
	return Args{keyexpr: keyexpr, config: utils.ParseConfig()}
}
