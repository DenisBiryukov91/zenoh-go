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

import (
	"fmt"
	"os"
	"strings"
	"zenoh-go/zenoh"

	"github.com/spf13/pflag"
)

type ConfigArgs struct {
	Path                string
	Mode                string
	Connect             []string
	Listen              []string
	Cfg                 []string
	NoMulticastScouting bool
}

func toListString(a []string) string {
	var buf strings.Builder
	buf.WriteString("[")
	for idx, value := range a {
		buf.WriteString("'")
		buf.WriteString(value)
		buf.WriteString("'")
		if idx+1 < len(a) {
			buf.WriteString(",")
		}
	}
	buf.WriteString("]")
	return buf.String()
}

func (args *ConfigArgs) toConfig() (zenoh.Config, error) {
	var config zenoh.Config
	var err error
	if args.Path != "" {
		config, err = zenoh.NewConfigFromFile(args.Path)
		if err != nil {
			return config, err
		}
	} else {
		config = zenoh.NewConfigDefault()
	}
	if args.NoMulticastScouting {
		config.InsertJson5(zenoh.ConfigMulticastScoutingKey, "false")
	}
	err = config.InsertJson5(zenoh.ConfigModeKey, fmt.Sprintf("'%s'", args.Mode))
	if err != nil {
		return config, err
	}
	if len(args.Connect) > 0 {
		err = config.InsertJson5(zenoh.ConfigConnectKey, toListString(args.Connect))
		if err != nil {
			return config, err
		}
	}
	if len(args.Listen) > 0 {
		err = config.InsertJson5(zenoh.ConfigListenKey, toListString(args.Listen))
		if err != nil {
			return config, err
		}
	}
	for _, value := range args.Cfg {
		key_value := strings.SplitN(value, ":", 2)
		if len(key_value) != 2 {
			return config, fmt.Errorf("--cfg` argument: expected KEY:VALUE pair, got %s ", value)
		}
		err = config.InsertJson5(key_value[0], key_value[1])
		if err != nil {
			return config, err
		}
	}

	return config, nil
}

func ParseConfig() zenoh.Config {
	args := ConfigArgs{}
	pflag.StringVarP(&args.Path, "config", "c", "", "The path to a configuration file for the session. If this option isn't passed, the default configuration will be used.")
	pflag.StringVarP(&args.Mode, "mode", "m", "peer", "Zenoh session mode [client, peer, router].")
	pflag.BoolVar(&args.NoMulticastScouting, "no-multicast-scouting", false, "By default zenohd replies to multicast scouting messages for being discovered by peers and clients. This option disables this feature.")
	pflag.StringArrayVarP(&args.Connect, "connect", "e", make([]string, 0), "Endpoint to connect to. Repeat option to pass multiple endpoints. If none are given, endpoints will be discovered through multicast-scouting if it is enabled (e.g.: '-e tcp/192.168.1.1:7447').")
	pflag.StringArrayVarP(&args.Listen, "listen", "l", make([]string, 0), "Locator to listen on. Repeat option to pass multiple locators. If none are given, the default configuration will be used (e.g.: '-l tcp/192.168.1.1:7447').")
	pflag.StringArrayVar(&args.Cfg, "cfg", make([]string, 0), "Allows arbitrary configuration changes as column-separated KEY:VALUE pairs. Where KEY must be a valid config path and VALUE must be a valid JSON5 string that can be deserialized to the expected type for the KEY field. Example: --cfg='transport/unicast/max_links:2'.")

	pflag.Parse()
	config, err := args.toConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	return config
}

func ParseSelector(selector string) (string, string) {
	res := strings.SplitN(selector, "?", 1)
	if len(res) == 2 {
		return res[0], res[1]
	} else {
		return res[0], ""
	}
}

func ParseQueryTarget(queryTarget string) (zenoh.QueryTarget, error) {
	switch {
	case queryTarget == "BEST_MATCHING":
		return zenoh.QueryTargetBestMatching, nil
	case queryTarget == "ALL":
		return zenoh.QueryTargetAll, nil
	case queryTarget == "ALL_COMPLETE":
		return zenoh.QueryTargetAllComplete, nil
	default:
		return zenoh.QueryTarget(0), fmt.Errorf("unsupported query target value: '%s'", queryTarget)
	}
}

func ParsePriority(value uint8) (zenoh.Priority, error) {
	if value < uint8(zenoh.PriorityRealTime) || value > uint8(zenoh.PriorityBackground) {
		return zenoh.PriorityDefault, fmt.Errorf("unsupported priority value: '%v'", value)
	}
	return zenoh.Priority(value), nil
}
