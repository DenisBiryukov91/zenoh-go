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
	"zenoh-go/examples/utils"
	"zenoh-go/zenoh"

	"github.com/BooleanCat/option"
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

	selectorKey, selectorParams := utils.ParseSelector(args.selector)

	keyExpr, err := zenoh.NewKeyExpr(selectorKey)
	if err != nil {
		fmt.Printf("%s is not a valid key expression: %v\n", selectorKey, err)
		os.Exit(-1)
	}

	fmt.Printf("Sending Query '%s'...\n", args.selector)

	// Set get options
	opts := zenoh.GetOptions{}
	if args.payload != "" {
		opts.Payload = option.Some(zenoh.NewZBytesFromString(args.payload))
	}
	opts.TimeoutMs = args.timeout
	opts.Target = option.Some(args.queryTarget)

	// send Query
	replies, _ := session.Get(
		keyExpr,
		selectorParams,
		zenoh.NewFifoChannel[zenoh.Reply](16),
		&opts)

	for reply := range replies {
		if reply.IsOk() {
			sample := reply.Ok().Unwrap()
			fmt.Printf(">> Received ('%s': '%s')\n",
				sample.KeyExpr(),
				sample.Payload())
		} else {
			fmt.Println("Received an error")
		}
	}
}

type Args struct {
	selector    string
	payload     string
	timeout     uint64
	queryTarget zenoh.QueryTarget
	config      zenoh.Config
}

const (
	defaultSelector = "demo/example/**"
	defaultTimeout  = 10000
	defaultValue    = ""
)

func parseArgs() Args {
	var selector string
	var payload string
	var timeout uint64
	var queryTargetString string

	pflag.StringVarP(&selector, "selector", "s", defaultSelector, "The selection of resources to query.")
	pflag.StringVarP(&payload, "payload", "p", defaultValue, "An optional value to put in the query.")
	pflag.StringVarP(&queryTargetString, "target", "t", "BEST_MATCHING", "Query target (BEST_MATCHING | ALL | ALL_COMPLETE).")
	pflag.Uint64VarP(&timeout, "timeout", "o", 10000, "Query timeout in milliseconds.")

	config := utils.ParseConfig()
	queryTarget, err := utils.ParseQueryTarget(queryTargetString)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	return Args{selector: selector, payload: payload, queryTarget: queryTarget, timeout: timeout, config: config}
}
