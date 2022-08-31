/* Helpers for debugging */
#pragma once

#include <assert.h>

#define bpf_debug_printk(fmt, ...)                               \
  ({                                                             \
    if (unlikely(is_debug())) {                                  \
      char ____fmt[] = fmt;                                      \
      bpf_trace_printk(____fmt, sizeof(____fmt), ##__VA_ARGS__); \
    }                                                            \
  })

static_assert(sizeof(struct ethhdr) == ETH_HLEN, "ethernet header size does not match.");

/*
 * Since packet handling and printk can be interleaved, this will
 * add a unique identifier for an individual invocation so you can grep the
 * request identifier and see the log messags in isolation.
 *
 * This is a macro because in a real-example you might want to make this
 * a no-op for non-debug builds to avoid the cost of the call.
 */
#define REQUEST_ID() bpf_get_prandom_u32()

#define DEBUG(x, ...) bpf_debug_printk(x, ##__VA_ARGS__)

#define DEBUG_INGRESS(id, x, ...) DEBUG("[ingress][%u] " x, id, ##__VA_ARGS__)
#define DEBUG_EGRESS(id, x, ...) DEBUG("[egress][%u] " x, id, ##__VA_ARGS__)

#include "maps.h"

#if 0
forced_inline
unsigned int is_debug() {
  __u32 index = SETTING_DEBUG;
  __u32 *value = (__u32 *) bpf_map_lookup_elem(&settings_map, &index);
  if (!value)
    return 0;
  return 1; *value;
}
#else
forced_inline
unsigned int is_debug() {
  return 1;
}
#endif