// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

/* Limited set of BPF related helpers */

#pragma once

typedef unsigned char u8;
typedef unsigned int u32;
typedef long long unsigned int u64;

#define __uint(name, val) int (*name)[val]

#define BPF_MAP_TYPE_RINGBUF 27

/*
 * Helper macro to place programs, maps, license in
 * different sections in elf_bpf file. Section names
 * are interpreted by libbpf depending on the context (BPF programs, BPF maps,
 * extern variables, etc).
 * To allow use of SEC() with externs (e.g., for extern .maps declarations),
 * make sure __attribute__((unused)) doesn't trigger compilation warning.
 */
#define SEC(name) \
	_Pragma("GCC diagnostic push")					    \
	_Pragma("GCC diagnostic ignored \"-Wignored-attributes\"")	    \
	__attribute__((section(name), used))				    \
	_Pragma("GCC diagnostic pop")


static u64 (*bpf_ktime_get_ns)(void) = (void *) 5;
static long (*bpf_probe_read_kernel)(void *dst, u32 size, const void *unsafe_ptr) = (void *) 113;
static void *(*bpf_ringbuf_reserve)(void *ringbuf, u64 size, u64 flags) = (void *) 131;
static void (*bpf_ringbuf_submit)(void *data, u64 flags) = (void *) 132;

#if defined(__TARGET_ARCH_x86)
	#define bpf_target_x86
	#define bpf_target_defined
#elif defined(__TARGET_ARCH_arm)
	#define bpf_target_arm
	#define bpf_target_defined
#elif defined(__TARGET_ARCH_arm64)
	#define bpf_target_arm64
	#define bpf_target_defined
#endif

#ifndef bpf_target_defined
#error "Must specify a BPF target arch"
#endif

#if defined(bpf_target_x86)

struct pt_regs {
	/*
	 * C ABI says these regs are callee-preserved. They aren't saved on kernel entry
	 * unless syscall needs a complete, fully filled "struct pt_regs".
	 */
    u64 r15;
    u64 r14;
    u64 r13;
    u64 r12;
    u64 rbp;
    u64 rbx;
	/* These regs are callee-clobbered. Always saved on kernel entry. */
	u64 r11;
	u64 r10;
	u64 r9;
	u64 r8;
	u64 rax;
	u64 rcx;
	u64 rdx;
	u64 rsi;
	u64 rdi;
	/*
	 * On syscall entry, this is syscall#. On CPU exception, this is error code.
	 * On hw interrupt, it's IRQ number:
	 */
	u64 orig_rax;
	/* Return frame for iretq */
	u64 rip;
	u64 cs;
	u64 eflags;
	u64 rsp;
	u64 ss;
	/* top of stack page */
};

#define PT_REGS_PARM2(x) ((x)->rsi)

#elif defined(bpf_target_arm)

#define PT_REGS_PARM2(x) ((x)->uregs[1])

#elif defined(bpf_target_arm64)

struct pt_regs;
struct user_pt_regs {
	u64 regs[31];
	u64 sp;
	u64 pc;
	u64 pstate;
};

#define PT_REGS_ARM64 const volatile struct user_pt_regs
#define PT_REGS_PARM2(x) (((PT_REGS_ARM64 *)(x))->regs[1])

#endif /* defined(bpf_target_arm64) */
