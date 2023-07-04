// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build ignore

#include "kernel/config.h"
#include "bpf_helpers.h"

char __license[] SEC("license") = "Dual MIT/GPL";

struct handshake {
	u64 ktime;

	u8 local_static_private_key[NOISE_PUBLIC_KEY_LEN];
	u8 local_ephemeral_private_key[NOISE_PUBLIC_KEY_LEN];
	u8 remote_static_public_key[NOISE_PUBLIC_KEY_LEN];
	u8 preshared_key[NOISE_SYMMETRIC_KEY_LEN];
};

struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 1 << 12);
} handshakes SEC(".maps");

// Force emitting struct event into the ELF.
const struct handshake *unused __attribute__((unused));

SEC("kprobe/wg_index_hashtable_insert")
int kprobe_wg_index_hashtable_insert(struct pt_regs *ctx) {
	// __le32 wg_index_hashtable_insert(struct index_hashtable *table, struct index_hashtable_entry *entry);

	struct handshake *rhs = bpf_ringbuf_reserve(&handshakes, sizeof(struct handshake), 0);
	if (!rhs)
		return 0;

	rhs->ktime = bpf_ktime_get_ns();

	// We are interested in the second argument "entry" which is of type struct noise_handshake
	char *nhs = (char *) PT_REGS_PARM2(ctx);
	char *static_identity;

	bpf_probe_read_kernel(&static_identity, sizeof(static_identity), nhs + LOCAL_STATIC_PRIVATE_KEY_IND_OFFSET);

	bpf_probe_read_kernel(rhs->local_static_private_key,    sizeof(rhs->local_ephemeral_private_key), static_identity + LOCAL_STATIC_PRIVATE_KEY_OFFSET);
	bpf_probe_read_kernel(rhs->remote_static_public_key,    sizeof(rhs->remote_static_public_key),    nhs + REMOTE_STATIC_PUBLIC_KEY_OFFSET);
	bpf_probe_read_kernel(rhs->preshared_key,               sizeof(rhs->preshared_key),               nhs + PRESHARED_KEY_OFFSET);
	bpf_probe_read_kernel(rhs->local_ephemeral_private_key, sizeof(rhs->local_ephemeral_private_key), nhs + LOCAL_EPHEMERAL_PRIVATE_KEY_OFFSET);

	bpf_ringbuf_submit(rhs, 0);

	return 0;
}