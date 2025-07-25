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

	fmt.Printf("Deleting resources matching '%s'...\n", keyexpr)
	if err := session.Delete(keyexpr, nil); err != nil {
		fmt.Printf("Delete failed: %v\n", err)
	}
}

const (
	defaultKeyexpr = "demo/example/zenoh-go-put"
)

type Args struct {
	keyexpr string
	config  zenoh.Config
}

func parseArgs() Args {
	var keyexpr string

	pflag.StringVarP(&keyexpr, "key", "k", defaultKeyexpr, "The key expression to publish to.")

	return Args{keyexpr: keyexpr, config: utils.ParseConfig()}
}
