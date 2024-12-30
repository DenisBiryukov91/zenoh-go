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

/*
#cgo LDFLAGS: -lzenohcd
#include "zenoh.h"
#include <stdlib.h>
*/
import "C"

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"unsafe"
	"zenoh-go/examples/utils"
)

const (
	defaultKeyexpr = "demo/example/zenoh-go-pub"
	defaultValue   = "Pub from Go!"
)

type Args struct {
	keyexpr string
	value   string
	common  utils.CommonArgs
}

func parseArgs() Args {
	var (
		keyexpr string
		value   string
	)

	flag.StringVar(&keyexpr, "k", defaultKeyexpr, "The key expression to publish to")
	flag.StringVar(&value, "p", defaultValue, "The value to publish")

	return Args{
		keyexpr: keyexpr,
		value:   value,
		common:  utils.ParseCommonArgs(),
	}
}

func main() {
	pinner := &runtime.Pinner{}
	defer pinner.Unpin()

	args := parseArgs()

	logLevel := C.CString("error")
	defer C.free(unsafe.Pointer(logLevel))
	C.zc_init_log_from_env_or(logLevel)

	var config C.z_owned_config_t
	utils.ConfigFromArgs((*utils.ZConfig)(unsafe.Pointer(&config)), &args.common)

	fmt.Println("Opening session...")
	var session C.z_owned_session_t
	if C.z_open(&session, C.z_config_move(&config), nil) < 0 {
		fmt.Println("Unable to open session!")
		os.Exit(-1)
	}
	defer C.z_session_drop(C.z_session_move(&session))

	fmt.Printf("Declaring Publisher on '%s'...\n", args.keyexpr)
	keyExprC := C.CString(args.keyexpr)
	defer C.free(unsafe.Pointer(keyExprC))
	var ke C.z_view_keyexpr_t
	C.z_view_keyexpr_from_str(&ke, keyExprC)

	var pub C.z_owned_publisher_t
	if C.z_declare_publisher(C.z_session_loan(&session), &pub, C.z_view_keyexpr_loan(&ke), nil) < 0 {
		fmt.Println("Unable to declare Publisher for key expression!")
		os.Exit(-1)
	}
	defer C.z_publisher_drop(C.z_publisher_move(&pub))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	idx := 0
	for {
		select {
		case <-stop:
			return
		default:
			C.z_sleep_s(1)
			message := fmt.Sprintf("[%4d] %s", idx, args.value)
			fmt.Printf("Putting Data ('%s': '%s')...\n", args.keyexpr, message)

			payload := C.z_owned_bytes_t{}
			payloadStr := C.CString(message)
			defer C.free(unsafe.Pointer(payloadStr))
			C.z_bytes_copy_from_str(&payload, payloadStr)

			options := C.z_publisher_put_options_t{}
			C.z_publisher_put_options_default(&options)

			var encoding C.z_owned_encoding_t
			pinner.Pin(&encoding)
			C.z_encoding_clone(&encoding, C.z_encoding_text_plain())
			options.encoding = C.z_encoding_move(&encoding)

			C.z_publisher_put(C.z_publisher_loan(&pub), C.z_bytes_move(&payload), &options)
			idx++
		}
	}
}
