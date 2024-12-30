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

void queryHandler(struct z_loaned_query_t *query, void *context);
*/
import "C"

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"unsafe"
	"zenoh-go/examples/utils"
)

const (
	defaultKeyexpr = "demo/example/zenoh-go-queryable"
	defaultValue   = "Queryable from Go!"
)

type Args struct {
	keyexpr  string
	value    string
	complete bool
	common   utils.CommonArgs
}

var value string

func parseArgs() Args {
	var (
		keyexpr  string
		value    string
		complete bool
	)

	flag.StringVar(&keyexpr, "k", defaultKeyexpr, "The key expression matching queries to reply to")
	flag.StringVar(&value, "p", defaultValue, "The value to reply to queries with")
	flag.BoolVar(&complete, "complete", false, "Flag to indicate whether queryable is complete or not")

	return Args{
		keyexpr:  keyexpr,
		value:    value,
		complete: complete,
		common:   utils.ParseCommonArgs(),
	}
}

//export queryHandler
func queryHandler(query *C.z_loaned_query_t, _ unsafe.Pointer) {
	var keyString C.z_view_string_t
	C.z_keyexpr_as_view_string(C.z_query_keyexpr(query), &keyString)

	var params C.z_view_string_t
	C.z_query_parameters(query, &params)

	keyStringGo := C.GoStringN(C.z_string_data(C.z_view_string_loan(&keyString)), (C.int)(C.z_string_len(C.z_view_string_loan(&keyString))))
	paramsGo := C.GoStringN(C.z_string_data(C.z_view_string_loan(&params)), (C.int)(C.z_string_len(C.z_view_string_loan(&params))))

	payload := C.z_query_payload(query)
	if payload != nil && C.z_bytes_len(payload) > 0 {
		var payloadString C.z_owned_string_t
		C.z_bytes_to_string(payload, &payloadString)
		defer C.z_string_drop(C.z_string_move(&payloadString))

		fmt.Printf(">> [Queryable ] Received Query '%s?%s' with value '%s'\n",
			keyStringGo, paramsGo,
			C.GoStringN(C.z_string_data(C.z_string_loan(&payloadString)), (C.int)(C.z_string_len(C.z_string_loan(&payloadString)))),
		)
	} else {
		fmt.Printf(">> [Queryable ] Received Query '%s?%s'\n", keyStringGo, paramsGo)
	}

	var options C.z_query_reply_options_t
	C.z_query_reply_options_default(&options)

	var replyPayload C.z_owned_bytes_t
	C.z_bytes_from_static_str(&replyPayload, C.CString(value))

	fmt.Printf(">> [Queryable ] Responding ('%s': '%s')\n", keyStringGo, value)

	C.z_query_reply(query, C.z_query_keyexpr(query), C.z_bytes_move(&replyPayload), &options)
}

func main() {
	args := parseArgs()
	value = args.value

	logLevel := C.CString("error")
	defer C.free(unsafe.Pointer(logLevel))
	C.zc_init_log_from_env_or(logLevel)

	var config C.z_owned_config_t
	utils.ConfigFromArgs((*utils.ZConfig)(unsafe.Pointer(&config)), &args.common)

	var session C.z_owned_session_t
	if C.z_open(&session, C.z_config_move(&config), nil) < 0 {
		fmt.Println("Unable to open session!")
		os.Exit(-1)
	}
	defer C.z_session_drop(C.z_session_move(&session))

	var ke C.z_view_keyexpr_t
	if C.z_view_keyexpr_from_str(&ke, C.CString(args.keyexpr)) < 0 {
		fmt.Printf("%s is not a valid key expression", args.keyexpr)
		os.Exit(-1)
	}

	fmt.Printf("Declaring Queryable on '%s'...\n", args.keyexpr)

	var callback C.z_owned_closure_query_t
	C.z_closure_query(&callback, (*[0]byte)(C.queryHandler), nil, nil)

	var qable C.z_owned_queryable_t

	var opts C.z_queryable_options_t
	C.z_queryable_options_default(&opts)
	opts.complete = C.bool(args.complete)

	if C.z_declare_queryable(C.z_session_loan(&session), &qable, C.z_view_keyexpr_loan(&ke), C.z_closure_query_move(&callback), &opts) < 0 {
		fmt.Println("Unable to create queryable.")
		os.Exit(-1)
	}
	defer C.z_queryable_drop(C.z_queryable_move(&qable))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")
	<-stop
}
