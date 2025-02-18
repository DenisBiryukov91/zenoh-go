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

#ifndef ZENOH_CGO_H
#define ZENOH_CGO_H
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
} zc_cgo_sample_data_t;

typedef struct {
  zc_internal_encoding_data_t encoding;
  zc_cgo_string_data_t keyexpr;
  zc_cgo_bytes_data_t payload;
  zc_cgo_bytes_data_t attachment;
  zc_cgo_string_data_t params;
  z_owned_query_t query;
  bool has_encoding;
  bool has_payload;
  bool has_attachment;
} zc_cgo_query_data_t;

typedef struct {
  bool is_ok;
  zc_internal_encoding_data_t encoding;
  zc_cgo_string_data_t keyexpr;
  zc_cgo_bytes_data_t payload;
  zc_cgo_bytes_data_t attachment;
  const z_timestamp_t *timestamp;
  z_sample_kind_t kind;
} zc_cgo_reply_data_t;

zc_cgo_bytes_data_t zc_cgo_bytes_get_data(const z_loaned_bytes_t *bytes);
zc_cgo_string_data_t zc_cgo_string_get_data(const z_loaned_string_t *s);
zc_cgo_string_data_t zc_cgo_keyexpr_get_data(const z_loaned_keyexpr_t *keyexpr);

zc_cgo_sample_data_t zc_cgo_sample_get_data(z_loaned_sample_t *sample);
zc_cgo_query_data_t zc_cgo_query_get_data(z_loaned_query_t *query);
zc_cgo_reply_data_t zc_cgo_reply_get_data(z_loaned_reply_t *reply);

void zc_cgo_bytes_read_all(const z_loaned_bytes_t *bytes, uint8_t *out);

void zc_cgo_string_drop(z_owned_string_t *s);
void zc_cgo_encoding_drop(z_owned_encoding_t *e);
void zc_cgo_query_drop(z_owned_query_t *q);

extern void zenohSubscriberCallbackData(zc_cgo_sample_data_t sample,
                                        void *context);
extern void zenohSubscriberDrop(void *context);
void zenohSubscriberCallback(struct z_loaned_sample_t *sample, void *context);

extern void zenohQueryableCallbackData(zc_cgo_query_data_t query,
                                       void *context);
extern void zenohQueryableDrop(void *context);
void zenohQueryableCallback(struct z_loaned_query_t *query, void *context);

extern void zenohGetCallbackData(zc_cgo_reply_data_t query, void *context);
extern void zenohGetDrop(void *context);
void zenohGetCallback(struct z_loaned_reply_t *reply, void *context);

z_result_t zc_cgo_publisher_put(z_owned_publisher_t *publisher,
                                zc_cgo_bytes_data_t payload_data,
                                z_publisher_put_options_t *opts,
                                zc_internal_encoding_data_t *encoding_data,
                                zc_cgo_bytes_data_t *attachment_data);
z_result_t zc_cgo_publisher_delete(z_owned_publisher_t *publisher,
                                   z_publisher_delete_options_t *opts);
z_result_t zc_cgo_put(z_owned_session_t *session,
                      zc_cgo_string_data_t keyexpr_data,
                      zc_cgo_bytes_data_t payload_data, z_put_options_t *opts,
                      zc_internal_encoding_data_t *encoding_data,
                      zc_cgo_bytes_data_t *attachment_data);
z_result_t zc_cgo_delete(z_owned_session_t *session,
                         zc_cgo_string_data_t keyexpr_data,
                         z_delete_options_t *opts);
z_result_t zc_cgo_query_reply(z_owned_query_t *query,
                              zc_cgo_string_data_t keyexpr_data,
                              zc_cgo_bytes_data_t payload_data,
                              z_query_reply_options_t *opts,
                              zc_internal_encoding_data_t *encoding_data,
                              zc_cgo_bytes_data_t *attachment_data);
z_result_t zc_cgo_query_reply_err(z_owned_query_t *query,
                                  zc_cgo_bytes_data_t payload_data,
                                  z_query_reply_err_options_t *opts,
                                  zc_internal_encoding_data_t *encoding_data);
z_result_t zc_cgo_query_reply_del(z_owned_query_t *query,
                                  zc_cgo_string_data_t keyexpr_data,
                                  z_query_reply_del_options_t *opts,
                                  zc_cgo_bytes_data_t *attachment_data);
z_result_t zc_cgo_get(z_owned_session_t *session,
                      zc_cgo_string_data_t keyexpr_data, const char *params,
                      void *context, z_get_options_t *opts,
                      zc_cgo_bytes_data_t *payload_data,
                      zc_internal_encoding_data_t *encoding_data,
                      zc_cgo_bytes_data_t *attachment_data);
z_result_t zc_cgo_liveliness_get(z_owned_session_t *session,
                                 zc_cgo_string_data_t keyexpr_data,
                                 void *context,
                                 z_liveliness_get_options_t *opts);
z_result_t zc_cgo_querier_get(z_owned_querier_t *querier, const char *params,
                              void *context, z_querier_get_options_t *opts,
                              zc_cgo_bytes_data_t *payload_data,
                              zc_internal_encoding_data_t *encoding_data,
                              zc_cgo_bytes_data_t *attachment_data);
#endif
