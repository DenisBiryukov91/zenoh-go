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
		fmt.Printf("Failed to open Zenoh session: %v\n", err)
		os.Exit(-1)
	}
	defer session.Drop()

	keyexpr, err := zenoh.NewKeyExpr(args.keyexpr)
	if err != nil {
		fmt.Printf("%s is not a valid key expression\n", args.keyexpr)
		os.Exit(-1)
	}

	storage := make(map[string]zenoh.Sample)

	subHandler := func(sample zenoh.Sample) {
		fmt.Printf(">> [Subscriber] Received %s ('%s': '%s')\n",
			kindToStr(sample.Kind()),
			sample.KeyExpr(),
			sample.Payload())

		switch sample.Kind() {
		case zenoh.SampleKindPut:
			storage[keyexpr.String()] = sample
		case zenoh.SampleKindDelete:
			delete(storage, keyexpr.String())
		}
	}

	fmt.Printf("Declaring Subscriber on '%s'...\n", keyexpr)
	sub, err := session.DeclareSubscriber(keyexpr, subHandler, nil, nil)
	if err != nil {
		fmt.Println("Unable to declare subscriber.")
		os.Exit(-1)
	}
	defer sub.Drop()

	opts := zenoh.QueryableOptions{}
	opts.Complete = args.complete
	queryableHandler := func(query zenoh.Query) {
		defer query.Drop()
		for _, s := range storage {
			if s.KeyExpr().Intersects(query.KeyExpr()) {
				query.Reply(s.KeyExpr(), s.Payload(), nil)
			}
		}
	}
	queryable, err := session.DeclareQueryable(keyexpr, queryableHandler, nil, &opts)
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
	keyexpr  string
	complete bool
	config   zenoh.Config
}

func parseArgs() Args {
	var keyexpr string
	var complete bool

	pflag.StringVarP(&keyexpr, "key", "k", defaultKeyexpr, "The selection of resources to store.")
	pflag.BoolVar(&complete, "complete", false, "Indicates whether storage is complete or not.")
	return Args{keyexpr: keyexpr, complete: complete, config: utils.ParseConfig()}
}
