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

package utils

/*
#cgo LDFLAGS: -lzenohc
#include "zenoh.h"
*/
import "C"

import (
	"flag"
	"fmt"
	"strings"
	"unsafe"
)

type ZConfig C.z_owned_config_t

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type CommonArgs struct {
	Config              string
	Mode                string
	Connect             arrayFlags
	Listen              arrayFlags
	NoMulticastScouting bool
}

func ParseCommonArgs() CommonArgs {
	args := CommonArgs{}
	flag.StringVar(&args.Config, "config", "", "Path to a configuration file.")
	flag.StringVar(&args.Mode, "mode", "peer", "Zenoh session mode [client, peer, router].")
	flag.BoolVar(&args.NoMulticastScouting, "no-multicast-scouting", false, "Disable multicast-based scouting.")
	flag.Var(&args.Connect, "connect", "Endpoints to connect to")
	flag.Var(&args.Listen, "listen", "Endpoints to listen on")

	flag.Parse()
	return args
}

func applyListConfig(arrayFlags []string, configKey *C.char, config *C.z_owned_config_t) error {
	var buf strings.Builder
	for _, value := range arrayFlags {
		buf.WriteString("'" + value + "',")
	}
	list := buf.String()
	if len(list) > 0 {
		list = "[" + list[:len(list)-1] + "]"
		cList := C.CString(list)
		defer C.free(unsafe.Pointer(cList))
		if C.zc_config_insert_json5(C.z_config_loan_mut(config), configKey, cList) < 0 {
			return fmt.Errorf("Couldn't insert list %s", list)
		}
	}

	return nil
}

func ConfigFromArgs(cfg *ZConfig, args *CommonArgs) error {
	config := (*C.z_owned_config_t)(cfg)

	if args.Config != "" {
		cConfig := C.CString(args.Config)
		defer C.free(unsafe.Pointer(cConfig))
		if C.zc_config_from_file(config, cConfig) < 0 {
			return fmt.Errorf("failed to read config file '%s'", args.Config)
		}
	} else {
		C.z_config_default(config)
	}

	if args.Mode != "" {
		cMode := C.CString("'" + args.Mode + "'")
		defer C.free(unsafe.Pointer(cMode))
		if C.zc_config_insert_json5(C.z_config_loan_mut(config), C.Z_CONFIG_MODE_KEY, cMode) < 0 {
			return fmt.Errorf("Couldn't insert mode %s", args.Mode)
		}
	}

	if args.NoMulticastScouting {
		cFalse := C.CString("false")
		defer C.free(unsafe.Pointer(cFalse))
		if C.zc_config_insert_json5(C.z_config_loan_mut(config), C.Z_CONFIG_MULTICAST_SCOUTING_KEY, cFalse) < 0 {
			return fmt.Errorf("Couldn't disable multicast-scouting.")
		}
	}

	if len(args.Connect) > 0 {
		if err := applyListConfig(args.Connect, C.Z_CONFIG_CONNECT_KEY, config); err != nil {
			return err
		}
	}

	if len(args.Listen) > 0 {
		if err := applyListConfig(args.Listen, C.Z_CONFIG_LISTEN_KEY, config); err != nil {
			return err
		}
	}
	return nil
}
