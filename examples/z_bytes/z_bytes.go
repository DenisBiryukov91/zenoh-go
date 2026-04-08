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

	"github.com/eclipse-zenoh/zenoh-go/zenoh"
	"github.com/eclipse-zenoh/zenoh-go/zenoh/zenohext"
)

func main() {
	// Using raw data
	// string
	{
		input := "test"
		payload := zenoh.NewZBytesFromString(input)
		output := payload.String()
		fmt.Printf("Input: %v, Output: %v\n", input, output)
	}
	// []byte
	{
		input := []byte{1, 2, 3, 4}
		payload := zenoh.NewZBytes(input)
		output := payload.Bytes()
		fmt.Printf("Input: %v, Output: %v\n", input, output)
	}

	// serialization
	// slice
	{
		input := []int32{1, 2, 3, 4}
		payload, _ := zenohext.ZSerialize(input)
		output, _ := zenohext.ZDeserialize[[]int32](payload)
		fmt.Printf("Input: %v, Output: %v\n", input, output)
	}

	// map
	{
		input := map[uint64]string{0: "abc", 1: "def"}
		payload, _ := zenohext.ZSerialize(input)
		output, _ := zenohext.ZDeserialize[map[uint64]string](payload)
		fmt.Printf("Input: %v, Output: %v\n", input, output)
	}

	// serializer, deserializer for struct or tuple deserialization
	{
		// serialization
		i1 := uint32(1234)
		i2 := "test"
		i3 := []int8{1, 2, 3, 4}

		serializer := zenohext.NewZSerializer()
		serializer.Serialize(i1)
		serializer.Serialize(i2)
		serializer.Serialize(i3)
		payload := serializer.Finish()

		// deserialization
		deserializer := zenohext.NewZDeserializer(payload)
		var o1 uint32
		var o2 string
		var o3 []int8

		deserializer.Deserialize(&o1)
		deserializer.Deserialize(&o2)
		deserializer.Deserialize(&o3)

		fmt.Printf("Input: %v; %v; %v; Output: %v; %v; %v \n", i1, i2, i3, o1, o2, o3)
	}
}
