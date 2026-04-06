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

	"github.com/BooleanCat/option"
	"github.com/spf13/pflag"
)

func matchingStatusCallback(status zenoh.MatchingStatus) {
	if status.Matching {
		fmt.Println("Querier has matching queryables.")
	} else {
		fmt.Println("Queerier has NO MORE matching queryables.")
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

	selectorKey, selectorParams := utils.ParseSelector(args.selector)

	keyExpr, err := zenoh.NewKeyExpr(selectorKey)
	if err != nil {
		fmt.Printf("%s is not a valid key expression: %v\n", selectorKey, err)
		os.Exit(-1)
	}

	fmt.Printf("Declaring Querier on '%s'...\n", keyExpr)
	querier, err := session.DeclareQuerier(
		keyExpr,
		&zenoh.QuerierOptions{Target: option.Some(args.queryTarget), TimeoutMs: args.timeout})

	if err != nil {
		fmt.Printf("Unable to declare Querier for key expression '%s': %v\n", keyExpr, err)
		os.Exit(-1)
	}
	defer querier.Drop()

	if args.addMatchingListener {
		querier.DeclareBackgroundMatchingListener(zenoh.Closure[zenoh.MatchingStatus]{Call: matchingStatusCallback})
	}

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
			payload := fmt.Sprintf("[%4d] %s", idx, args.payload)

			fmt.Printf("Querying '%s' with payload '%s'...\n", args.selector, payload)

			getOpts := zenoh.QuerierGetOptions{}
			getOpts.Payload = option.Some(zenoh.NewZBytesFromString(payload))

			// send Query
			replies, err := querier.Get(
				selectorParams,
				zenoh.NewFifoChannel[zenoh.Reply](16),
				&getOpts)

			if err != nil {
				fmt.Printf("Failed to send query: %v\n", err)
			}

			for reply := range replies {
				if reply.IsOk() {
					sample := reply.Ok().Unwrap()
					fmt.Printf(">> Received ('%s': '%s')\n",
						sample.KeyExpr(),
						sample.Payload())
				} else {
					err := reply.Err().Unwrap()
					fmt.Printf("Received (ERROR: '%s')\n", err.Payload())
				}
			}

			idx++
		}
	}
}

type Args struct {
	selector            string
	payload             string
	timeout             uint64
	queryTarget         zenoh.QueryTarget
	addMatchingListener bool
	config              zenoh.Config
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
	var addMatchingListener bool

	pflag.StringVarP(&selector, "selector", "s", defaultSelector, "The selection of resources to query.")
	pflag.StringVarP(&payload, "payload", "p", defaultValue, "An optional value to put in the query.")
	pflag.StringVarP(&queryTargetString, "target", "t", "BEST_MATCHING", "Query target (BEST_MATCHING | ALL | ALL_COMPLETE).")
	pflag.Uint64VarP(&timeout, "timeout", "o", 10000, "Query timeout in milliseconds.")
	pflag.BoolVar(&addMatchingListener, "add-matching-listener", false, "Add matching listener.")

	config := utils.ParseConfig()
	queryTarget, err := utils.ParseQueryTarget(queryTargetString)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	return Args{
		selector:            selector,
		payload:             payload,
		queryTarget:         queryTarget,
		timeout:             timeout,
		addMatchingListener: addMatchingListener,
		config:              config}
}
