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

#include "zenoh_cgo.h"

zc_cgo_bytes_data_t zc_cgo_bytes_get_data(const z_loaned_bytes_t *bytes) {
  if (bytes == NULL) {
    return (zc_cgo_bytes_data_t){.data = NULL, .bytes = NULL, .len = 0};
  }
  z_view_slice_t s;
  if (z_bytes_get_contiguous_view(bytes, &s) == Z_OK) {
    return (zc_cgo_bytes_data_t){.data = z_slice_data(z_loan(s)),
                                 .bytes = NULL,
                                 .len = z_slice_len(z_loan(s))};
  } else {
    return (zc_cgo_bytes_data_t){
        .data = NULL, .bytes = bytes, .len = z_bytes_len(bytes)};
  }
}
zc_cgo_string_data_t zc_cgo_string_get_data(const z_loaned_string_t *s) {
  return (zc_cgo_string_data_t){.str_ptr = z_string_data(s),
                                .len = z_string_len(s)};
}

zc_cgo_string_data_t
zc_cgo_keyexpr_get_data(const z_loaned_keyexpr_t *keyexpr) {
  z_view_string_t s;
  z_keyexpr_as_view_string(keyexpr, &s);
  return zc_cgo_string_get_data(z_loan(s));
}

zc_cgo_sample_data_t zc_cgo_sample_get_data(z_loaned_sample_t *sample) {
  return (zc_cgo_sample_data_t){
      .payload = zc_cgo_bytes_get_data(z_sample_payload(sample)),
      .encoding = zc_internal_encoding_get_data(z_sample_encoding(sample)),
      .attachment = zc_cgo_bytes_get_data(z_sample_attachment(sample)),
      .keyexpr = zc_cgo_keyexpr_get_data(z_sample_keyexpr(sample)),
      .timestamp = z_sample_timestamp(sample),
      .kind = z_sample_kind(sample)};
}

zc_cgo_query_data_t zc_cgo_query_get_data(z_loaned_query_t *query) {
  zc_cgo_query_data_t data = {0};
  const z_loaned_bytes_t *payload = z_query_payload(query);
  if (payload != NULL) {
    data.has_payload = true;
    data.payload = zc_cgo_bytes_get_data(payload);
  }

  const z_loaned_bytes_t *attachment = z_query_attachment(query);
  if (attachment != NULL) {
    data.has_attachment = true;
    data.attachment = zc_cgo_bytes_get_data(attachment);
  }
  data.keyexpr = zc_cgo_keyexpr_get_data(z_query_keyexpr(query));
  const z_loaned_encoding_t *encoding = z_query_encoding(query);
  if (encoding != NULL) {
    data.has_encoding = true;
    data.encoding = zc_internal_encoding_get_data(encoding);
  }
  z_view_string_t s;
  z_query_parameters(query, &s);
  data.params = zc_cgo_string_get_data(z_loan(s));
  z_query_clone(&data.query, query);
  return data;
}

zc_cgo_reply_data_t zc_cgo_reply_get_data(z_loaned_reply_t *reply) {
  const z_loaned_sample_t *sample = z_reply_ok(reply);
  if (sample != NULL) {
    return (zc_cgo_reply_data_t){
        .is_ok = true,
        .payload = zc_cgo_bytes_get_data(z_sample_payload(sample)),
        .encoding = zc_internal_encoding_get_data(z_sample_encoding(sample)),
        .attachment = zc_cgo_bytes_get_data(z_sample_attachment(sample)),
        .keyexpr = zc_cgo_keyexpr_get_data(z_sample_keyexpr(sample)),
        .timestamp = z_sample_timestamp(sample),
        .kind = z_sample_kind(sample)};
  } else {
    const z_loaned_reply_err_t *err = z_reply_err(reply);
    zc_cgo_reply_data_t out = {0};
    out.payload = zc_cgo_bytes_get_data(z_reply_err_payload(err));
    out.encoding = zc_internal_encoding_get_data(z_reply_err_encoding(err));
    return out;
  }
}

void zc_cgo_bytes_read_all(const z_loaned_bytes_t *bytes, uint8_t *out) {
  z_bytes_reader_t reader = z_bytes_get_reader(bytes);
  z_bytes_reader_read(&reader, out, z_bytes_len(bytes));
}

void zc_cgo_string_drop(z_owned_string_t *s) { z_drop(z_move(*s)); }

void zc_cgo_encoding_drop(z_owned_encoding_t *e) { z_drop(z_move(*e)); }

void zc_cgo_query_drop(z_owned_query_t *q) { z_drop(z_move(*q)); }

void zenohSubscriberCallback(struct z_loaned_sample_t *sample, void *context) {
  zenohSubscriberCallbackData(zc_cgo_sample_get_data(sample), context);
}

void zenohQueryableCallback(struct z_loaned_query_t *query, void *context) {
  zenohQueryableCallbackData(zc_cgo_query_get_data(query), context);
}

void zenohGetCallback(struct z_loaned_reply_t *reply, void *context) {
  zenohGetCallbackData(zc_cgo_reply_get_data(reply), context);
}

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

z_result_t zc_cgo_publisher_put(z_owned_publisher_t *publisher,
                                zc_cgo_bytes_data_t payload_data,
                                z_publisher_put_options_t *opts,
                                zc_internal_encoding_data_t *encoding_data,
                                zc_cgo_bytes_data_t *attachment_data) {
  z_owned_bytes_t payload, attachment;
  z_owned_encoding_t encoding;
  z_bytes_copy_from_buf(&payload, payload_data.data, payload_data.len);
  if (encoding_data != NULL) {
    opts->encoding = _create_moved_encoding_from_data(encoding_data, &encoding);
  }
  if (attachment_data != NULL) {
    opts->attachment =
        _create_moved_bytes_from_data(attachment_data, &attachment);
  }

  return z_publisher_put(z_loan(*publisher), z_move(payload), opts);
}

z_result_t zc_cgo_publisher_delete(z_owned_publisher_t *publisher,
                                   z_publisher_delete_options_t *opts) {
  return z_publisher_delete(z_loan(*publisher), opts);
}

z_result_t zc_cgo_put(z_owned_session_t *session,
                      zc_cgo_string_data_t keyexpr_data,
                      zc_cgo_bytes_data_t payload_data, z_put_options_t *opts,
                      zc_internal_encoding_data_t *encoding_data,
                      zc_cgo_bytes_data_t *attachment_data) {
  z_owned_bytes_t payload, attachment;
  z_owned_encoding_t encoding;
  z_bytes_copy_from_buf(&payload, payload_data.data, payload_data.len);
  z_view_keyexpr_t keyexpr;
  z_view_keyexpr_from_substr_unchecked(&keyexpr, keyexpr_data.str_ptr,
                                       keyexpr_data.len);
  if (encoding_data != NULL) {
    opts->encoding = _create_moved_encoding_from_data(encoding_data, &encoding);
  }
  if (attachment_data != NULL) {
    opts->attachment =
        _create_moved_bytes_from_data(attachment_data, &attachment);
  }
  return z_put(z_loan(*session), z_loan(keyexpr), z_move(payload), opts);
}

z_result_t zc_cgo_delete(z_owned_session_t *session,
                         zc_cgo_string_data_t keyexpr_data,
                         z_delete_options_t *opts) {
  z_view_keyexpr_t keyexpr;
  z_view_keyexpr_from_substr_unchecked(&keyexpr, keyexpr_data.str_ptr,
                                       keyexpr_data.len);
  return z_delete(z_loan(*session), z_loan(keyexpr), opts);
}

z_result_t zc_cgo_query_reply(z_owned_query_t *query,
                              zc_cgo_string_data_t keyexpr_data,
                              zc_cgo_bytes_data_t payload_data,
                              z_query_reply_options_t *opts,
                              zc_internal_encoding_data_t *encoding_data,
                              zc_cgo_bytes_data_t *attachment_data) {
  z_owned_bytes_t payload, attachment;
  z_owned_encoding_t encoding;
  z_bytes_copy_from_buf(&payload, payload_data.data, payload_data.len);
  z_view_keyexpr_t keyexpr;
  z_view_keyexpr_from_substr_unchecked(&keyexpr, keyexpr_data.str_ptr,
                                       keyexpr_data.len);
  if (encoding_data != NULL) {
    opts->encoding = _create_moved_encoding_from_data(encoding_data, &encoding);
  }
  if (attachment_data != NULL) {
    opts->attachment =
        _create_moved_bytes_from_data(attachment_data, &attachment);
  }
  return z_query_reply(z_loan(*query), z_loan(keyexpr), z_move(payload), opts);
}

z_result_t zc_cgo_query_reply_err(z_owned_query_t *query,
                                  zc_cgo_bytes_data_t payload_data,
                                  z_query_reply_err_options_t *opts,
                                  zc_internal_encoding_data_t *encoding_data) {
  z_owned_bytes_t payload;
  z_owned_encoding_t encoding;
  z_bytes_copy_from_buf(&payload, payload_data.data, payload_data.len);
  if (encoding_data != NULL) {
    opts->encoding = _create_moved_encoding_from_data(encoding_data, &encoding);
  }
  return z_query_reply_err(z_loan(*query), z_move(payload), opts);
}

z_result_t zc_cgo_query_reply_del(z_owned_query_t *query,
                                  zc_cgo_string_data_t keyexpr_data,
                                  z_query_reply_del_options_t *opts,
                                  zc_cgo_bytes_data_t *attachment_data) {
  z_view_keyexpr_t keyexpr;
  z_owned_bytes_t attachment;
  z_view_keyexpr_from_substr_unchecked(&keyexpr, keyexpr_data.str_ptr,
                                       keyexpr_data.len);
  if (attachment_data != NULL) {
    opts->attachment =
        _create_moved_bytes_from_data(attachment_data, &attachment);
  }
  return z_query_reply_del(z_loan(*query), z_loan(keyexpr), opts);
}

z_result_t zc_cgo_get(z_owned_session_t *session,
                      zc_cgo_string_data_t keyexpr_data, const char *params,
                      void *context, z_get_options_t *opts,
                      zc_cgo_bytes_data_t *payload_data,
                      zc_internal_encoding_data_t *encoding_data,
                      zc_cgo_bytes_data_t *attachment_data) {
  z_view_keyexpr_t keyexpr;
  z_view_keyexpr_from_substr_unchecked(&keyexpr, keyexpr_data.str_ptr,
                                       keyexpr_data.len);
  z_owned_closure_reply_t closure;
  z_closure(&closure, zenohGetCallback, zenohGetDrop, context);
  z_owned_bytes_t payload, attachment;
  z_owned_encoding_t encoding;
  if (payload_data != NULL) {
    opts->payload = _create_moved_bytes_from_data(payload_data, &payload);
  }
  if (encoding_data != NULL) {
    opts->encoding = _create_moved_encoding_from_data(encoding_data, &encoding);
  }
  if (attachment_data != NULL) {
    opts->attachment =
        _create_moved_bytes_from_data(attachment_data, &attachment);
  }
  return z_get(z_loan(*session), z_loan(keyexpr), params, z_move(closure),
               opts);
}

z_result_t zc_cgo_liveliness_get(z_owned_session_t *session,
                                 zc_cgo_string_data_t keyexpr_data,
                                 void *context,
                                 z_liveliness_get_options_t *opts) {
  z_view_keyexpr_t keyexpr;
  z_view_keyexpr_from_substr_unchecked(&keyexpr, keyexpr_data.str_ptr,
                                       keyexpr_data.len);
  z_owned_closure_reply_t closure;
  z_closure(&closure, zenohGetCallback, zenohGetDrop, context);
  return z_liveliness_get(z_loan(*session), z_loan(keyexpr), z_move(closure),
                          opts);
}

z_result_t zc_cgo_querier_get(z_owned_querier_t *querier, const char *params,
                              void *context, z_querier_get_options_t *opts,
                              zc_cgo_bytes_data_t *payload_data,
                              zc_internal_encoding_data_t *encoding_data,
                              zc_cgo_bytes_data_t *attachment_data) {
  z_owned_closure_reply_t closure;
  z_closure(&closure, zenohGetCallback, zenohGetDrop, context);
  z_owned_bytes_t payload, attachment;
  z_owned_encoding_t encoding;
  if (payload_data != NULL) {
    opts->payload = _create_moved_bytes_from_data(payload_data, &payload);
  }
  if (encoding_data != NULL) {
    opts->encoding = _create_moved_encoding_from_data(encoding_data, &encoding);
  }
  if (attachment_data != NULL) {
    opts->attachment =
        _create_moved_bytes_from_data(attachment_data, &attachment);
  }
  return z_querier_get(z_loan(*querier), params, z_move(closure), opts);
}
