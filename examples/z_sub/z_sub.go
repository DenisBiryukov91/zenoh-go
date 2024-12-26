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

const defaultKeyexpr = "demo/example/**"

type Args struct {
	keyexpr string
	common  utils.CommonArgs
}

func parseArgs() Args {
	args := Args{common: utils.ParseCommonArgs()}

	flag.StringVar(&args.keyexpr, "k", defaultKeyexpr, "The key expression to publish to")

	return args
}

func kindToStr(kind uint32) string {
	switch kind {
	case C.Z_SAMPLE_KIND_PUT:
		return "PUT"
	case C.Z_SAMPLE_KIND_DELETE:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

//export dataHandler
func dataHandler(sample *C.z_loaned_sample_t, _ unsafe.Pointer) {
	var keyString C.z_view_string_t
	C.z_keyexpr_as_view_string(C.z_sample_keyexpr(sample), &keyString)

	var payloadString C.z_owned_string_t
	C.z_bytes_to_string(C.z_sample_payload(sample), &payloadString)
	defer C.z_string_drop(C.z_string_move(&payloadString))

	fmt.Printf(">> [Subscriber] Received %s ('%s': '%s')\n",
		kindToStr(C.z_sample_kind(sample)),
		C.GoStringN(C.z_string_data(C.z_view_string_loan(&keyString)), (C.int)(C.z_string_len(C.z_view_string_loan(&keyString)))),
		C.GoStringN(C.z_string_data(C.z_string_loan(&payloadString)), (C.int)(C.z_string_len(C.z_string_loan(&payloadString)))))
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

	var callback C.z_owned_closure_sample_t
	C.z_closure_sample(&callback, (*[0]byte)(C.dataHandler), nil, nil)

	fmt.Printf("Declaring Subscriber on '%s'...\n", args.keyexpr)

	keyExprC := C.CString(args.keyexpr)
	defer C.free(unsafe.Pointer(keyExprC))
	var ke C.z_view_keyexpr_t
	C.z_view_keyexpr_from_str(&ke, keyExprC)

	var sub C.z_owned_subscriber_t
	if C.z_declare_subscriber(C.z_session_loan(&session), &sub, C.z_view_keyexpr_loan(&ke), C.z_closure_sample_move(&callback), nil) != 0 {
		fmt.Println("Unable to declare subscriber.")
		os.Exit(-1)
	}
	defer C.z_subscriber_drop(C.z_subscriber_move(&sub))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")
	<-stop
}
