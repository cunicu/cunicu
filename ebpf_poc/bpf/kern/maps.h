/* Definition of maps */
#pragma once

#include "types.h"

// For including the type into the BTF output
// So bpf2go can use it for generating the type in our Go wrapper
// typedef struct state state;
const struct state *_unused_state __attribute__((unused));

/**
 * A really simple BPF map that controls a switch
 * whether the debug printk messages are emitted.
 */
struct bpf_elf_map SEC("maps") settings_map = {
  .type = BPF_MAP_TYPE_ARRAY,
  .size_key = sizeof(__u32),
  .size_value = sizeof(__u32),
  .max_elem = SETTING_LAST,
};

struct bpf_elf_map SEC("maps") ingress_map = {
  .type = BPF_MAP_TYPE_HASH_OF_MAPS,
  .size_key = sizeof(__u16),
  .size_value = sizeof(struct state),
  .max_elem = 1 << 12,
};

struct bpf_elf_map SEC("maps") egress_map = {
  .type = BPF_MAP_TYPE_HASH_OF_MAPS,
  .size_key = sizeof(__u16),
  .size_value = sizeof(struct state),
  .max_elem = 1 << 12,
};
