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

void dataHandler(struct z_loaned_sample_t *sample, void *context);
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

const defaultKeyexpr = "group1/**"

type Args struct {
	keyexpr string
	history bool
	common  utils.CommonArgs
}

func parseArgs() Args {
	var keyexpr string
	var history bool

	flag.StringVar(&keyexpr, "k", defaultKeyexpr, "The key expression to subscribe to")
	flag.BoolVar(&history, "history", false, "Get historical liveliness tokens")

	return Args{
		keyexpr: keyexpr,
		history: history,
		common:  utils.ParseCommonArgs(),
	}
}

//export dataHandler
func dataHandler(sample *C.z_loaned_sample_t, _ unsafe.Pointer) {
	var keyString C.z_view_string_t
	C.z_keyexpr_as_view_string(C.z_sample_keyexpr(sample), &keyString)

	keyStringGo := C.GoStringN(C.z_string_data(C.z_view_string_loan(&keyString)), (C.int)(C.z_string_len(C.z_view_string_loan(&keyString))))

	switch C.z_sample_kind(sample) {
	case C.Z_SAMPLE_KIND_PUT:
		fmt.Printf(">> [LivelinessSubscriber] New alive token ('%s')\n", keyStringGo)
	case C.Z_SAMPLE_KIND_DELETE:
		fmt.Printf(">> [LivelinessSubscriber] Dropped token ('%s')\n", keyStringGo)
	}
}

func main() {
	args := parseArgs()

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

	fmt.Printf("Declaring liveliness subscriber on '%s'...\n", args.keyexpr)

	keyExprC := C.CString(args.keyexpr)
	defer C.free(unsafe.Pointer(keyExprC))
	var ke C.z_view_keyexpr_t
	if C.z_view_keyexpr_from_str(&ke, keyExprC) < 0 {
		fmt.Printf("%s is not a valid key expression\n", args.keyexpr)
		os.Exit(-1)
	}

	var callback C.z_owned_closure_sample_t
	C.z_closure_sample(&callback, (*[0]byte)(C.dataHandler), nil, nil)

	var opts C.z_liveliness_subscriber_options_t
	C.z_liveliness_subscriber_options_default(&opts)
	opts.history = C.bool(args.history)

	var sub C.z_owned_subscriber_t
	if C.z_liveliness_declare_subscriber(C.z_session_loan(&session), &sub, C.z_view_keyexpr_loan(&ke), C.z_closure_sample_move(&callback), &opts) < 0 {
		fmt.Println("Unable to declare liveliness subscriber.")
		os.Exit(-1)
	}
	defer C.z_subscriber_drop(C.z_subscriber_move(&sub))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")
	<-stop
}
