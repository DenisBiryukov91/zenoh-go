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

package zenoh_ext

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"unsafe"
	"zenoh-go/zenoh"

	"github.com/go-delve/delve/pkg/dwarf/leb128"
)

const maxInt = math.MaxInt
const maxIntLeb128Bytes = uint32((unsafe.Sizeof(int(0))*8 - 1 + (7 - 1)) / 7) // (sizeof(int) * 8 - 1 (because of sign)) / 7 (useful bits per byte in LEB128)
const uint16Size = int(unsafe.Sizeof(uint16(0)))
const uint32Size = int(unsafe.Sizeof(uint32(0)))
const uint64Size = int(unsafe.Sizeof(uint64(0)))

// Interface for adding support for custom types serialization.
type ZSerializeable interface {
	SerializeWithZSerializer(*ZSerializer) error
}

// A Zenoh serializer.
// Provides functionality for tuple-like serialization.
type ZSerializer struct {
	buffer []byte
}

func (serializer *ZSerializer) SerializeInt8(value int8) {
	serializer.buffer = append(serializer.buffer, uint8(value))
}

func (serializer *ZSerializer) SerializeUint8(value uint8) {
	serializer.buffer = append(serializer.buffer, value)
}

func (serializer *ZSerializer) SerializeInt16(value int16) {
	serializer.buffer = binary.LittleEndian.AppendUint16(serializer.buffer, uint16(value))
}

func (serializer *ZSerializer) SerializeUint16(value uint16) {
	serializer.buffer = binary.LittleEndian.AppendUint16(serializer.buffer, value)
}

func (serializer *ZSerializer) SerializeInt32(value int32) {
	serializer.buffer = binary.LittleEndian.AppendUint32(serializer.buffer, uint32(value))
}

func (serializer *ZSerializer) SerializeUint32(value uint32) {
	serializer.buffer = binary.LittleEndian.AppendUint32(serializer.buffer, value)
}

func (serializer *ZSerializer) SerializeInt64(value int64) {
	serializer.buffer = binary.LittleEndian.AppendUint64(serializer.buffer, uint64(value))
}

func (serializer *ZSerializer) SerializeUint64(value uint64) {
	serializer.buffer = binary.LittleEndian.AppendUint64(serializer.buffer, value)
}

func (serializer *ZSerializer) SerializeBool(value bool) {
	if value {
		serializer.SerializeUint8(1)
	} else {
		serializer.SerializeUint8(0)
	}
}

func (serializer *ZSerializer) SerializeFloat32(value float32) {
	serializer.SerializeUint32(math.Float32bits(value))
}

func (serializer *ZSerializer) SerializeFloat64(value float64) {
	serializer.SerializeUint64(math.Float64bits(value))
}

// Serialize length of the sequence. Can be used when defining serialization for custom containers.
func (serializer *ZSerializer) SerializeSequenceLen(value int) {
	buf := bytes.NewBuffer(serializer.buffer)
	leb128.EncodeUnsigned(buf, uint64(value))
	serializer.buffer = buf.Bytes()
}

func (serializer *ZSerializer) SerializeBytes(value []byte) {
	serializer.SerializeSequenceLen(len(value))
	serializer.buffer = append(serializer.buffer, value...)
}

func (serializer *ZSerializer) SerializeString(value string) {
	serializer.SerializeSequenceLen(len(value))
	for i := 0; i < len(value); i++ {
		serializer.buffer = append(serializer.buffer, value[i])
	}
}

func (serializer *ZSerializer) serializeSlice(value reflect.Value) error {
	serializer.SerializeSequenceLen(value.Len())
	for i := 0; i < value.Len(); i++ {
		err := serializer.Serialize(value.Index(i).Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

func (serializer *ZSerializer) serializeMap(value reflect.Value) error {
	serializer.SerializeSequenceLen(value.Len())
	iter := value.MapRange()
	for iter.Next() {
		err := serializer.Serialize(iter.Key().Interface())
		if err != nil {
			return err
		}
		err = serializer.Serialize(iter.Value().Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

// Serialize any supported type and append it to existing serialized payload.
// Supported types are:
//   - built-in primitive types: int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64, bool,
//   - string,
//   - types that implement ZSerializeable interface,
//   - arrays, maps and slices of supported types.
//
// A non-nil error will be returned, when passing an unsupported type.
func (serializer *ZSerializer) Serialize(value interface{}) error {
	var err error
	switch v := value.(type) {
	case int8:
		serializer.SerializeInt8(v)
	case uint8:
		serializer.SerializeUint8(v)
	case int16:
		serializer.SerializeInt16(v)
	case uint16:
		serializer.SerializeUint16(v)
	case int32:
		serializer.SerializeInt32(v)
	case uint32:
		serializer.SerializeUint32(v)
	case int64:
		serializer.SerializeInt64(v)
	case uint64:
		serializer.SerializeUint64(v)
	case float32:
		serializer.SerializeFloat32(v)
	case float64:
		serializer.SerializeFloat64(v)
	case bool:
		serializer.SerializeBool(v)
	case []byte:
		serializer.SerializeBytes(v)
	case string:
		serializer.SerializeString(v)
	case ZSerializeable:
		err = v.SerializeWithZSerializer(serializer)
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			return serializer.serializeSlice(rv)
		} else if rv.Kind() == reflect.Map {
			return serializer.serializeMap(rv)
		}
		return fmt.Errorf("unable to serialize type %T", v)
	}
	return err
}

// Extract serialized data and reset serializer to an empty state.
func (serializer *ZSerializer) Finish() zenoh.ZBytes {
	buf := serializer.buffer
	serializer.buffer = []byte{}
	return zenoh.NewZBytes(buf)
}

// Construct serializer in an empty state
func NewZSerializer() ZSerializer {
	return ZSerializer{}
}

// Serialize any supported type.
// Supported types are:
//   - built-in primitive types: int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64, bool,
//   - string,
//   - types that implement ZSerializeable interface,
//   - arrays, maps and slices of supported types.
//
// A non-nil error will be returned, when passing an unsupported type.
func ZSerialize[T any](value T) (zenoh.ZBytes, error) {
	serializer := ZSerializer{}
	err := serializer.Serialize(value)
	return serializer.Finish(), err
}

// Interface for adding support for deserialization into custom types.
// DeserializeWithZDeserializer must be defined for pointer receiver.
type ZDeserializeable interface {
	DeserializeWithZDeserializer(*ZDeserializer) error
}

// A Zenoh deserializer.
// Provides functionality for deserializing data, previously serialized using [ZSerializer].
type ZDeserializer struct {
	buf bytes.Buffer
}

// Construct a deserializer for specified payload.
func NewZDeserializer(zbytes zenoh.ZBytes) ZDeserializer {
	return ZDeserializer{buf: *bytes.NewBuffer(zbytes.Bytes())}
}

func (deserializer *ZDeserializer) DeserializeInt8() (int8, error) {
	res, err := deserializer.buf.ReadByte()
	return int8(res), err
}

func (deserializer *ZDeserializer) DeserializeUint8() (uint8, error) {
	return deserializer.buf.ReadByte()
}

func (deserializer *ZDeserializer) DeserializeBool() (bool, error) {
	res, err := deserializer.buf.ReadByte()
	if err == nil && res > 1 {
		return false, fmt.Errorf("unsupported bool value %v", res)
	}
	return res == 1, err
}

func (deserializer *ZDeserializer) DeserializeInt16() (int16, error) {
	res, err := deserializer.DeserializeUint16()
	return int16(res), err
}

func (deserializer *ZDeserializer) DeserializeUint16() (uint16, error) {
	b := deserializer.buf.Next(uint16Size)
	if len(b) != uint16Size {
		return 0, io.EOF
	}
	return binary.LittleEndian.Uint16(b), nil
}

func (deserializer *ZDeserializer) DeserializeInt32() (int32, error) {
	res, err := deserializer.DeserializeUint32()
	return int32(res), err
}

func (deserializer *ZDeserializer) DeserializeUint32() (uint32, error) {
	b := deserializer.buf.Next(uint32Size)
	if len(b) != uint32Size {
		return 0, io.EOF
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (deserializer *ZDeserializer) DeserializeInt64() (int64, error) {
	res, err := deserializer.DeserializeUint64()
	return int64(res), err
}

func (deserializer *ZDeserializer) DeserializeUint64() (uint64, error) {
	b := deserializer.buf.Next(uint64Size)
	if len(b) != uint64Size {
		return 0, io.EOF
	}
	return binary.LittleEndian.Uint64(b), nil
}

func (deserializer *ZDeserializer) DeserializeFloat32() (float32, error) {
	res, err := deserializer.DeserializeUint32()
	return math.Float32frombits(res), err
}

func (deserializer *ZDeserializer) DeserializeFloat64() (float64, error) {
	res, err := deserializer.DeserializeUint64()
	return math.Float64frombits(res), err
}

// Read length of the sequence previously written by [ZSerializer.SerializeSequenceLen]
func (deserializer *ZDeserializer) DeserializeSequenceLen() (len int, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = io.EOF
		}
	}()
	res, n := leb128.DecodeUnsigned(&deserializer.buf)
	if n > maxIntLeb128Bytes || res > maxInt {
		err = fmt.Errorf("leb128 decoding overflow")
	}
	return int(res), err
}

func (deserializer *ZDeserializer) DeserializeBytes() ([]byte, error) {
	l, err := deserializer.DeserializeSequenceLen()
	if err != nil {
		return []byte{}, err
	}
	res := deserializer.buf.Next(l)
	if len(res) != l {
		return []byte{}, io.EOF
	}
	return res, nil
}

func (deserializer *ZDeserializer) DeserializeString() (string, error) {
	res, err := deserializer.DeserializeBytes()
	// is utf8 check necessary here, given that go allows strings to contain non-valid utf8 byte sequences ?
	return string(res), err
}

func (deserializer *ZDeserializer) deserializeSlice(value reflect.Value) error {
	l, err := deserializer.DeserializeSequenceLen()
	if err != nil {
		return err
	}
	value.SetZero()
	value.Grow(l)
	value.SetLen(l)

	for i := 0; i < l; i++ {
		err := deserializer.Deserialize(value.Index(i).Addr().Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

func (deserializer *ZDeserializer) deserializeArray(value reflect.Value) error {
	l, err := deserializer.DeserializeSequenceLen()
	if err != nil {
		return err
	}
	if value.Len() != l {
		return fmt.Errorf("array size error - expected: %d, decoded: %d", value.Len(), l)
	}

	for i := 0; i < l; i++ {
		err := deserializer.Deserialize(value.Index(i).Addr().Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

func (deserializer *ZDeserializer) deserializeMap(value reflect.Value) error {
	l, err := deserializer.DeserializeSequenceLen()
	if err != nil {
		return err
	}
	keyType := value.Type().Key()
	valueType := value.Type().Elem()
	m := reflect.MakeMapWithSize(reflect.MapOf(keyType, valueType), l)
	for i := 0; i < l; i++ {
		key := reflect.New(keyType)
		value := reflect.New(valueType)
		err := deserializer.Deserialize(key.Interface())
		if err != nil {
			return err
		}
		err = deserializer.Deserialize(value.Interface())
		if err != nil {
			return err
		}
		m.SetMapIndex(key.Elem(), value.Elem())
	}
	value.Set(m)
	return nil
}

// Deserialize next portion of data into any supported type and advance the reading position.
// Supported types are:
//   - built-in primitive types: int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64, bool,
//   - string,
//   - types that implement ZDeserializeable interface,
//   - arrays, maps and slices of supported types.
//
// A non-nil error will be returned, when passing an unsupported type, or when deserialization fails.
// out must be a non-nil pointer to the target type instance.
func (deserializer *ZDeserializer) Deserialize(out any) error {
	switch v := out.(type) {
	case *int8:
		res, err := deserializer.DeserializeInt8()
		*v = res
		return err
	case *uint8:
		res, err := deserializer.DeserializeUint8()
		*v = res
		return err
	case *int16:
		res, err := deserializer.DeserializeInt16()
		*v = res
		return err
	case *uint16:
		res, err := deserializer.DeserializeUint16()
		*v = res
		return err
	case *int32:
		res, err := deserializer.DeserializeInt32()
		*v = res
		return err
	case *uint32:
		res, err := deserializer.DeserializeUint32()
		*v = res
		return err
	case *int64:
		res, err := deserializer.DeserializeInt64()
		*v = res
		return err
	case *uint64:
		res, err := deserializer.DeserializeUint64()
		*v = res
		return err
	case *float32:
		res, err := deserializer.DeserializeFloat32()
		*v = res
		return err
	case *float64:
		res, err := deserializer.DeserializeFloat64()
		*v = res
		return err
	case *bool:
		res, err := deserializer.DeserializeBool()
		*v = res
		return err
	case *[]byte:
		res, err := deserializer.DeserializeBytes()
		*v = res
		return err
	case *string:
		res, err := deserializer.DeserializeString()
		*v = res
		return err
	default:
		rv := reflect.ValueOf(out)
		if rv.Kind() != reflect.Pointer || rv.IsNil() {
			return fmt.Errorf("'out' argument should be a non-nil pointer")
		}
		elt := rv.Elem()
		if elt.Kind() == reflect.Slice {
			return deserializer.deserializeSlice(elt)
		} else if elt.Kind() == reflect.Array {
			return deserializer.deserializeArray(elt)
		} else if elt.Kind() == reflect.Map {
			return deserializer.deserializeMap(elt)
		} else if rv.Kind() != reflect.Interface { // element should be a typed value, not an abstract interface
			if d, ok := rv.Interface().(ZDeserializeable); ok {
				return d.DeserializeWithZDeserializer(deserializer)
			}
		}
		return fmt.Errorf("unable to deserialize into %T", elt.Type())
	}
}

// Return `true` if all bytes were used for deserialization, `false` otherwise.
func (deserializer *ZDeserializer) IsDone() bool {
	return deserializer.buf.Len() == 0
}

// Deserialize into any supported type.
// Supported types are:
//   - built-in primitive types: int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64, bool,
//   - string,
//   - types that implement ZDeserializeable interface,
//   - arrays, maps and slices of supported types.
//
// A non-nil error will be returned, when passing an unsupported type, when deserialization fails, or if
// zbytes contains more bytes than required for deserialization.
func ZDeserialize[T any](zbytes zenoh.ZBytes) (T, error) {
	var t T
	deserializer := NewZDeserializer(zbytes)
	err := deserializer.Deserialize(&t)
	if err == nil && !deserializer.IsDone() {
		err = fmt.Errorf("payload contains more bytes than required for deserialization")
	}
	return t, err
}
