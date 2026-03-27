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

package zenoh_test

import (
	"testing"
	"zenoh-go/zenoh/zenohext"

	"github.com/stretchr/testify/assert"
)

func checkSerDe[T any](t *testing.T, value T) {
	zbytes, err := zenohext.ZSerialize(value)
	assert.Equal(t, err, nil)
	res, err := zenohext.ZDeserialize[T](zbytes)
	assert.Equal(t, err, nil)
	assert.Equal(t, value, res)
}

func TestPrimitive(t *testing.T) {
	checkSerDe(t, uint8(5))
	checkSerDe(t, uint16(500))
	checkSerDe(t, uint32(50000))
	checkSerDe(t, uint64(500000000000))

	checkSerDe(t, int8(-5))
	checkSerDe(t, int16(500))
	checkSerDe(t, int32(50000))
	checkSerDe(t, int64(500000000000))

	checkSerDe(t, float32(0.5))
	checkSerDe(t, float64(123.45))
	checkSerDe(t, true)
	checkSerDe(t, false)
}

func TestContainer(t *testing.T) {
	checkSerDe(t, "abcdefg")
	checkSerDe(t, []float32{0.1, 0.2, -0.5, 1000.578})
	checkSerDe(t, []int32{1, 2, 3, -5, 10000, -999999999})
	checkSerDe(t, []int16{1, 2, 3, -5, 5, -500})
	checkSerDe(t, map[uint64]string{100: "abc", 10000: "def", 2000000000: "hij"})
}

type CustomStruct struct {
	vd []float64
	i  int32
	s  string
}

func (cs CustomStruct) SerializeWithZSerializer(serializer *zenohext.ZSerializer) error {
	if err := serializer.Serialize(cs.vd); err != nil {
		return err
	}
	if err := serializer.Serialize(cs.i); err != nil {
		return err
	}
	return serializer.Serialize(cs.s)
}

func (cs *CustomStruct) DeserializeWithZDeserializer(deserializer *zenohext.ZDeserializer) error {
	if err := deserializer.Deserialize(&cs.vd); err != nil {
		return err
	}
	if err := deserializer.Deserialize(&cs.i); err != nil {
		return err
	}
	return deserializer.Deserialize(&cs.s)
}

func TestCustom(t *testing.T) {
	s := CustomStruct{vd: []float64{0.1, 0.2, -1000.55}, i: 32, s: "test"}
	checkSerDe(t, s)
}

func TestSerializerDeserializer(t *testing.T) {
	serializer := zenohext.NewZSerializer()
	err := serializer.Serialize(float64(0.5))
	assert.Equal(t, err, nil)
	err = serializer.Serialize("test")
	assert.Equal(t, err, nil)
	err = serializer.Serialize(uint8(1))
	assert.Equal(t, err, nil)

	data := serializer.Finish()
	assert.Equal(t, len(serializer.Finish().Bytes()), 0)

	var d float64
	var s string
	var u uint8

	deserializer := zenohext.NewZDeserializer(data)
	assert.False(t, deserializer.IsDone())

	err = deserializer.Deserialize(&d)
	assert.Equal(t, err, nil)
	assert.Equal(t, d, 0.5)
	assert.False(t, deserializer.IsDone())

	err = deserializer.Deserialize(&s)
	assert.Equal(t, err, nil)
	assert.Equal(t, s, "test")
	assert.False(t, deserializer.IsDone())

	err = deserializer.Deserialize(&u)
	assert.Equal(t, err, nil)
	assert.Equal(t, u, uint8(1))
	assert.True(t, deserializer.IsDone())
}

type TestTuple struct {
	_0 uint16
	_1 float32
	_2 string
}

func (t TestTuple) SerializeWithZSerializer(serializer *zenohext.ZSerializer) error {
	if err := serializer.Serialize(t._0); err != nil {
		return err
	}
	if err := serializer.Serialize(t._1); err != nil {
		return err
	}
	return serializer.Serialize(t._2)
}

type TestPair struct {
	_0 string
	_1 int16
}

func (t TestPair) SerializeWithZSerializer(serializer *zenohext.ZSerializer) error {
	if err := serializer.Serialize(t._0); err != nil {
		return err
	}
	return serializer.Serialize(t._1)
}

func TestBinaryFormat(t *testing.T) {
	data, _ := zenohext.ZSerialize(int32(1234566))
	assert.Equal(t, data.Bytes(), []byte{134, 214, 18, 0})

	data, _ = zenohext.ZSerialize(int32(-49245))
	assert.Equal(t, data.Bytes(), []byte{163, 63, 255, 255})

	data, _ = zenohext.ZSerialize("test")
	assert.Equal(t, data.Bytes(), []byte{4, 116, 101, 115, 116})

	data, _ = zenohext.ZSerialize(TestTuple{_0: 500, _1: 1234.0, _2: "test"})
	assert.Equal(t, data.Bytes(), []byte{244, 1, 0, 64, 154, 68, 4, 116, 101, 115, 116})

	data, _ = zenohext.ZSerialize([]TestPair{{_0: "s1", _1: 10}, {_0: "s2", _1: -10000}})
	assert.Equal(t, data.Bytes(), []byte{2, 2, 115, 49, 10, 0, 2, 115, 50, 240, 216})

	data, _ = zenohext.ZSerialize([]int64{-100, 500, 100000, -20000000})
	assert.Equal(t, data.Bytes(), []byte{4, 156, 255, 255, 255, 255, 255, 255, 255, 244, 1, 0, 0, 0, 0, 0, 0,
		160, 134, 1, 0, 0, 0, 0, 0, 0, 211, 206, 254, 255, 255, 255, 255})
}
