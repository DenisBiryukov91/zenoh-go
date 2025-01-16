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

	queryableKeyexpr, err := zenoh.NewKeyExpr(args.keyexpr)
	if err != nil {
		fmt.Printf("%s is not a valid key expression\n", args.keyexpr)
		os.Exit(-1)
	}

	fmt.Printf("Declaring Queryable on '%s'...\n", queryableKeyexpr)
	queryHandler := func(query zenoh.Query) {
		defer query.Drop()
		keyexpr := query.KeyExpr()
		params := query.Parameters()
		payload := query.Payload()

		if payload.IsSome() && payload.Unwrap().Len() > 0 {
			fmt.Printf(">> [Queryable ] Received Query '%s?%s' with value '%s'\n",
				keyexpr.String(), params, payload.Unwrap())
		} else {
			fmt.Printf(">> [Queryable ] Received Query '%s?%s'\n", keyexpr, params)
		}
		fmt.Printf(">> [Queryable ] Responding ('%s': '%s')\n", queryableKeyexpr, args.payload)
		query.Reply(queryableKeyexpr, zenoh.NewZBytesFromString(args.payload), nil)
	}

	opts := zenoh.QueryableOptions{}
	opts.Complete = args.complete
	queryable, err := session.DeclareQueryable(queryableKeyexpr, queryHandler, nil, &opts)
	if err != nil {
		fmt.Println("Unable to declare queryable.")
		os.Exit(-1)
	}
	defer queryable.Drop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")
	<-stop
}

type Args struct {
	keyexpr  string
	payload  string
	complete bool
	config   zenoh.Config
}

const (
	defaultKeyexpr = "demo/example/zenoh-go-queryable"
	defaultValue   = "Queryable from Go!"
)

func parseArgs() Args {
	var keyexpr string
	var payload string
	var complete bool

	pflag.StringVarP(&keyexpr, "key", "k", defaultKeyexpr, "The key expression matching queries to reply to.")
	pflag.StringVarP(&payload, "payload", "p", defaultValue, "The value to reply to queries with.")
	pflag.BoolVar(&complete, "complete", false, "Indicates whether queryable is complete or not.")

	return Args{keyexpr: keyexpr, payload: payload, complete: complete, config: utils.ParseConfig()}
}
