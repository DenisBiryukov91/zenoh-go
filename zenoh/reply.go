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

// #include "zenoh.h"
import "C"
import "github.com/BooleanCat/option"

// A Zenoh reply error - a combination of reply error payload and its encoding.
type ReplyError struct {
	payload  ZBytes
	encoding Encoding
}

// Return the error payload data.
func (reply_error *ReplyError) Payload() ZBytes {
	return reply_error.payload
}

// Return the encoding associated with the error data.
func (reply_error *ReplyError) Encoding() Encoding {
	return reply_error.encoding
}

func newReplyErrorFromC(c_reply_error *C.z_loaned_reply_err_t) ReplyError {
	var e ReplyError
	e.payload = newZBytesFromC(C.z_reply_err_payload(c_reply_error))
	e.encoding = newEncodingFromC(C.z_reply_err_encoding(c_reply_error))
	return e
}

type replyErr struct {
	value ReplyError
}

func (reply_error *replyErr) Ok() option.Option[Sample] {
	return option.None[Sample]()
}
func (reply_error *replyErr) Err() option.Option[ReplyError] {
	return option.Some(reply_error.value)
}
func (reply_error *replyErr) IsOk() bool {
	return false
}

type replyOk struct {
	value Sample
}

func (sample *replyOk) Ok() option.Option[Sample] {
	return option.Some(sample.value)
}
func (sample *replyOk) Err() option.Option[ReplyError] {
	return option.None[ReplyError]()
}
func (sample *replyOk) IsOk() bool {
	return true
}

// A Zenoh reply from a queryable.
type Reply interface {
	Ok() option.Option[Sample]      // Yield the contents of the reply by asserting it indicates a success.
	Err() option.Option[ReplyError] // Yield the contents of the reply by asserting it indicates a error.
	IsOk() bool                     // Return ``true`` if reply contains a valid response, ``false`` otherwise (in this case it contains a error value).
}

func newReplyFromC(c_reply *C.z_loaned_reply_t) Reply {
	cSample := C.z_reply_ok(c_reply)
	if cSample != nil {
		return &replyOk{value: newSampleFromC(cSample)}
	} else {
		return &replyErr{value: newReplyErrorFromC(C.z_reply_err(c_reply))}
	}
}
