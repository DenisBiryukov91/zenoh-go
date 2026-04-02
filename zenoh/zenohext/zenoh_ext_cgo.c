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
  z_bytes_copy_from_buf(dst, (const uint8_t *)bytes_data->data,
                        bytes_data->len);
  return z_move(*dst);
}

z_result_t
zc_cgo_advanced_publisher_put(ze_owned_advanced_publisher_t *publisher,
                              zc_cgo_bytes_data_t *payload_data,
                              zc_cgo_advanced_publisher_put_options_t *opts) {
  z_owned_bytes_t payload, attachment;
  z_owned_encoding_t encoding;
  if (opts == NULL) {
    return ze_advanced_publisher_put(
        ze_advanced_publisher_loan_mut(publisher),
        _create_moved_bytes_from_data(payload_data, &payload), NULL);
  }
  ze_advanced_publisher_put_options_t options;
  ze_advanced_publisher_put_options_default(&options);
  if (opts->has_encoding) {
    options.put_options.encoding =
        _create_moved_encoding_from_data(&opts->encoding_data, &encoding);
  }
  if (opts->has_attachment) {
    options.put_options.attachment =
        _create_moved_bytes_from_data(&opts->attachment_data, &attachment);
  }
  options.put_options.timestamp = opts->has_timestamp ? &opts->timestamp : NULL;
  options.put_options.source_info =
      opts->has_source_info ? &opts->source_info : NULL;
  return ze_advanced_publisher_put(
      ze_advanced_publisher_loan_mut(publisher),
      _create_moved_bytes_from_data(payload_data, &payload), &options);
}

z_result_t zc_cgo_advanced_publisher_delete(
    ze_owned_advanced_publisher_t *publisher,
    zc_cgo_advanced_publisher_delete_options_t *opts) {
  if (opts == NULL) {
    return ze_advanced_publisher_delete(
        ze_advanced_publisher_loan_mut(publisher), NULL);
  }
  ze_advanced_publisher_delete_options_t options;
  ze_advanced_publisher_delete_options_default(&options);
  options.delete_options.timestamp =
      opts->has_timestamp ? &opts->timestamp : NULL;
  return ze_advanced_publisher_delete(ze_advanced_publisher_loan_mut(publisher),
                                      &options);
}

void zenohMissListenerCCallback(const ze_miss_t *miss, void *context) {
  zenohMissListenerCallback((zc_cgo_const_miss_t *)miss, context);
}
