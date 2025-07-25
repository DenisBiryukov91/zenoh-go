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
	"strconv"
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

	keyexpr, _ := zenoh.NewKeyExpr("test/thr")

	opts := zenoh.PublisherOptions{Priority: option.Some(args.priority), IsExpress: args.isExpress}
	fmt.Printf("Declaring Publisher on '%s'...\n", keyexpr)
	pub, err := session.DeclarePublisher(keyexpr, &opts)
	if err != nil {
		fmt.Println("Unable to declare Publisher for key expression!")
		os.Exit(-1)
	}
	defer pub.Drop()

	data := make([]byte, args.size)
	for i := range data {
		data[i] = byte(i % 10)
	}
	toSend := zenoh.NewZBytes(data)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")

	for {
		select {
		case <-stop:
			return
		default:
			pub.Put(toSend, nil)
		}
	}
}

const (
	defaultPriority = zenoh.PriorityDefault
)

type Args struct {
	size      uint64
	priority  zenoh.Priority
	isExpress bool
	config    zenoh.Config
}

func parseArgs() Args {
	pflag.Usage = printUsage

	var priorityValue uint8
	var isExpress bool

	pflag.Uint8VarP(&priorityValue, "priority", "p", uint8(defaultPriority), fmt.Sprintf("Priority for sending data [%d - %d].", int(zenoh.PriorityRealTime), int(zenoh.PriorityBackground)))
	pflag.BoolVar(&isExpress, "express", false, "Disable message batching")
	var args Args
	args.config = utils.ParseConfig()
	args.isExpress = isExpress

	priority, err := utils.ParsePriority(priorityValue)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	args.priority = priority
	positional := pflag.Args()
	if len(positional) != 1 {
		printUsage()
		os.Exit(-1)
	}
	args.size, err = strconv.ParseUint(positional[0], 0, 0)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	return args
}

func printUsage() {
	fmt.Printf("Usage: %s [OPTIONS] <PAYLOAD_SIZE>\nOptions:\n", os.Args[0])
	pflag.PrintDefaults()
}
