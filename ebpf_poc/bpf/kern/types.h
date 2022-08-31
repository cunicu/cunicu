#pragma once

#include <linux/types.h>

#define bool _Bool

// https://www.rfc-editor.org/rfc/rfc8489.html#section-5
const __u32 stun_cookie = 0x2112A442;

enum setting {
  SETTING_DEBUG = 0,
  SETTING_LAST
};

struct state {
  __u16 channel_id;
  __u16 lport;
};

struct stun_hdr {
    __be16 msg_type;
    __be16 len;
    __u32 cookie;
    __u32 tid[3];
};

// https://www.rfc-editor.org/rfc/rfc8656.html#name-the-channeldata-message
struct turn_cdata {
    __be16 ch_num;
    __be16 len;
};

/* 
 * ELF map definition used by iproute2.
 * Cannot figure out how to get bpf_elf.h installed on system, so we've copied it here.
 * iproute2 claims this struct will remain backwards compatible
 * https://github.com/kinvolk/iproute2/blob/be55416addf76e76836af6a4dd94b19c4186e1b2/include/bpf_elf.h
 */
struct bpf_elf_map {
	/*
	 * The various BPF MAP types supported (see enum bpf_map_type)
	 * https://elixir.bootlin.com/linux/latest/source/include/uapi/linux/bpf.h
	 */
	__u32 type;
	__u32 size_key;
	__u32 size_value;
	__u32 max_elem;
	/*
	 * Various flags you can place such as `BPF_F_NO_COMMON_LRU`
	 */
	__u32 flags;
	__u32 id;
	/*
	 * Pinning is how the map are shared across process boundary.
	 * Cillium has a good explanation of them: http://docs.cilium.io/en/v1.3/bpf/#llvm
	 * PIN_GLOBAL_NS - will get pinned to `/sys/fs/bpf/tc/globals/${variable-name}`
	 * PIN_OBJECT_NS - will get pinned to a directory that is unique to this object
	 * PIN_NONE - the map is not placed into the BPF file system as a node,
	 			  and as a result will not be accessible from user space
	 */
	__u32 pinning;
};
