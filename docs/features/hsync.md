---
# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Hosts-file Synchronization

The hosts-file synchronization updates your local [hosts(5)](https://man7.org/linux/man-pages/man5/hosts.5.html) file (`/etc/hosts`) with entries for each peer.

As hostname, cunicu uses the first 8 characters of the Base64-encoded public key as well as an optional hostname.
This optional hostname can either be configured by the user in the configuration file or is discovered via the [peer-discovery feature](./pdisc.md).

## Example

The following snippet shows the local hosts file of an Ubuntu 20.04 system with two entries added by cunicu.
As shown here, all entries managed by cunicu are marked with a comment prefixed with `# cunicu:`

```bash title="/etc/hosts"
127.0.0.1 localhost
127.0.1.1 ubuntu

# The following lines are desirable for IPv6 capable hosts
::1     ip6-localhost ip6-loopback
fe00::0 ip6-localnet
ff00::0 ip6-mcastprefix
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters

fe80::13a9:c799:cead:4f28 buxfBfaN.wg-local fra-1.wg-local # cunicu: ifname=wg0, ifindex=9, pk=buxfBfaNZI8UFT0cB1aj9YanhbLfxlTfd/hH3DrGaFA=
fe80::1fed:fabb:a9f6:d78 ZEki/XKE.wg-local # cunicu: ifname=wg1, ifindex=10, pk=ZEki/XKEsqdjFyURo5Sm+g3vXSKJKpV5WmwWKAQqo2c=
```
