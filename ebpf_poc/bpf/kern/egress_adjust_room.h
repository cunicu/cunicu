/* Egress filter */
#pragma once

SEC("egress") int egress_filter(struct __sk_buff *skb)
{
  struct ethhdr *eth;
  struct iphdr *iph;
  struct udphdr *udp, udp_old;
  struct turn_cdata *cdata;
  struct state *state;
  
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
  eth = (struct ethhdr *) (data);
  if ((void *) (eth + 1) > data_end)
    return TC_ACT_SHOT;

  if (eth->h_proto != bpf_htons(ETH_P_IP))
    return TC_ACT_OK;

  iph = (struct iphdr *) (void *) (eth + 1);
  if ((void *) (iph + 1) > data_end)
    return TC_ACT_SHOT;

  if (iph->protocol != IPPROTO_UDP)
    return TC_ACT_OK;

  // multiply ip header by 4 (bytes) to get the number of bytes of the header.
  int iph_len = iph->ihl << 2;
  
  udp = (struct udphdr *) (void *) ((void *) (iph) + iph_len);
  if ((void *) (udp + 1) > data_end)
    return TC_ACT_SHOT;

  __u16 dport = bpf_ntohs(udp->dest);

  void *map = bpf_map_lookup_elem(&egress_map, &dport);
  if (map == NULL)
    return TC_ACT_OK;

  DEBUG_EGRESS(request_id, "found entry in egress: %d\n", bpf_ntohs(udp->dest));

  state = (struct state*) bpf_map_lookup_elem(map, &iph->daddr);
  if (state == NULL)
    return TC_ACT_OK;

  // Rewrite destination port
  if (state->lport != 0) {
    __u16 udp_old_port = udp->dest;
    __u16 udp_new_port = bpf_htons(state->lport);

    DEBUG_EGRESS(request_id, "rewriting destination port: %d => %d", bpf_ntohs(udp_old_port), bpf_ntohs(udp_new_port));

    udp->dest = udp_new_port;

    bpf_l4_csum_replace(skb, UDP_CSUM_OFF, udp_old_port, udp_new_port, 2);
  }

  if (state->channel_id != 0) {
    DEBUG_EGRESS(request_id, "inserting turn channel id: %d", state->channel_id);

    int padlen = sizeof(struct turn_cdata);

    data_end = (void *) (long) skb->data_end;
    data = (void *) (long) skb->data;
    udp = (struct udphdr *) (void *) (data + sizeof(struct ethhdr) + iph_len);
    if ((void *) (udp + 1) > data_end)
      return TC_ACT_SHOT;

    udp_old = *udp;

    // Make space for TURN channel data indication header
    int ret = bpf_skb_adjust_room(skb, padlen, BPF_ADJ_ROOM_NET, BPF_F_ADJ_ROOM_ENCAP_L3_IPV4 | BPF_F_ADJ_ROOM_ENCAP_L4_UDP);
    if (ret) {
      DEBUG_EGRESS(request_id, "failed bpf_skb_adjust_room");
      return TC_ACT_SHOT;
    }

    // Fix length field in IP header
    data_end = (void *) (long) skb->data_end;
    data = (void *) (long) skb->data;
    iph = (struct iphdr *) (void *) (data + sizeof(struct ethhdr));
    if ((void *) (iph + 1) > data_end)
      return TC_ACT_SHOT;

    __u16 iph_old_len = iph->tot_len;
    __u16 iph_new_len = bpf_htons(bpf_ntohs(iph->tot_len) + padlen);

    iph->tot_len = iph_new_len;

    // Adjust L3 checksum
    bpf_l3_csum_replace(skb, IP_CSUM_OFF, iph_old_len, iph_new_len, 2);

    // Update pointer to new UDP header
    data_end = (void *) (long) skb->data_end;
    data = (void *) (long) skb->data;
    udp = (struct udphdr *) (void *) (data + sizeof(struct ethhdr) + iph_len);
    if ((void *) (udp + 1) > data_end)
      return TC_ACT_SHOT;

    // Restore previous UDP header
    *udp = udp_old;

    // Fix length field in UDP header
    __u16 udp_old_len = udp_old.len;
    __u16 udp_new_len = bpf_htons(bpf_ntohs(udp->len) + padlen);
    udp->len = udp_new_len;
    
    // Construct TURN channel data indication header
    cdata = (struct turn_cdata*) (void *) (udp + 1);
    if ((void *) (cdata + 1) > data_end)
      return TC_ACT_SHOT;

    struct turn_cdata cd = {
      .ch_num = bpf_htons(0xAABB),
      .len = bpf_htons(0xCCDD)
    };

    *cdata = cd;

    // Adjust L4 checksum
    // bpf_l4_csum_replace(skb, UDP_CSUM_OFF, udp_old_len, udp_new_len, 2);
    // bpf_l4_csum_replace(skb, UDP_CSUM_OFF, iph_old_len, iph_new_len, BPF_F_PSEUDO_HDR | 2);
    // bpf_l4_csum_replace(skb, UDP_CSUM_OFF, 0, cd.ch_num, 2);
    // bpf_l4_csum_replace(skb, UDP_CSUM_OFF, 0, cd.len, 2);
    // bpf_l4_csum_replace(skb, UDP_CSUM_OFF, 0, 16, 2);
    
  }

  // return bpf_redirect(1, BPF_F_INGRESS);
  return TC_ACT_OK;
}

static char _license[] SEC("license") = "GPL";
