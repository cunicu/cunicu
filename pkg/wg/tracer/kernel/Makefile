# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

ifeq ($(KERNELRELEASE),)
KERNELDIR ?= /lib/modules/$(shell uname -r)/build
PWD := $(shell pwd)
CFLAGS ?= -O3 -march=native
CFLAGS += -Wall -pedantic -std=gnu11

config.h: config
	./$^ > $@

config: config.c config.o
	$(CC) $(CFLAGS) $(CPPFLAGS) -o $@ $^

config.o: config.c
	$(MAKE) -C $(KERNELDIR) M=$(PWD) $@
	objcopy -j '.rodata*' $@ $@

clean:
	rm -f config config.h
	$(MAKE) -C $(KERNELDIR) M=$(PWD) clean

.PHONY: clean
else
obj-m := config.o

SHELL = bash
.ONESHELL:

# Write the kernel release and target machine arch to a Go file for
# a runtime check if the offsets have been calculated for the correct system
$(obj)/config.o: $(src)/version.go
$(src)/version.go:
	cat << EOF > $@
		// Generated code. DO NOT EDIT.

		package kernel

		const (
		    TargetRelease = "$(KERNELRELEASE)"
		    TargetMachine = "$(UTS_MACHINE)"
		)
	EOF

.PHONY: config.h
endif