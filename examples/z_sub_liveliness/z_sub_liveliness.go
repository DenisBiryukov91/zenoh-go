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
	"zenoh-go/examples/utils"
	"zenoh-go/zenoh"

	"github.com/spf13/pflag"
)

func dataHandler(sample zenoh.Sample) {
	switch sample.Kind() {
	case zenoh.SampleKindPut:
		fmt.Printf(">> [LivelinessSubscriber] New alive token ('%s')\n", sample.KeyExpr())
	case zenoh.SampleKindDelete:
		fmt.Printf(">> [LivelinessSubscriber] Dropped token ('%s')\n", sample.KeyExpr())
	}
}

func main() {
	zenoh.InitLoggerFromEnvOr("error")
	args := parseArgs()

	fmt.Println("Opening session...")
	session, err := zenoh.Open(args.config, nil)
	if err != nil {
		fmt.Println("Failed to open Zenoh session")
		os.Exit(-1)
	}
	defer session.Drop()

	keyExpr, err := zenoh.NewKeyExpr(args.keyexpr)
	if err != nil {
		fmt.Printf("%s is not a valid key expression\n", args.keyexpr)
		os.Exit(-1)
	}

	fmt.Printf("Declaring liveliness subscriber on '%s'...\n", args.keyexpr)
	sub, err := session.Liveliness().DeclareSubscriber(keyExpr, dataHandler, nil, &zenoh.LivelinessSubscriberOptions{History: args.history})
	if err != nil {
		fmt.Println("Unable to declare liveliness subscriber.")
		os.Exit(-1)
	}
	defer sub.Drop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")
	<-stop
}

const defaultKeyexpr = "group1/**"

type Args struct {
	keyexpr string
	history bool
	config  zenoh.Config
}

func parseArgs() Args {
	var keyexpr string
	var history bool

	pflag.StringVarP(&keyexpr, "key", "k", defaultKeyexpr, "The key expression matching liveliness tokens to subscribe to.")
	pflag.BoolVar(&history, "history", false, "Get historical liveliness tokens.")

	return Args{keyexpr: keyexpr, history: history, config: utils.ParseConfig()}
}
