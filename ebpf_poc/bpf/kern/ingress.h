/* Egress filter */
#pragma once

/*
 * The Internet Protocol (IP) is defined in RFC 791.
 * The RFC specifies the format of the IP header.
 * In the header there is the IHL (Internet Header Length) field which is 4bit
 * long
 * and specifies the header length in 32bit words.
 * The IHL field can hold values from 0 (Binary 0000) to 15 (Binary 1111).
 * 15 * 32bits = 480bits = 60 bytes
 */
#define MAX_IP_HDR_LEN 60

SEC(".ingress") int ingress_filter(struct __sk_buff *skb) {
  // Generate a unique request id so we can identify each flow in
  // the trace logs
  unsigned long long request_id = REQUEST_ID();

  /*
   * the redundant casts are needed according to the documentation.
   * possibly for the BPF verifier.
   * https://www.spinics.net/lists/xdp-newbies/msg00181.html
   */
  void *data_end = (void *) (long) skb->data_end;
  void *data = (void *) (long) skb->data;

  // The packet starts with the ethernet header, so let's get that going:
  struct ethhdr *eth = (struct ethhdr *)(data);

  /*
   * Now, we can't just go "eth->h_proto", that's illegal.  We have to
   * explicitly test that such an access is in range and doesn't go
   * beyond "data_end" -- again for the verifier.
   * The eBPF verifier will see that "eth" holds a packet pointer,
   * and also that you have made sure that from "eth" to "eth + 1"
   * is inside the valid access range for the packet.
   */
  if ((void *)(eth + 1) > data_end) {
    return TC_ACT_SHOT;
  }

  /*
   * We only care about IP packet frames. Don't do anything to other ethernet
   * packets like ARP.
   * hton -> host to network order. Network order is always big-endian.
   * pedantic: the protocol is also directly accessible from __sk_buf
   */
  if (eth->h_proto != bpf_htons(ETH_P_IP)) {
    return TC_ACT_OK;
  }

  struct iphdr *iph = (struct iphdr *)(void *)(eth + 1);

  if ((void *)(iph + 1) > data_end) {
    return TC_ACT_SHOT;
  }

  // multiply ip header by 4 (bytes) to get the number of bytes of the header.
  int iph_len = iph->ihl << 2;
  if (iph_len > MAX_IP_HDR_LEN) {
    return TC_ACT_SHOT;
  }

  if (iph->protocol != IPPROTO_UDP) {
    return TC_ACT_OK;
  }

  struct udphdr *udp = (struct udphdr *)((void *)(iph) + iph_len);

  if ((void *)(udp + 1) > data_end) {
    return TC_ACT_SHOT;
  }

  if (udp->dest != bpf_htons(2222))
    return TC_ACT_OK;

  DEBUG_INGRESS(request_id, "found ingress data!!!!!!!.\n");
  
  /*
   * This is the amount of padding we need to remove to be just left
   * with eth * iphdr.
   */
  // int padlen = sizeof(struct turn_cdata);

  /*
   * Grow or shrink the room for data in the packet associated to
   * skb by length and according to the selected mode.
   * BPF_ADJ_ROOM_NET: Adjust room at the network layer
   *  (room space is added or removed below the layer 3 header).
   */
  // int ret = bpf_skb_adjust_room(skb, -padlen, BPF_ADJ_ROOM_NET, 0);
  // if (ret) {
  //   DEBUG_INGRESS(request_id, "error calling skb adjust room.\n");
  //   return TC_ACT_SHOT;
  // }

  return TC_ACT_OK;
}
