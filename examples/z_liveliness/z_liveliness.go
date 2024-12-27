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
	"os/signal"
	"runtime"
	"unsafe"
	"zenoh-go/examples/utils"
)

const (
	defaultKeyExpr = "group1/zenoh-rs"
)

type Args struct {
	keyexpr string
	common  utils.CommonArgs
}

func parseArgs() Args {
	var keyexpr string

	flag.StringVar(&keyexpr, "k", defaultKeyExpr, "The key expression for the liveliness token")

	return Args{
		keyexpr: keyexpr,
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

	// Create key expression
	keyExprC := C.CString(args.keyexpr)
	defer C.free(unsafe.Pointer(keyExprC))

	var keyExpr C.z_view_keyexpr_t
	if C.z_view_keyexpr_from_str(&keyExpr, keyExprC) < 0 {
		fmt.Printf("%s is not a valid key expression\n", args.keyexpr)
		os.Exit(-1)
	}

	// Open session
	fmt.Printf("Opening session...\n")
	var session C.z_owned_session_t
	if C.z_open(&session, C.z_config_move(&config), nil) < 0 {
		fmt.Println("Unable to open session!")
		os.Exit(-1)
	}
	defer C.z_session_drop(C.z_session_move(&session))

	// Declare liveliness token
	fmt.Printf("Declaring liveliness token '%s'...\n", args.keyexpr)
	var token C.z_owned_liveliness_token_t
	if C.z_liveliness_declare_token(C.z_session_loan(&session), &token, C.z_view_keyexpr_loan(&keyExpr), nil) < 0 {
		fmt.Println("Unable to create liveliness token!")
		os.Exit(-1)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press CTRL-C to quit...")
	<-stop

	fmt.Println("Undeclaring liveliness token...")
	C.z_liveliness_token_drop(C.z_liveliness_token_move(&token))
}
