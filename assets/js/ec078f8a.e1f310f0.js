"use strict";(self.webpackChunkwice=self.webpackChunkwice||[]).push([[6692],{3905:(e,t,r)=>{r.d(t,{Zo:()=>p,kt:()=>f});var n=r(7294);function o(e,t,r){return t in e?Object.defineProperty(e,t,{value:r,enumerable:!0,configurable:!0,writable:!0}):e[t]=r,e}function i(e,t){var r=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);t&&(n=n.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),r.push.apply(r,n)}return r}function a(e){for(var t=1;t<arguments.length;t++){var r=null!=arguments[t]?arguments[t]:{};t%2?i(Object(r),!0).forEach((function(t){o(e,t,r[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(r)):i(Object(r)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(r,t))}))}return e}function l(e,t){if(null==e)return{};var r,n,o=function(e,t){if(null==e)return{};var r,n,o={},i=Object.keys(e);for(n=0;n<i.length;n++)r=i[n],t.indexOf(r)>=0||(o[r]=e[r]);return o}(e,t);if(Object.getOwnPropertySymbols){var i=Object.getOwnPropertySymbols(e);for(n=0;n<i.length;n++)r=i[n],t.indexOf(r)>=0||Object.prototype.propertyIsEnumerable.call(e,r)&&(o[r]=e[r])}return o}var s=n.createContext({}),c=function(e){var t=n.useContext(s),r=t;return e&&(r="function"==typeof e?e(t):a(a({},t),e)),r},p=function(e){var t=c(e.components);return n.createElement(s.Provider,{value:t},e.children)},u={inlineCode:"code",wrapper:function(e){var t=e.children;return n.createElement(n.Fragment,{},t)}},d=n.forwardRef((function(e,t){var r=e.components,o=e.mdxType,i=e.originalType,s=e.parentName,p=l(e,["components","mdxType","originalType","parentName"]),d=c(r),f=o,h=d["".concat(s,".").concat(f)]||d[f]||u[f]||i;return r?n.createElement(h,a(a({ref:t},p),{},{components:r})):n.createElement(h,a({ref:t},p))}));function f(e,t){var r=arguments,o=t&&t.mdxType;if("string"==typeof e||o){var i=r.length,a=new Array(i);a[0]=d;var l={};for(var s in t)hasOwnProperty.call(t,s)&&(l[s]=t[s]);l.originalType=e,l.mdxType="string"==typeof e?e:o,a[1]=l;for(var c=2;c<i;c++)a[c]=r[c];return n.createElement.apply(null,a)}return n.createElement.apply(null,r)}d.displayName="MDXCreateElement"},3850:(e,t,r)=>{r.r(t),r.d(t,{assets:()=>s,contentTitle:()=>a,default:()=>u,frontMatter:()=>i,metadata:()=>l,toc:()=>c});var n=r(7462),o=(r(7294),r(3905));const i={},a="Proxying",l={unversionedId:"development/proxying",id:"development/proxying",title:"Proxying",description:"cun\u012bcu implements multiple ways of running an ICE agent alongside WireGuard on the same UDP ports.",source:"@site/docs/development/proxying.md",sourceDirName:"development",slug:"/development/proxying",permalink:"/docs/development/proxying",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/development/proxying.md",tags:[],version:"current",frontMatter:{},sidebar:"tutorialSidebar",previous:{title:"Development",permalink:"/docs/development/"},next:{title:"Session Signaling",permalink:"/docs/development/signaling"}},s={},c=[{value:"Kernel WireGuard module",id:"kernel-wireguard-module",level:2},{value:"User-space",id:"user-space",level:3},{value:"RAW Sockets + BPF filter (Kernel)",id:"raw-sockets--bpf-filter-kernel",level:3},{value:"NFtables port-redirection (Kernel)",id:"nftables-port-redirection-kernel",level:3},{value:"IPTables port-redirection",id:"iptables-port-redirection",level:2},{value:"User-space WireGuard implementation",id:"user-space-wireguard-implementation",level:2},{value:"User-space Proxy",id:"user-space-proxy",level:3},{value:"In-process socket",id:"in-process-socket",level:3},{value:"Flowchart",id:"flowchart",level:2}],p={toc:c};function u(e){let{components:t,...i}=e;return(0,o.kt)("wrapper",(0,n.Z)({},p,i,{components:t,mdxType:"MDXLayout"}),(0,o.kt)("h1",{id:"proxying"},"Proxying"),(0,o.kt)("p",null,"cun\u012bcu implements multiple ways of running an ICE agent alongside WireGuard on the same UDP ports."),(0,o.kt)("h2",{id:"kernel-wireguard-module"},"Kernel WireGuard module"),(0,o.kt)("h3",{id:"user-space"},"User-space"),(0,o.kt)("p",null,"For each WG peer a new local UDP socket is opened.\ncun\u012bcu will update the endpoint address of the peer to this the local address of the new sockets."),(0,o.kt)("p",null,"WireGuard traffic is proxied by cun\u012bcu between the local UDP and the ICE socket."),(0,o.kt)("h3",{id:"raw-sockets--bpf-filter-kernel"},"RAW Sockets + BPF filter (Kernel)"),(0,o.kt)("p",null,"We allocate a single ",(0,o.kt)("a",{parentName:"p",href:"https://squidarth.com/networking/systems/rc/2018/05/28/using-raw-sockets.html"},"Linux RAW socket")," and assign a ",(0,o.kt)("a",{parentName:"p",href:"https://riyazali.net/posts/berkeley-packet-filter-in-golang/"},"eBPF")," filter to this socket which will only match STUN traffic to a specific UDP port.\nUDP headers are parsed/produced by cun\u012bcu.\ncun\u012bcu uses a UDPMux to mux all peers ICE Agents over this single RAW socket. "),(0,o.kt)("h3",{id:"nftables-port-redirection-kernel"},"NFtables port-redirection (Kernel)"),(0,o.kt)("p",null,"Two ",(0,o.kt)("a",{parentName:"p",href:"https://www.netfilter.org/projects/nftables/manpage.html"},"Nftables")," (nft) rules are added to filter input & output chains respectively.\nThe input rule will match all non-STUN traffic directed at the local port of the ICE candidate and rewrites the UDP destination port to the local listen port of the WireGuard interface.\nThe output rule will mach all traffic originating from the listen port of the WG interface and directed to the port of the remote candidate and rewrites the source port to the port of the local ICE candidate.  "),(0,o.kt)("p",null,"WireGuard traffic passes only through the Netfilter chains and remains inside the kernel.\nOnly STUN binding requests are passed to cun\u012bcu."),(0,o.kt)("pre",null,(0,o.kt)("code",{parentName:"pre",className:"language-bash"},"$ sudo nft list ruleset\ntable inet cunicu {\n    chain ingress {\n        type filter hook input priority raw; policy accept;\n        udp dport 37281 @th,96,32 != 554869826 notrack udp dport set 1001\n    }\n\n    chain egress {\n        type filter hook output priority raw; policy accept;\n        udp sport 1001 udp dport 38767 notrack udp sport set 37281\n    }\n}\n")),(0,o.kt)("h2",{id:"iptables-port-redirection"},"IPTables port-redirection"),(0,o.kt)("p",null,"Similar to NFTables port-natting by using the legacy IPTables API."),(0,o.kt)("h2",{id:"user-space-wireguard-implementation"},"User-space WireGuard implementation"),(0,o.kt)("h3",{id:"user-space-proxy"},"User-space Proxy"),(0,o.kt)("p",null,"Just like for the Kernel WireGuard module, a dedicated UDP socket for each WG peer is created.\ncun\u012bcu will update the endpoint address of the peer to this the local address of the new sockets."),(0,o.kt)("p",null,"WireGuard traffic is proxied by cun\u012bcu between the local UDP and the ICE socket."),(0,o.kt)("h3",{id:"in-process-socket"},"In-process socket"),(0,o.kt)("p",null,"cun\u012bcu implements wireguard-go's ",(0,o.kt)("inlineCode",{parentName:"p"},"conn.Bind")," interface to handle WireGuard's network IO."),(0,o.kt)("p",null,"WireGuard traffic is passed directly between ",(0,o.kt)("inlineCode",{parentName:"p"},"conn.Bind")," and Pion's ",(0,o.kt)("inlineCode",{parentName:"p"},"ice.Conn"),".\nNo round-trip through the kernel stack is required."),(0,o.kt)("admonition",{type:"note"},(0,o.kt)("p",{parentName:"admonition"},"This variant only works for the compiled-in version of wireguard-go in cun\u012bcu.")),(0,o.kt)("h2",{id:"flowchart"},"Flowchart"),(0,o.kt)("p",null,(0,o.kt)("img",{src:r(7169).Z,width:"874",height:"914"})))}u.isMDXComponent=!0},7169:(e,t,r)=>{r.d(t,{Z:()=>n});const n=r.p+"assets/images/proxy-561858b0dc7f68600875c73c14ebef67.svg"}}]);