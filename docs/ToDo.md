# TODOs

-   Encrypt all signaling messages

-   Add peer discovery

-   Add libp2p backend

-   Contribute code into existing packages
    -   wgctrl
        -   Watch for interfaces
        -   Configuration parsing
        -   Device Dump()

-   Investigate distributed management of Acccess Control Lists
    -   <https://ieeexplore.ieee.org/document/1437269>
    -   <https://www.it.iitb.ac.in/~madhumita/access/gcs/A%20Trust%20based%20Access%20Control%20Framework%20for%20P2P%20File%20Sharing%20Systems.pdf>
    -   <https://www.springerprofessional.de/en/decentralized-access-control-technique-with-multi-tier-authentic/19543988>
    -   <https://link.springer.com/chapter/10.1007%2F978-3-319-28865-9_28>

-   Single socket per Wireguard interface / ICE Agent
    -   Pass traffic in-process between userspace Wireguard and ICE sockets
    -   Use Wireguard-go's conn.Bind interface

-   Use in-process pipe for wireguard-go's UAPI (Bind interface)

-   Support eBPF proxying for relay candidates
    -   transform UDP packets with eBPF programm to insert/strip TURN channel IDs

    -   Related:
        -   <https://lwn.net/Articles/708020/>
        -   <https://blogs.oracle.com/linux/post/bpf-using-bpf-to-do-packet-transformation>
        -   <https://github.com/coturn/coturn/issues/759>

-   Update proxy instances instead of recreating them.
    -   Avoids possible packet loss during change of candidate pairs

-   Add better proxy implementations for OpenBSD, FreeBSD, Android and Windows

-   Test co-existance of multipe `wice` instances
    -   nft tables might collide

-   Use netlink multicast subscription for notification of Wireguard peer changes
    -   [Patch](https://lore.kernel.org/patchwork/patch/1366219/)

-   Use netlink multicast group RTMGRP_LINK to for notification of new Wireguard interfaces

-   Add links to code in README

-   Add sub-commands for controlling `wice` deaemon:
    -   `wice show [[INTF] [PEER]]`
    -   `wice add INTF`
    -   `wice delete INTF`
    -   `wice discover INTF GROUP`
    -   `wice sync [INTF]`
    -   `wice restart INTF PEER`
    -   `wice monitor`

-   Add check for availability of turnserver

-   Add check for handshakes before attempting to ping

-   Add context for waiting for events

-   Use mermaid actor diagram for signaling docs

-   Use RTT & packet loss for selecting ICE relay candidates

-   Embed TURN/STUN server into WICE
    -   Checkout pion/stun & pion/turn

-   Embed routing daemon into WICE
    -   Facilitates reachability for non-directly connected peers

    -   Candidates:
        -   [bio-rd](https://github.com/bio-routing/bio-rd)
        -   [gobgp](https://github.com/osrg/gobgp/)

-   Add new signaling backend by contacting already connected peers via through Wireguard via gRPC
