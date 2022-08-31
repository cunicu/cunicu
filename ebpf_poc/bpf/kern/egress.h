/* Egress filter */
#pragma once

#define MAX_MTU 40
#define MAX_PACKET_OFF 0xffff

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

    int pad_len = sizeof(struct turn_cdata);

    data_end = (void *) (long) skb->data_end;
    data = (void *) (long) skb->data;
    udp = (struct udphdr *) (void *) (data + sizeof(struct ethhdr) + iph_len);
    if ((void *) (udp + 1) > data_end)
      return TC_ACT_SHOT;

    __u16 newlen = sizeof(struct ethhdr) + iph_len + bpf_ntohs(udp->len) + pad_len;

    // Make space for TURN channel data indication header
    int ret = bpf_skb_change_tail(skb, newlen, 0);
    if (ret) {
      DEBUG_EGRESS(request_id, "failed bpf_skb_change_tail");
      return TC_ACT_SHOT;
    }

    // Fix length field in IP header
    data_end = (void *) (long) skb->data_end;
    data = (void *) (long) skb->data;
    iph = (struct iphdr *) (void *) (data + sizeof(struct ethhdr));
    if ((void *) (iph + 1) > data_end)
      return TC_ACT_SHOT;

    __u16 iph_old_len = iph->tot_len;
    __u16 iph_new_len = bpf_htons(bpf_ntohs(iph->tot_len) + pad_len);

    iph->tot_len = iph_new_len;

    // Adjust L3 checksum
    bpf_l3_csum_replace(skb, IP_CSUM_OFF, iph_old_len, iph_new_len, 2);

    // Update pointer to new UDP header
    data_end = (void *) (long) skb->data_end;
    data = (void *) (long) skb->data;
    udp = (struct udphdr *) (void *) (data + sizeof(struct ethhdr) + iph_len);
    if ((void *) (udp + 1) > data_end) {
      DEBUG_EGRESS(request_id, "drop");
      return TC_ACT_SHOT;
    }

    // Fix length field in UDP header
    __u16 udp_old_len = udp->len;
    __u16 udp_old_len_h = bpf_ntohs(udp_old_len);
    __u16 udp_new_len = bpf_htons(udp_old_len_h + pad_len);
    udp->len = udp_new_len;

#if 0
    __u16 pl_len = udp_old_len_h - sizeof(struct udphdr);
    char *pl = (char *) (udp + 1);

    char buf[256];
    __u16 pl_off = sizeof(struct ethhdr) + sizeof(struct iphdr) + sizeof(struct udphdr);

    if (pl_len > 256)
      return TC_ACT_SHOT;

    ret = bpf_skb_load_bytes(skb, pl_off, buf, 5);
    if (ret) {
       DEBUG_EGRESS(request_id, "failed bpf_skb_load_bytes");
       return TC_ACT_SHOT;
    }

    ret = bpf_skb_store_bytes(skb, pl_off + pad_len, buf, 5, 0);
    if (ret) {
       DEBUG_EGRESS(request_id, "failed bpf_skb_store_bytes");
      return TC_ACT_SHOT;
    }
#else
  __u16 pl_len = bpf_ntohs(udp->len) - sizeof(struct udphdr) - pad_len;

  DEBUG_EGRESS(request_id, "pad_len %d", pad_len);
  DEBUG_EGRESS(request_id, "pl_len %d", pl_len);
  DEBUG_EGRESS(request_id, "data_end %u", (__u64) data_end);
  DEBUG_EGRESS(request_id, "pl %u", (__u64) pl);

  data_end = (void *) (long) skb->data_end;
  data = (void *) (long) skb->data;

  char *pl = (char *) (udp + 1);

  __u32 *src = (__u32 *) pl;
  __u32 *dst = (__u32 *) (pl + pad_len);
  __u32 temp = *dst;

  for(__u32 i = 0; i < MAX_MTU; i += sizeof(temp)) {
    if (i >= pl_len) {
      break;
    }

    if ((void *) (dst + 1) > data_end)
      return TC_ACT_SHOT;

    if ((void *) (src + 1) > data_end)
      return TC_ACT_SHOT;

    *dst++ = temp;
    temp = *dst; 

    DEBUG_EGRESS(request_id, "assign %u = %u", (__u64) (dst), (__u64)(src));
  }
#endif

    // Construct TURN channel data indication header
    cdata = (struct turn_cdata *) pl;
    if ((void *) (cdata + 1) > data_end)
      return TC_ACT_SHOT;

    struct turn_cdata cd = {
      .ch_num = bpf_htons(0xAABB),
      .len = bpf_htons(0xCCDD)
    };

    *cdata = cd;

    // Adjust L4 checksum
    bpf_l4_csum_replace(skb, UDP_CSUM_OFF, udp_old_len, udp_new_len, 2);
    bpf_l4_csum_replace(skb, UDP_CSUM_OFF, iph_old_len, iph_new_len, BPF_F_PSEUDO_HDR | 2);
    // bpf_l4_csum_replace(skb, UDP_CSUM_OFF, 0, cd.ch_num, 2);
    // bpf_l4_csum_replace(skb, UDP_CSUM_OFF, 0, cd.len, 2);
    // bpf_l4_csum_replace(skb, UDP_CSUM_OFF, 0, 16, 2);
    
  }

  // return bpf_redirect(1, BPF_F_INGRESS);
  return TC_ACT_OK;
}

static char _license[] SEC("license") = "GPL";
