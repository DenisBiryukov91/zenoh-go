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

	// Declare liveliness token
	fmt.Printf("Declaring liveliness token '%s'...\n", args.keyexpr)
	token, err := session.Liveliness().DeclareToken(keyExpr, nil)
	if err != nil {
		fmt.Println("Unable to create liveliness token!")
		os.Exit(-1)
	}
	defer token.Drop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")
	<-stop

	fmt.Println("Undeclaring liveliness token...")
}

const defaultKeyexpr = "group1/**"

type Args struct {
	keyexpr string
	config  zenoh.Config
}

func parseArgs() Args {
	var keyexpr string

	pflag.StringVarP(&keyexpr, "key", "k", defaultKeyexpr, "The key expression for the liveliness token.")

	return Args{keyexpr: keyexpr, config: utils.ParseConfig()}
}
