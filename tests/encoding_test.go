//
// Copyright (c) 2026 ZettaScale Technology
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

	"zenoh-go/zenoh"

	"github.com/stretchr/testify/assert"
)

func TestEncodingDefault(t *testing.T) {
	e := zenoh.NewEncodingDefault()
	assert.EqualValues(t, "zenoh/bytes", e.String())
}

func TestEncodingFromString(t *testing.T) {
	e := zenoh.NewEncodinFromString("text/plain")
	assert.EqualValues(t, "text/plain", e.String())
}

func TestEncodingFromStringWithSchema(t *testing.T) {
	e := zenoh.NewEncodinFromString("text/plain;utf-8")
	s := e.String()
	assert.EqualValues(t, "text/plain;utf-8", s)
}

func TestEncodingSetSchema(t *testing.T) {
	e := zenoh.PredefinedEncodings.TextPlain()
	e.SetSchema("utf-8")
	s := e.String()
	assert.EqualValues(t, "text/plain;utf-8", s)
}

func TestEncodingSetSchemaEmpty(t *testing.T) {
	e := zenoh.PredefinedEncodings.TextPlain()
	before := e.String()
	e.SetSchema("")
	// Setting empty schema is a no-op; string representation must not change.
	assert.Equal(t, before, e.String())
}

func TestEncodingFromStringCustom(t *testing.T) {
	e := zenoh.NewEncodinFromString("my/encoding")
	assert.EqualValues(t, "my/encoding", e.String())
	e.SetSchema("my-schema")
	assert.EqualValues(t, "my/encoding;my-schema", e.String())
}
