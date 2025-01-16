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
	"zenoh-go/zenoh"
)

func main() {
	zenoh.InitLoggerFromEnvOr("error")

	fmt.Println("Scouting...")
	// Create FIFO channel
	hellos := make(chan zenoh.Hello, 16)
	zenoh.Scout(zenoh.NewConfigDefault(),
		func(hello zenoh.Hello) { hellos <- hello },
		func() { close(hellos) },
		&zenoh.ScoutOptions{TimeoutMs: 1000})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	for {
		select {
		case <-stop:
			return
		case hello, ok := <-hellos:
			if !ok {
				return
			}
			fmt.Printf("%s\n", hello)
		}
	}
}
