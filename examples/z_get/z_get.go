package main

/*
#cgo LDFLAGS: -lzenohcd
#include "zenoh.h"
*/
import "C"

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"unsafe"
	"zenoh-go/examples/utils"
)

const (
	defaultSelector = "demo/example/**"
	defaultTimeout  = 10000
	defaultValue    = ""
)

type Args struct {
	selector  string
	value     string
	timeoutMs uint64
	common    utils.CommonArgs
}

func parseArgs() Args {
	var (
		selector  string
		value     string
		timeoutMs uint64
	)

	flag.StringVar(&selector, "s", defaultSelector, "The selector of resources to query")
	flag.StringVar(&value, "p", defaultValue, "An optional value to put in the query")
	flag.Uint64Var(&timeoutMs, "o", defaultTimeout, "Query timeout in milliseconds")

	return Args{
		selector:  selector,
		value:     value,
		timeoutMs: timeoutMs,
		common:    utils.ParseCommonArgs(),
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

	// Open session
	var session C.z_owned_session_t
	if C.z_open(&session, C.z_config_move(&config), nil) < 0 {
		fmt.Println("Unable to open session!")
		os.Exit(-1)
	}
	defer C.z_session_drop(C.z_session_move(&session))

	// Create key expression
	ke := args.selector
	keLen := strings.Index(ke, "?")
	var paramsC *C.char
	pinner.Pin(&paramsC)
	if keLen == -1 {
		keLen = len(ke)
		paramsC = C.CString("")
	} else {
		paramsC = C.CString(ke[keLen+1:])
	}
	defer C.free(unsafe.Pointer(paramsC))
	keyExprC := C.CString(ke[:keLen])
	defer C.free(unsafe.Pointer(keyExprC))

	var keyExpr C.z_view_keyexpr_t
	if C.z_view_keyexpr_from_substr(&keyExpr, keyExprC, C.size_t(keLen)) < 0 {
		fmt.Printf("%s is not a valid key expression\n", ke[:keLen])
		os.Exit(-1)
	}

	fmt.Printf("Sending Query '%s'...\n", args.selector)

	// Create FIFO handler and closure
	var handler C.z_owned_fifo_handler_reply_t
	var closure C.z_owned_closure_reply_t
	C.z_fifo_channel_reply_new(&closure, &handler, 16)

	// Set get options
	var opts C.z_get_options_t
	C.z_get_options_default(&opts)
	opts.timeout_ms = C.uint64_t(args.timeoutMs)

	if args.value != "" {
		payloadC := C.CString(args.value)
		defer C.free(unsafe.Pointer(payloadC))

		var payload C.z_owned_bytes_t
		pinner.Pin(&payload)
		C.z_bytes_from_static_str(&payload, payloadC)
		opts.payload = C.z_bytes_move(&payload)
	}

	// Send query
	C.z_get(C.z_session_loan(&session), C.z_view_keyexpr_loan(&keyExpr), paramsC, C.z_closure_reply_move(&closure), &opts)

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

			var replyStr C.z_owned_string_t
			C.z_bytes_to_string(C.z_sample_payload(sample), &replyStr)

			fmt.Printf(">> Received ('%s': '%s')\n",
				C.GoStringN(C.z_string_data(C.z_view_string_loan(&keyStr)), C.int(C.z_string_len(C.z_view_string_loan(&keyStr)))),
				C.GoStringN(C.z_string_data(C.z_string_loan(&replyStr)), C.int(C.z_string_len(C.z_string_loan(&replyStr)))))

			C.z_string_drop(C.z_string_move(&replyStr))
		} else {
			fmt.Println("Received an error")
		}
		C.z_reply_drop(C.z_reply_move(&reply))
	}

	C.z_fifo_handler_reply_drop(C.z_fifo_handler_reply_move(&handler))
	C.z_session_drop(C.z_session_move(&session))
}
