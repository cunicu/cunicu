#!/bin/bash

# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

# This script demonstrates the usage of the WireGuard handshake tracing feature.
# cunicu uses an eBPF program attached to a Kprobe to extract ephemeral keys from the
# Linux kernel via a ringbuffer. These keys are required to fully decrypt and dissect
# WireGuard trafic.
# 
# Please note, that the handshake tracing feature is not enabled by default as it requires
# Linux kernel sources to be built since the Kprobe requires detailed knowledge of the in-
# kernel memory layout.
#
# Please point the KERNELDIR environment variable to a directory containing the
# **full** kernel sources (headers are not sufficient) while running "go generate":
#
#  $ export KERNELDIR=/usr/src/linux-5.15
#  $ git clone github.com/stv0g/cunicu
#  $ cd cunicu
#  $ go generate -tags tracer 
#  $ go build -tags tracer ./cmd/cunicu
#  $ sudo ./cunicu wg 

set -e

# Generate Keys
PSK=$(wg genpsk)

SK_LEFT=$(wg genkey)
SK_RIGHT=$(wg genkey)

PK_LEFT=$(echo ${SK_LEFT} | wg pubkey)
PK_RIGHT=$(echo ${SK_RIGHT} | wg pubkey)

echo "=== Generated WireGuard keys"
echo
echo "  Interface wg-left:"
echo "    PrivateKey: ${SK_LEFT}"
echo "    PublicKey:  ${PK_LEFT}"
echo
echo "  Interface wg-right:"
echo "    PrivateKey: ${SK_RIGHT}"
echo "    PublicKey:  ${PK_RIGHT}"
echo
echo "  PresharedKey: ${PSK}"
echo

TMP_FILE=$(mktemp /tmp/cunicu-XXXXXX)
PCAP_FILE="${TMP_FILE}.pcapng"
KEYS_FILE="${TMP_FILE}.keys"

# Cleanup stuff from previous runs
(
    ip link delete wg-left
    ip link delete wg-right
    ip netns delete cunicu-left 
    ip netns delete cunicu-right
) 1> /dev/null 2>&1 || true

function cleanup() {
    rm "${KEYS_FILE}" "${PCAP_FILE}"
    kill ${TRACER_PID} ${TSHARK_PID} 2> /dev/null
}
trap cleanup EXIT

echo -e "\n=== Start probing for WireGuard handshakes"
cunicu wg extract-handshakes 2> /dev/null > "${KEYS_FILE}" &
TRACER_PID=$!

echo -e "\n=== Start tshark capture"
tshark -i lo -w "${PCAP_FILE}" udp port 51820 or udp port 51821 &
TSHARK_PID=$!

# Wait until tshark is actually ready to capture packets
sleep 2

echo -e "\n=== Network link and ns setup"

# Create WireGuard interfaces
ip link add wg-left type wireguard
ip link add wg-right type wireguard

# Configure WireGuard interfaces
wg set wg-left  listen-port 51820 private-key <(echo ${SK_LEFT})  peer ${PK_RIGHT} preshared-key <(echo ${PSK}) endpoint 127.0.0.1:51821 allowed-ips 10.0.0.2/32
wg set wg-right listen-port 51821 private-key <(echo ${SK_RIGHT}) peer ${PK_LEFT}  preshared-key <(echo ${PSK}) endpoint 127.0.0.1:51820 allowed-ips 10.0.0.1/32

ip netns add cunicu-left
ip netns add cunicu-right

ip link set wg-left  netns cunicu-left
ip link set wg-right netns cunicu-right

ip -n cunicu-left  addr add dev wg-left  10.0.0.1/24
ip -n cunicu-right addr add dev wg-right 10.0.0.2/24

ip -n cunicu-left  link set up dev wg-left
ip -n cunicu-right link set up dev wg-right

# Generate some traffic via the WireGuard interface
echo -e "\n=== Running ping"
ip netns exec cunicu-left ping -c 3 10.0.0.2

echo -e "\n=== Stopping tshark and handshake tracer"
kill ${TSHARK_PID} ${TRACER_PID}

echo -e "\n=== WireGuard keys"
cat "${KEYS_FILE}"

echo -e "\n=== Decrypted capture"
tshark -r "${PCAP_FILE}" -o "wg.keylog_file:${KEYS_FILE}" -o wg.dissect_packet:true