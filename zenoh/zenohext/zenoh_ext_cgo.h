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

#ifndef ZENOH_EXT_CGO_H
#define ZENOH_EXT_CGO_H
#include "zenoh.h"
#include "zenoh_cgo.h"

extern void zenohMissListenerDrop(void *context);
typedef const struct ze_miss_t zc_cgo_const_miss_t;
extern void zenohMissListenerCallback(zc_cgo_const_miss_t *miss, void *context);
void zenohMissListenerCCallback(const ze_miss_t *miss, void *context);

typedef struct zc_cgo_advanced_publisher_put_options_t {
  zc_internal_encoding_data_t encoding_data;
  zc_cgo_bytes_data_t attachment_data;
  z_timestamp_t timestamp;
  z_source_info_t source_info;
  bool has_encoding;
  bool has_attachment;
  bool has_timestamp;
  bool has_source_info;
} zc_cgo_advanced_publisher_put_options_t;

typedef struct zc_cgo_advanced_publisher_delete_options_t {
  z_timestamp_t timestamp;
  bool has_timestamp;
} zc_cgo_advanced_publisher_delete_options_t;

z_result_t
zc_cgo_advanced_publisher_put(ze_owned_advanced_publisher_t *publisher,
                              zc_cgo_bytes_data_t *payload_data,
                              zc_cgo_advanced_publisher_put_options_t *opts);
z_result_t zc_cgo_advanced_publisher_delete(
    ze_owned_advanced_publisher_t *publisher,
    zc_cgo_advanced_publisher_delete_options_t *opts);
#endif
