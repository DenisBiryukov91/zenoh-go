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
#cgo LDFLAGS: -lzenohc
#include "zenoh.h"
*/
import "C"

import (
	"flag"
	"fmt"
	"os"
	"unsafe"
	"zenoh-go/examples/utils"
)

const (
	defaultKeyExpr = "group1/**"
	defaultTimeout = 10000
)

type Args struct {
	keyexpr   string
	timeoutMs uint64
	common    utils.CommonArgs
}

func parseArgs() Args {
	var (
		keyexpr   string
		timeoutMs uint64
	)

	flag.StringVar(&keyexpr, "k", defaultKeyExpr, "The key expression to query")
	flag.Uint64Var(&timeoutMs, "o", defaultTimeout, "Query timeout in milliseconds")
	flag.Parse()

	return Args{
		keyexpr:   keyexpr,
		timeoutMs: timeoutMs,
		common:    utils.ParseCommonArgs(),
	}
}

func main() {
	args := parseArgs()

	logLevel := C.CString("error")
	defer C.free(unsafe.Pointer(logLevel))
	C.zc_init_log_from_env_or(logLevel)

	var config C.z_owned_config_t
	utils.ConfigFromArgs((*utils.ZConfig)(unsafe.Pointer(&config)), &args.common)

	// Open session
	var session C.z_owned_session_t
	if C.z_open(&session, C.z_config_move(&config), nil) < 0 {
		fmt.Println("Unable to open session!")
		os.Exit(-1)
	}
	defer C.z_session_drop(C.z_session_move(&session))

	// Create key expression
	keyExprC := C.CString(args.keyexpr)
	defer C.free(unsafe.Pointer(keyExprC))

	var keyExpr C.z_view_keyexpr_t
	if C.z_view_keyexpr_from_str(&keyExpr, keyExprC) < 0 {
		fmt.Printf("%s is not a valid key expression\n", args.keyexpr)
		os.Exit(-1)
	}

	fmt.Printf("Sending liveliness query '%s'...\n", args.keyexpr)

	// Create FIFO handler and closure
	var handler C.z_owned_fifo_handler_reply_t
	var closure C.z_owned_closure_reply_t
	C.z_fifo_channel_reply_new(&closure, &handler, 16)

	// Set liveliness get options
	var opts C.z_liveliness_get_options_t
	C.z_liveliness_get_options_default(&opts)
	opts.timeout_ms = C.uint64_t(args.timeoutMs)

	// Send liveliness query
	C.z_liveliness_get(C.z_session_loan(&session), C.z_view_keyexpr_loan(&keyExpr), C.z_closure_reply_move(&closure), &opts)

	// Receive replies
	for {
		var reply C.z_owned_reply_t
		res := C.z_fifo_handler_reply_recv(C.z_fifo_handler_reply_loan(&handler), &reply)
		if res != C.Z_OK {
			break
		}

		if C.z_reply_is_ok(C.z_reply_loan(&reply)) {
			sample := C.z_reply_ok(C.z_reply_loan(&reply))

			var keyStr C.z_view_string_t
			C.z_keyexpr_as_view_string(C.z_sample_keyexpr(sample), &keyStr)

			fmt.Printf(">> Alive token ('%s')\n",
				C.GoStringN(C.z_string_data(C.z_view_string_loan(&keyStr)), C.int(C.z_string_len(C.z_view_string_loan(&keyStr)))))
		} else {
			fmt.Println("Received an error")
		}
		C.z_reply_drop(C.z_reply_move(&reply))
	}

	C.z_fifo_handler_reply_drop(C.z_fifo_handler_reply_move(&handler))
	C.z_session_drop(C.z_session_move(&session))
}
