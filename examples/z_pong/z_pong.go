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

	"github.com/BooleanCat/option"
	"github.com/spf13/pflag"
)

func main() {
	zenoh.InitLoggerFromEnvOr("error")
	args := parseArgs()

	fmt.Println("Opening session...")
	session, _ := zenoh.Open(args.config, nil)
	defer session.Drop()

	keyexprPing, _ := zenoh.NewKeyExpr("test/ping")
	keyexprPong, _ := zenoh.NewKeyExpr("test/pong")

	pub, _ := session.DeclarePublisher(
		keyexprPong,
		&zenoh.PublisherOptions{
			CongestionControl: option.Some(zenoh.CongestionControlBlock),
			IsExpress:         !args.noExpress,
		})
	defer pub.Drop()

	session.DeclareBackgroundSubscriber(
		keyexprPing,
		zenoh.Closure[zenoh.Sample]{Call: func(sample zenoh.Sample) { pub.Put(sample.Payload(), nil) }},
		nil)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")
	<-stop
}

type Args struct {
	noExpress bool
	config    zenoh.Config
}

func parseArgs() Args {
	var noExpress bool

	pflag.BoolVar(&noExpress, "express", false, "Disable message batching.")
	var args Args
	args.config = utils.ParseConfig()
	args.noExpress = noExpress

	return args
}
