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

#include "zenoh_ext_cgo.h"

static z_moved_encoding_t *_create_moved_encoding_from_data(
    const zc_internal_encoding_data_t *encoding_data, z_owned_encoding_t *dst) {
  zc_internal_encoding_from_data(dst, *encoding_data);
  return z_move(*dst);
}

static z_moved_bytes_t *
_create_moved_bytes_from_data(const zc_cgo_bytes_data_t *bytes_data,
                              z_owned_bytes_t *dst) {
  z_bytes_copy_from_buf(dst, bytes_data->data, bytes_data->len);
  return z_move(*dst);
}

z_result_t
zc_cgo_advanced_publisher_put(ze_owned_advanced_publisher_t *publisher,
                              zc_cgo_bytes_data_t *payload_data,
                              ze_advanced_publisher_put_options_t *opts,
                              zc_internal_encoding_data_t *encoding_data,
                              zc_cgo_bytes_data_t *attachment_data) {
  z_owned_bytes_t payload, attachment;
  z_owned_encoding_t encoding;
  z_bytes_copy_from_buf(&payload, payload_data->data, payload_data->len);
  if (encoding_data != NULL) {
    opts->put_options.encoding =
        _create_moved_encoding_from_data(encoding_data, &encoding);
  }
  if (attachment_data != NULL) {
    opts->put_options.attachment =
        _create_moved_bytes_from_data(attachment_data, &attachment);
  }
  return ze_advanced_publisher_put(ze_advanced_publisher_loan_mut(publisher),
                                   z_move(payload), opts);
}

z_result_t
zc_cgo_advanced_publisher_delete(ze_owned_advanced_publisher_t *publisher,
                                 ze_advanced_publisher_delete_options_t *opts) {
  return ze_advanced_publisher_delete(ze_advanced_publisher_loan_mut(publisher),
                                      opts);
}

void zenohMissListenerCCallback(const ze_miss_t *miss, void *context) {
  zenohMissListenerCallback((zc_cgo_const_miss_t *)miss, context);
}
