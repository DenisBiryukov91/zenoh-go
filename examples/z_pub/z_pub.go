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
		fmt.Println("Publisher has matching subscribers.")
	} else {
		fmt.Println("Publisher has NO MORE matching subscribers.")
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

	keyexpr, err := zenoh.NewKeyExpr(args.keyexpr)
	if err != nil {
		fmt.Printf("%s is not a valid key expression: %v\n", args.keyexpr, err)
		os.Exit(-1)
	}

	fmt.Printf("Declaring Publisher on '%s'...\n", keyexpr)
	pub, err := session.DeclarePublisher(keyexpr, nil)
	if err != nil {
		fmt.Printf("Unable to declare Publisher for key expression '%s': %v\n", keyexpr, err)
		os.Exit(-1)
	}
	defer pub.Drop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")

	if args.addMatchingListener {
		pub.DeclareBackgroundMatchingListener(zenoh.Closure[zenoh.MatchingStatus]{Call: matchingStatusCallback})
	}

	idx := 0
	for {
		select {
		case <-stop:
			return
		default:
			time.Sleep(time.Second)
			message := fmt.Sprintf("[%4d] %s", idx, args.payload)
			fmt.Printf("Putting Data ('%s': '%s')...\n", keyexpr, message)

			putOpts := zenoh.PublisherPutOptions{}
			if len(args.attachment) != 0 {
				putOpts.Attachement = option.Some(zenoh.NewZBytesFromString(args.attachment))
			}
			if err := pub.Put(zenoh.NewZBytesFromString(message), &putOpts); err != nil {
				fmt.Printf("Failed to put data: %v\n", err)
			}
			idx++
		}
	}
}

const (
	defaultKeyexpr    = "demo/example/zenoh-go-pub"
	defaultValue      = "Pub from Go!"
	defaultAttachment = ""
)

type Args struct {
	keyexpr             string
	payload             string
	attachment          string
	addMatchingListener bool
	config              zenoh.Config
}

func parseArgs() Args {
	var keyexpr string
	var payload string
	var attachment string
	var addMatchingListener bool

	pflag.StringVarP(&keyexpr, "key", "k", defaultKeyexpr, "The key expression to publish to.")
	pflag.StringVarP(&payload, "payload", "p", defaultValue, "The value to publish.")
	pflag.StringVarP(&attachment, "attach", "a", defaultAttachment, "The attachment to add to each put.")
	pflag.BoolVar(&addMatchingListener, "add-matching-listener", false, "Add matching listener.")

	return Args{
		keyexpr:             keyexpr,
		payload:             payload,
		attachment:          attachment,
		addMatchingListener: addMatchingListener,
		config:              utils.ParseConfig()}
}
