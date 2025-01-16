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

package zenoh

import "fmt"

type ZError struct {
	code int8
	msg  string
}

func (e ZError) Error() string { return fmt.Sprintf("%s (Error Code: %d)", e.msg, e.code) }

func NewZError(code int8, msg string) ZError {
	return ZError{code: code, msg: msg}
}
