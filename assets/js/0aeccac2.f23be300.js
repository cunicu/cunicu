"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[414],{3905:(e,t,r)=>{r.d(t,{Zo:()=>c,kt:()=>d});var a=r(67294);function i(e,t,r){return t in e?Object.defineProperty(e,t,{value:r,enumerable:!0,configurable:!0,writable:!0}):e[t]=r,e}function n(e,t){var r=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);t&&(a=a.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),r.push.apply(r,a)}return r}function l(e){for(var t=1;t<arguments.length;t++){var r=null!=arguments[t]?arguments[t]:{};t%2?n(Object(r),!0).forEach((function(t){i(e,t,r[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(r)):n(Object(r)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(r,t))}))}return e}function o(e,t){if(null==e)return{};var r,a,i=function(e,t){if(null==e)return{};var r,a,i={},n=Object.keys(e);for(a=0;a<n.length;a++)r=n[a],t.indexOf(r)>=0||(i[r]=e[r]);return i}(e,t);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);for(a=0;a<n.length;a++)r=n[a],t.indexOf(r)>=0||Object.prototype.propertyIsEnumerable.call(e,r)&&(i[r]=e[r])}return i}var s=a.createContext({}),p=function(e){var t=a.useContext(s),r=t;return e&&(r="function"==typeof e?e(t):l(l({},t),e)),r},c=function(e){var t=p(e.components);return a.createElement(s.Provider,{value:t},e.children)},m="mdxType",u={inlineCode:"code",wrapper:function(e){var t=e.children;return a.createElement(a.Fragment,{},t)}},f=a.forwardRef((function(e,t){var r=e.components,i=e.mdxType,n=e.originalType,s=e.parentName,c=o(e,["components","mdxType","originalType","parentName"]),m=p(r),f=i,d=m["".concat(s,".").concat(f)]||m[f]||u[f]||n;return r?a.createElement(d,l(l({ref:t},c),{},{components:r})):a.createElement(d,l({ref:t},c))}));function d(e,t){var r=arguments,i=t&&t.mdxType;if("string"==typeof e||i){var n=r.length,l=new Array(n);l[0]=f;var o={};for(var s in t)hasOwnProperty.call(t,s)&&(o[s]=t[s]);o.originalType=e,o[m]="string"==typeof e?e:i,l[1]=o;for(var p=2;p<n;p++)l[p]=r[p];return a.createElement.apply(null,l)}return a.createElement.apply(null,r)}f.displayName="MDXCreateElement"},55449:(e,t,r)=>{r.r(t),r.d(t,{assets:()=>s,contentTitle:()=>l,default:()=>u,frontMatter:()=>n,metadata:()=>o,toc:()=>p});var a=r(87462),i=(r(67294),r(3905));const n={sidebar_position:20},l="Design",o={unversionedId:"design",id:"design",title:"Design",description:"Architecture",source:"@site/docs/design.md",sourceDirName:".",slug:"/design",permalink:"/docs/design",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/design.md",tags:[],version:"current",sidebarPosition:20,frontMatter:{sidebar_position:20},sidebar:"tutorialSidebar",previous:{title:"JSON Schema",permalink:"/docs/config/schema"},next:{title:"Comparison",permalink:"/docs/comparison"}},s={},p=[{value:"Architecture",id:"architecture",level:2},{value:"Objectives",id:"objectives",level:2},{value:"Related RFCs",id:"related-rfcs",level:2}],c={toc:p},m="wrapper";function u(e){let{components:t,...n}=e;return(0,i.kt)(m,(0,a.Z)({},c,n,{components:t,mdxType:"MDXLayout"}),(0,i.kt)("h1",{id:"design"},"Design"),(0,i.kt)("h2",{id:"architecture"},"Architecture"),(0,i.kt)("p",null,(0,i.kt)("img",{src:r(62215).Z,width:"901",height:"629"})),(0,i.kt)("h2",{id:"objectives"},"Objectives"),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Encrypt all signaling messages")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Plug-able signaling backends:"),(0,i.kt)("ul",{parentName:"li"},(0,i.kt)("li",{parentName:"ul"},"GRPC"),(0,i.kt)("li",{parentName:"ul"},"Kubernetes API-server"),(0,i.kt)("li",{parentName:"ul"},"WebSocket"))),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Support ",(0,i.kt)("a",{parentName:"p",href:"https://datatracker.ietf.org/doc/html/rfc8838"},"Trickle ICE"))),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Support ",(0,i.kt)("a",{parentName:"p",href:"https://datatracker.ietf.org/doc/html/rfc8445#section-2.4"},"ICE restart"))),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Support ",(0,i.kt)("a",{parentName:"p",href:"https://datatracker.ietf.org/doc/html/rfc6544"},"ICE-TCP"))),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Encrypt exchanged ICE offers with WireGuard keys")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Seamless switch between ICE candidates and relays")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Zero configuration"),(0,i.kt)("ul",{parentName:"li"},(0,i.kt)("li",{parentName:"ul"},"Alleviate users of exchanging endpoint IPs & ports"))),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Enables direct communication of WireGuard peers behind NAT / UDP-blocking firewalls")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Single-binary, zero dependency installation"),(0,i.kt)("ul",{parentName:"li"},(0,i.kt)("li",{parentName:"ul"},"Bundled ICE agent & ",(0,i.kt)("a",{parentName:"li",href:"https://git.zx2c4.com/wireguard-go"},"WireGuard user-space daemon")),(0,i.kt)("li",{parentName:"ul"},"Portability"))),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Support for user and kernel-space WireGuard implementations")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Zero performance impact"),(0,i.kt)("ul",{parentName:"li"},(0,i.kt)("li",{parentName:"ul"},"Kernel-side filtering / redirection of WireGuard traffic"),(0,i.kt)("li",{parentName:"ul"},"Fallback to user-space proxying only if no Kernel features are available "))),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Minimized attack surface"),(0,i.kt)("ul",{parentName:"li"},(0,i.kt)("li",{parentName:"ul"},"Drop privileges after initial configuration"))),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Compatible with existing WireGuard configuration utilities like:"),(0,i.kt)("ul",{parentName:"li"},(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://github.com/max-moser/network-manager-wireguard"},"NetworkManager")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://www.freedesktop.org/software/systemd/man/systemd.netdev.html#%5BWireGuard%5D%20Section%20Options"},"systemd-networkd")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://manpages.debian.org/unstable/wireguard-tools/wg-quick.8.en.html"},"wg-quick")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://kilo.squat.ai"},"Kilo")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://seashell.github.io/drago/"},"drago")))),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("p",{parentName:"li"},"Monitoring for new WireGuard interfaces and peers"),(0,i.kt)("ul",{parentName:"li"},(0,i.kt)("li",{parentName:"ul"},"Inotify for new UAPI sockets in /var/run/wireguard"),(0,i.kt)("li",{parentName:"ul"},"Netlink subscription for link updates (patch is pending)")))),(0,i.kt)("h2",{id:"related-rfcs"},"Related RFCs"),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://datatracker.ietf.org/doc/html/rfc6544"},"RFC6544")," TCP Candidates with Interactive Connectivity Establishment (ICE)"),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://datatracker.ietf.org/doc/html/rfc8838"},"RFC8838")," Trickle ICE: Incremental Provisioning of Candidates for the Interactive Connectivity Establishment (ICE) Protocol"),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://datatracker.ietf.org/doc/html/rfc8445"},"RFC8445")," Interactive Connectivity Establishment (ICE): A Protocol for Network Address Translator (NAT) Traversal"),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://datatracker.ietf.org/doc/html/rfc8863"},"RFC8863")," Interactive Connectivity Establishment Patiently Awaiting Connectivity (ICE PAC)"),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://datatracker.ietf.org/doc/html/rfc8839"},"RFC8839")," Session Description Protocol (SDP) Offer/Answer Procedures for Interactive Connectivity Establishment (ICE)"),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://datatracker.ietf.org/doc/html/rfc6062"},"RFC6062")," Traversal Using Relays around NAT (TURN) Extensions for TCP Allocations"),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://datatracker.ietf.org/doc/html/rfc8656"},"RFC8656")," Traversal Using Relays around NAT (TURN): Relay Extensions to Session Traversal Utilities for NAT (STUN)"),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://datatracker.ietf.org/doc/html/rfc8489"},"RFC8489")," Session Traversal Utilities for NAT (STUN)"),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://datatracker.ietf.org/doc/html/rfc8866"},"RFC8866")," SDP: Session Description Protocol"),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://datatracker.ietf.org/doc/html/rfc3264"},"RFC3264")," An Offer/Answer Model with the Session Description Protocol (SDP)"),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://datatracker.ietf.org/doc/html/rfc7064"},"RFC7064")," URI Scheme for the Session Traversal Utilities for NAT (STUN) Protocol"),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://datatracker.ietf.org/doc/html/rfc7065"},"RFC7065")," Traversal Using Relays around NAT (TURN) Uniform Resource Identifiers")))}u.isMDXComponent=!0},62215:(e,t,r)=>{r.d(t,{Z:()=>a});const a=r.p+"assets/images/architecture-698f935e44bbe4e44537cc165a669ff3.svg"}}]);