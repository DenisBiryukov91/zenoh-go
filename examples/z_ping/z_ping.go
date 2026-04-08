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
	"strconv"
	"time"

	"github.com/eclipse-zenoh/zenoh-go/examples/utils"

	"github.com/eclipse-zenoh/zenoh-go/zenoh"

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

	keyexprPing, _ := zenoh.NewKeyExpr("test/ping")
	keyexprPong, _ := zenoh.NewKeyExpr("test/pong")

	sub, err := session.DeclareSubscriber(keyexprPong, zenoh.NewFifoChannel[zenoh.Sample](16), nil)
	if err != nil {
		fmt.Printf("Unable to declare subscriber for key expression '%s': %v\n", keyexprPong, err)
		os.Exit(-1)
	}
	defer sub.Drop()

	pub, err := session.DeclarePublisher(
		keyexprPing,
		&zenoh.PublisherOptions{
			CongestionControl: option.Some(zenoh.CongestionControlBlock),
			IsExpress:         !args.noExpress,
		})
	if err != nil {
		fmt.Printf("Unable to declare publisher for key expression '%s': %v\n", keyexprPing, err)
		os.Exit(-1)
	}
	defer pub.Drop()

	data := make([]byte, args.size)
	for i := range data {
		data[i] = byte(i % 10)
	}
	toSend := zenoh.NewZBytes(data)
	duration := time.Duration(args.warmup * float32(time.Second))
	fmt.Printf("Warming up for %v\n", duration)

	recv := sub.Handler()
	start := time.Now()
	for time.Since(start) < duration {
		pub.Put(toSend, nil)
		<-recv
	}

	ts := []int64{}
	for i := 0; i < int(args.samples); i++ {
		writeTime := time.Now()
		pub.Put(toSend, nil)
		<-recv
		ts = append(ts, time.Since(writeTime).Microseconds())
	}

	for i := 0; i < int(args.samples); i++ {
		fmt.Printf("%v bytes: seq=%v rtt=%vus lat=%vus\n", args.size, i, ts[i], ts[i]/2)
	}
}

const (
	defaultSamples = 100
	defaultWarmup  = 1.0
)

type Args struct {
	size      uint64
	samples   uint32
	warmup    float32
	noExpress bool
	config    zenoh.Config
}

func parseArgs() Args {
	pflag.Usage = printUsage

	var samples uint32
	var warmup float32
	var noExpress bool

	pflag.Uint32VarP(&samples, "samples", "n", defaultSamples, "The number of pings to be attempted.")
	pflag.Float32VarP(&warmup, "warmup", "w", defaultWarmup, "The warmup time in seconds during which pings will be emitted but not measured.")
	pflag.BoolVar(&noExpress, "no-express", false, "Disable message batching.")
	var args Args
	args.config = utils.ParseConfig()
	args.samples = samples
	args.warmup = warmup
	args.noExpress = noExpress

	positional := pflag.Args()
	if len(positional) != 1 {
		printUsage()
		os.Exit(-1)
	}
	var err error
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
