// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build ignore

struct define {
	const char *name;
	unsigned long off;
};
extern const struct define defines[];

#ifdef __KERNEL__

#include <generated/utsrelease.h>
#include "../drivers/net/wireguard/noise.h"

const struct define defines[] = {
	{ "NOISE_PUBLIC_KEY_LEN", NOISE_PUBLIC_KEY_LEN },
	{ "NOISE_SYMMETRIC_KEY_LEN", NOISE_SYMMETRIC_KEY_LEN },
	{ "LOCAL_STATIC_PRIVATE_KEY_OFFSET", offsetof(struct noise_static_identity, static_private) },
    { "LOCAL_STATIC_PRIVATE_KEY_IND_OFFSET", offsetof(struct noise_handshake, static_identity) },
	{ "LOCAL_EPHEMERAL_PRIVATE_KEY_OFFSET", offsetof(struct noise_handshake, ephemeral_private) },
	{ "REMOTE_STATIC_PUBLIC_KEY_OFFSET", offsetof(struct noise_handshake, remote_static) },
	{ "PRESHARED_KEY_OFFSET", offsetof(struct noise_handshake, preshared_key) },
	{ NULL, 0 }
};

#else

#include <stdio.h>

int main(int argc, char *argv[])
{
	puts("// Generated code. DO NOT EDIT.");
	puts("");

	for (const struct define *d = defines; d->name; ++d)
		printf("#define %s %lu\n", d->name, d->off);

	return 0;
}

#endif