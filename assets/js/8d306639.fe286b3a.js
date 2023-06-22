"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[7833],{2810:(e,t,n)=>{n.d(t,{Z:()=>a});var i=n(7294),s=n(7926);const o="# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>\n# SPDX-License-Identifier: Apache-2.0\n\n# This is an example of a simple cunicu configuration file.\n# For a full example please look at cunicu.advanced.yaml\n\n\n## WireGuard interface settings\n#\n# These settings configure WireGuard specific settings\n# of the interface.\n#\n# The following settings can be overwritten for each interface\n# using the 'interfaces' settings (see below).\n# The following settings will be used as default.\n\n# A base64 private key generated by wg genkey.\n# Will be automatically generated if not provided.\nprivate_key: KLoqDLKgoqaUkwctTd+Ov3pfImOfadkkvTdPlXsuLWM=\n\n# The remote WireGuard peers provided as a dictionary\n# The keys of this dictionary are used as names for the peers\npeers:  \n  test:\n    # A base64 public key calculated by wg pubkey from a private key,\n    # and usually transmitted out of band\n    # to the author of the configuration file.\n    public_key: FlKHqqQQx+bTAq7+YhwEECwWRg2Ih7NQ48F/SeOYRH8=\n\n    # A base64 pre-shared key generated by wg genpsk.\n    # Optional, and may be omitted.\n    # This option adds an additional layer of symmetric-key\n    # cryptography to be mixed into the already existing\n    # public-key cryptography, for post-quantum resistance.\n    preshared_key: zu86NBVsWOU3cx4UKOQ6MgNj3gv8GXsV9ATzSemdqlI=\n\n    # An endpoint IP or hostname, followed by a colon,\n    # and then a port number. This endpoint will be updated\n    # automatically to the most recent source IP address and\n    # port of correctly authenticated packets from the peer.\n    # If provided, no endpoint discovery will be performed.\n    endpoint: vpn.example.com:51820\n\n    # A time duration, between 1 and 65535s inclusive, of how\n    # often to send an authenticated empty packet to the peer\n    # for the purpose of keeping a stateful firewall or NAT mapping\n    # valid persistently. For example, if the interface very rarely\n    # sends traffic, but it might at anytime receive traffic from a\n    # peer, and it is behind NAT, the interface might benefit from\n    # having a persistent keepalive interval of 25 seconds.\n    # If set to zero, this option is disabled.\n    # By default or when unspecified, this option is off.\n    # Most users will not need this. Optional.\n    persistent_keepalive: 120s\n\n    # A comma-separated list of IP (v4 or v6) addresses with\n    # CIDR masks from which incoming traffic for this peer is\n    # allowed and to which outgoing  traffic for this peer is directed.\n    # The catch-all 0.0.0.0/0 may be specified for matching\n    # all IPv4 addresses, and ::/0 may be specified for matching\n    # all IPv6 addresses. May be specified multiple times.\n    allowed_ips:\n    - 192.168.5.0/24\n\n## Basic interface settings\n#\n\n# The Maximum Transfer Unit of the WireGuard interface.\n# If not specified, the MTU is automatically determined from\n# the endpoint addresses or the system default route,\n# which is usually a sane choice.\n# However, to manually specify an MTU to override this\n# automatic discovery, this value may be specified explicitly.\nmtu: 1420\n\n# A list of IP (v4 or v6) addresses (optionally with CIDR masks)\n# to be assigned to the interface.\n# May be specified multiple times.\naddresses:\n- 10.10.0.1/24\n\n# A list of prefixes which cunicu uses to derive local addresses\n# from the interfaces public key\nprefixes:\n- fc2f:9a4d::/32\n- 10.237.0.0/16\n\n## Peer discovery\n#\n# Peer discovery finds new peers within the same community and adds them to the respective interface\ndiscover_peers: true\n\n# The hostname which gets advertised to remote peers\nhostname: my-node\n\n# A passphrase shared among all peers of the same community\ncommunity: \"some-common-password\"\n\n# Networks which are reachable via this peer and get advertised to remote peers\n# These will be part of this interfaces AllowedIPs at the remote peers.\nnetworks:\n- 192.168.1.0/24\n- 10.2.0.0/24\n\n\n## Endpoint discovery\n#\n# Endpoint discovery uses Interactive Connectivity Establishment (ICE) as used by WebRTC to\n# gather a list of candidate endpoints and performs connectivity checks to find a suitable\n# endpoint address which can be used by WireGuard\ndiscover_endpoints: true\n";function a(e){let t={...e};t.language||(t.language="yaml"),t.title="/etc/cunicu.yaml";let n=o;if(t.section){const e=n.split("\n");let i=[],s=[],o=!1;for(let n of e){let e=!1,a=!1,r=n.startsWith("#"),c=""===n.trim(),d=n.match(/^([a-zA-z]+):/);null!==d&&(e=d[1]==t.section,a=d[1]!=t.section),r&&(o=!1,i.push(n)),e&&(o=!0,s.push(...i),i=[]),a&&(o=!1),c&&(i=[]),o&&s.push(n)}""==s[s.length-1]&&(s=s.slice(0,-1)),n=s.join("\n"),t.title=`Section "${t.section}" of ${t.title}`}return i.createElement(s.Z,t,n)}},220:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>d,contentTitle:()=>r,default:()=>u,frontMatter:()=>a,metadata:()=>c,toc:()=>l});var i=n(7462),s=(n(7294),n(4137)),o=n(2810);const a={},r="Hooks",c={unversionedId:"features/hooks",id:"features/hooks",title:"Hooks",description:"The hooks feature allows the user to configure a list of hook functions which are triggered by certain events within the daemon.",source:"@site/docs/features/hooks.md",sourceDirName:"features",slug:"/features/hooks",permalink:"/docs/features/hooks",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/features/hooks.md",tags:[],version:"current",frontMatter:{},sidebar:"tutorialSidebar",previous:{title:"Endpoint Discovery",permalink:"/docs/features/epdisc"},next:{title:"Hosts-file Synchronization",permalink:"/docs/features/hsync"}},d={},l=[{value:"Configuration",id:"configuration",level:2}],f={toc:l},h="wrapper";function u(e){let{components:t,...n}=e;return(0,s.kt)(h,(0,i.Z)({},f,n,{components:t,mdxType:"MDXLayout"}),(0,s.kt)("h1",{id:"hooks"},"Hooks"),(0,s.kt)("p",null,"The hooks feature allows the user to configure a list of hook functions which are triggered by certain events within the daemon."),(0,s.kt)("h2",{id:"configuration"},"Configuration"),(0,s.kt)(o.Z,{section:"hooks",mdxType:"ExampleConfig"}))}u.isMDXComponent=!0}}]);