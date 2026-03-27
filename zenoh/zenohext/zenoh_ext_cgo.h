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

typedef struct {
  const char *str_ptr;
  size_t len;
} zc_cgo_string_data_t;

typedef struct {
  const uint8_t
      *data; // will be null if there is more than one slice in z_bytes_t
  const z_loaned_bytes_t *bytes; // will be null if there is only one slice
  size_t len;
} zc_cgo_bytes_data_t;

typedef struct {
  zc_internal_encoding_data_t encoding;
  zc_cgo_string_data_t keyexpr;
  zc_cgo_bytes_data_t payload;
  zc_cgo_bytes_data_t attachment;
  const z_timestamp_t *timestamp;
  z_sample_kind_t kind;
  z_reliability_t reliability;
} zc_cgo_sample_data_t;

zc_cgo_bytes_data_t zc_cgo_bytes_get_data(const z_loaned_bytes_t *bytes);
zc_cgo_string_data_t zc_cgo_string_get_data(const z_loaned_string_t *s);
zc_cgo_string_data_t zc_cgo_keyexpr_get_data(const z_loaned_keyexpr_t *keyexpr);

zc_cgo_sample_data_t zc_cgo_sample_get_data(z_loaned_sample_t *sample);

extern void zenohSubscriberCallbackData(zc_cgo_sample_data_t sample,
                                        void *context);
extern void zenohSubscriberDrop(void *context);
void zenohSubscriberCallback(struct z_loaned_sample_t *sample, void *context);

extern void zenohMatchingListenerDrop(void *context);
typedef const struct z_matching_status_t zc_cgo_const_matching_status;
extern void zenohMatchingListenerCallback(zc_cgo_const_matching_status *status,
                                          void *context);

extern void zenohMissListenerDrop(void *context);
typedef const struct ze_miss_t zc_cgo_const_miss_t;
extern void zenohMissListenerCallback(zc_cgo_const_miss_t *miss, void *context);
void zenohMissListenerCCallback(const ze_miss_t *miss, void *context);

z_result_t
zc_cgo_advanced_publisher_put(ze_owned_advanced_publisher_t *publisher,
                              zc_cgo_bytes_data_t *payload_data,
                              ze_advanced_publisher_put_options_t *opts,
                              zc_internal_encoding_data_t *encoding_data,
                              zc_cgo_bytes_data_t *attachment_data);
z_result_t
zc_cgo_advanced_publisher_delete(ze_owned_advanced_publisher_t *publisher,
                                 ze_advanced_publisher_delete_options_t *opts);
#endif
