/*******************************************************************************************
 *                              MPLSinIP eBPF
 * This file contains a BPF (Berkeley Packet Filter) for use within tc
 * (traffic-control).
 *
 * BPF is a virtual-machine within the Linux kernel that supports a limited
 * instruction set (not Turing complete). It allows user supplied code to be
 * executed during key points within the kernel. The kernel verifies all BPF
 * programs so that they don't address invalid memory & that the time spent in
 * the BPF program is limited by dis-allowing loops & setting a maximum number
 * of instructions.
 *
 *
 * ----------------------------------------------------------------------------------------
 *  eBPF Guide & Checklist
 *
 * 1. eBPF does not support method calls so any function called from the
 *    entry-point needs to be inlined.
 * 2. The kernel can JIT the eBPF however prior to 4.15, it was off by default
 *    (value 0). echo 1 > /proc/sys/net/core/bpf_jit_enable
 *
 * @author Farid Zakaria <farid.m.zakaria\@gmail.com>
 *******************************************************************************************/

#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <linux/udp.h>
#include <linux/pkt_cls.h>
#include <linux/bpf.h>

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

#include "helpers.h"
#include "types.h"
#include "debug.h"
#include "maps.h"
#include "egress.h"
#include "ingress.h"