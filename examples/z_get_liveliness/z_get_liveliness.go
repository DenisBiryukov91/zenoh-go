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

	keyExpr, err := zenoh.NewKeyExpr(args.keyexpr)
	if err != nil {
		fmt.Printf("%s is not a valid key expression: %v\n", args.keyexpr, err)
		os.Exit(-1)
	}

	fmt.Printf("Sending liveliness query '%s'...\n", args.keyexpr)

	// send Query
	replies, _ := session.Liveliness().Get(
		keyExpr,
		zenoh.NewFifoChannel[zenoh.Reply](16),
		&zenoh.LivelinessGetOptions{TimeoutMs: args.timeout})

	for reply := range replies {
		if reply.IsOk() {
			sample := reply.Ok().Unwrap()
			fmt.Printf(">> Alive token ('%s')\n", sample.KeyExpr())
		} else {
			fmt.Println("Received an error")
		}
	}
}

type Args struct {
	keyexpr string
	timeout uint64
	config  zenoh.Config
}

const (
	defaultKeyexpr = "group1/**"
	defaultTimeout = 10000
)

func parseArgs() Args {
	var keyexpr string
	var timeout uint64

	pflag.StringVarP(&keyexpr, "selector", "s", defaultKeyexpr, "The key expression to query")
	pflag.Uint64VarP(&timeout, "timeout", "o", 10000, "Query timeout in milliseconds.")

	return Args{keyexpr: keyexpr, timeout: timeout, config: utils.ParseConfig()}
}
