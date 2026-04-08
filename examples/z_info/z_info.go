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

	fmt.Printf("own id: %s\n", session.ZId())

	fmt.Println("routers ids:")
	ids, _ := session.RoutersZId()
	for _, id := range ids {
		fmt.Printf("%s\n", id)
	}

	fmt.Printf("peers ids:\n")
	ids, _ = session.PeersZId()
	for _, id := range ids {
		fmt.Printf("%s\n", id)
	}
}

type Args struct {
	config zenoh.Config
}

func parseArgs() Args {
	return Args{config: utils.ParseConfig()}
}
