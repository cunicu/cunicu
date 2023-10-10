"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[6929],{3905:(e,n,t)=>{t.d(n,{Zo:()=>l,kt:()=>f});var a=t(67294);function r(e,n,t){return n in e?Object.defineProperty(e,n,{value:t,enumerable:!0,configurable:!0,writable:!0}):e[n]=t,e}function i(e,n){var t=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);n&&(a=a.filter((function(n){return Object.getOwnPropertyDescriptor(e,n).enumerable}))),t.push.apply(t,a)}return t}function o(e){for(var n=1;n<arguments.length;n++){var t=null!=arguments[n]?arguments[n]:{};n%2?i(Object(t),!0).forEach((function(n){r(e,n,t[n])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(t)):i(Object(t)).forEach((function(n){Object.defineProperty(e,n,Object.getOwnPropertyDescriptor(t,n))}))}return e}function u(e,n){if(null==e)return{};var t,a,r=function(e,n){if(null==e)return{};var t,a,r={},i=Object.keys(e);for(a=0;a<i.length;a++)t=i[a],n.indexOf(t)>=0||(r[t]=e[t]);return r}(e,n);if(Object.getOwnPropertySymbols){var i=Object.getOwnPropertySymbols(e);for(a=0;a<i.length;a++)t=i[a],n.indexOf(t)>=0||Object.prototype.propertyIsEnumerable.call(e,t)&&(r[t]=e[t])}return r}var c=a.createContext({}),s=function(e){var n=a.useContext(c),t=n;return e&&(t="function"==typeof e?e(n):o(o({},n),e)),t},l=function(e){var n=s(e.components);return a.createElement(c.Provider,{value:n},e.children)},p="mdxType",m={inlineCode:"code",wrapper:function(e){var n=e.children;return a.createElement(a.Fragment,{},n)}},d=a.forwardRef((function(e,n){var t=e.components,r=e.mdxType,i=e.originalType,c=e.parentName,l=u(e,["components","mdxType","originalType","parentName"]),p=s(t),d=r,f=p["".concat(c,".").concat(d)]||p[d]||m[d]||i;return t?a.createElement(f,o(o({ref:n},l),{},{components:t})):a.createElement(f,o({ref:n},l))}));function f(e,n){var t=arguments,r=n&&n.mdxType;if("string"==typeof e||r){var i=t.length,o=new Array(i);o[0]=d;var u={};for(var c in n)hasOwnProperty.call(n,c)&&(u[c]=n[c]);u.originalType=e,u[p]="string"==typeof e?e:r,o[1]=u;for(var s=2;s<i;s++)o[s]=t[s];return a.createElement.apply(null,o)}return a.createElement.apply(null,t)}d.displayName="MDXCreateElement"},51786:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>c,contentTitle:()=>o,default:()=>m,frontMatter:()=>i,metadata:()=>u,toc:()=>s});var a=t(87462),r=(t(67294),t(3905));const i={title:"cunicu",sidebar_class_name:"command-name",slug:"/usage/man/",hide_title:!0,keywords:["manpage"]},o=void 0,u={unversionedId:"usage/md/cunicu",id:"usage/md/cunicu",title:"cunicu",description:"cunicu",source:"@site/docs/usage/md/cunicu.md",sourceDirName:"usage/md",slug:"/usage/man/",permalink:"/docs/usage/man/",draft:!1,editUrl:"https://github.com/cunicu/cunicu/edit/main/docs/usage/md/cunicu.md",tags:[],version:"current",frontMatter:{title:"cunicu",sidebar_class_name:"command-name",slug:"/usage/man/",hide_title:!0,keywords:["manpage"]},sidebar:"tutorialSidebar",previous:{title:"Usage",permalink:"/docs/usage/"},next:{title:"addresses",permalink:"/docs/usage/man/addresses"}},c={},s=[{value:"cunicu",id:"cunicu",level:2},{value:"Synopsis",id:"synopsis",level:3},{value:"Options",id:"options",level:3},{value:"SEE ALSO",id:"see-also",level:3}],l={toc:s},p="wrapper";function m(e){let{components:n,...t}=e;return(0,r.kt)(p,(0,a.Z)({},l,t,{components:n,mdxType:"MDXLayout"}),(0,r.kt)("h2",{id:"cunicu"},"cunicu"),(0,r.kt)("p",null,"cun\u012bcu is a user-space daemon managing WireGuard\xae interfaces to establish peer-to-peer connections in harsh network environments."),(0,r.kt)("h3",{id:"synopsis"},"Synopsis"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},'   (\\(\\       \u259f\u2580\u2580\u2599 \u2588  \u2588 \u2588\u2580\u2580\u2599 \u2580\u2580\u2580 \u259f\u2580\u2580\u2599 \u2588  \u2599     \n   (-,-)      \u2588    \u2588  \u2588 \u2588  \u2588 \u2580\u2588  \u2588    \u2588  \u2588     (\\_/)\n o_(")(")     \u259c\u2584\u2584\u259b \u259c\u2584\u2584\u259b \u2588  \u2588 \u2584\u2588\u2584 \u259c\u2584\u2584\u259b \u259c\u2584\u2584\u259b     (\u2022_\u2022)\n              zero-conf \u2022 p2p \u2022 mesh \u2022 vpn     /> \u2764\ufe0f  WireGuard\u2122\n')),(0,r.kt)("p",null,"cun\u012bcu is a user-space daemon managing WireGuard\xae interfaces to\nestablish peer-to-peer connections in harsh network environments."),(0,r.kt)("p",null,"It relies on the awesome pion/ice package for the interactive\nconnectivity establishment as well as bundles the Go user-space\nimplementation of WireGuard in a single binary for environments\nin which WireGuard kernel support has not landed yet."),(0,r.kt)("h3",{id:"options"},"Options"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},'  -q, --color string            Enable colorization of output (one of: auto, always, never) (default "auto")\n  -l, --log-file stringArray    path of a file to write logs to\n  -d, --log-level stringArray   log level filter rule (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")\n  -h, --help                    help for cunicu\n')),(0,r.kt)("h3",{id:"see-also"},"SEE ALSO"),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/addresses"},"cunicu addresses"),"\t - Derive IPv4 and IPv6 addresses from a WireGuard X25519 public key"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/completion"},"cunicu completion"),"\t - Generate the autocompletion script for the specified shell"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/config"},"cunicu config"),"\t - Manage configuration of a running cun\u012bcu daemon."),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/daemon"},"cunicu daemon"),"\t - Start the main daemon"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/invite"},"cunicu invite"),"\t - Add a new peer to the local daemon configuration and return the required configuration for this new peer"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/monitor"},"cunicu monitor"),"\t - Monitor the cun\u012bcu daemon for events"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/relay"},"cunicu relay"),"\t - Start relay API server"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/restart"},"cunicu restart"),"\t - Restart the cun\u012bcu daemon"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/selfupdate"},"cunicu selfupdate"),"\t - Update the cun\u012bcu binary"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/signal"},"cunicu signal"),"\t - Start gRPC signaling server"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/status"},"cunicu status"),"\t - Show current status of the cun\u012bcu daemon, its interfaces and peers"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/stop"},"cunicu stop"),"\t - Shutdown the cun\u012bcu daemon"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/sync"},"cunicu sync"),"\t - Synchronize cun\u012bcu daemon state"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/version"},"cunicu version"),"\t - Show version of the cun\u012bcu binary and optionally also a running daemon"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"/docs/usage/man/wg"},"cunicu wg"),"\t - WireGuard commands")))}m.isMDXComponent=!0}}]);