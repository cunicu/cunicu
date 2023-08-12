"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[1503],{3905:(e,t,n)=>{n.d(t,{Zo:()=>u,kt:()=>m});var r=n(67294);function i(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function a(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,r)}return n}function o(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?a(Object(n),!0).forEach((function(t){i(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):a(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function s(e,t){if(null==e)return{};var n,r,i=function(e,t){if(null==e)return{};var n,r,i={},a=Object.keys(e);for(r=0;r<a.length;r++)n=a[r],t.indexOf(n)>=0||(i[n]=e[n]);return i}(e,t);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);for(r=0;r<a.length;r++)n=a[r],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(i[n]=e[n])}return i}var l=r.createContext({}),c=function(e){var t=r.useContext(l),n=t;return e&&(n="function"==typeof e?e(t):o(o({},t),e)),n},u=function(e){var t=c(e.components);return r.createElement(l.Provider,{value:t},e.children)},p="mdxType",d={inlineCode:"code",wrapper:function(e){var t=e.children;return r.createElement(r.Fragment,{},t)}},f=r.forwardRef((function(e,t){var n=e.components,i=e.mdxType,a=e.originalType,l=e.parentName,u=s(e,["components","mdxType","originalType","parentName"]),p=c(n),f=i,m=p["".concat(l,".").concat(f)]||p[f]||d[f]||a;return n?r.createElement(m,o(o({ref:t},u),{},{components:n})):r.createElement(m,o({ref:t},u))}));function m(e,t){var n=arguments,i=t&&t.mdxType;if("string"==typeof e||i){var a=n.length,o=new Array(a);o[0]=f;var s={};for(var l in t)hasOwnProperty.call(t,l)&&(s[l]=t[l]);s.originalType=e,s[p]="string"==typeof e?e:i,o[1]=s;for(var c=2;c<a;c++)o[c]=n[c];return r.createElement.apply(null,o)}return r.createElement.apply(null,n)}f.displayName="MDXCreateElement"},58548:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>l,contentTitle:()=>o,default:()=>d,frontMatter:()=>a,metadata:()=>s,toc:()=>c});var r=n(87462),i=(n(67294),n(3905));const a={title:"cunicu wg show",sidebar_label:"wg show",sidebar_class_name:"command-name",slug:"/usage/man/wg/show",hide_title:!0,keywords:["manpage"]},o=void 0,s={unversionedId:"usage/md/cunicu_wg_show",id:"usage/md/cunicu_wg_show",title:"cunicu wg show",description:"cunicu wg show",source:"@site/docs/usage/md/cunicu_wg_show.md",sourceDirName:"usage/md",slug:"/usage/man/wg/show",permalink:"/docs/usage/man/wg/show",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/usage/md/cunicu_wg_show.md",tags:[],version:"current",frontMatter:{title:"cunicu wg show",sidebar_label:"wg show",sidebar_class_name:"command-name",slug:"/usage/man/wg/show",hide_title:!0,keywords:["manpage"]},sidebar:"tutorialSidebar",previous:{title:"wg pubkey",permalink:"/docs/usage/man/wg/pubkey"},next:{title:"wg showconf",permalink:"/docs/usage/man/wg/showconf"}},l={},c=[{value:"cunicu wg show",id:"cunicu-wg-show",level:2},{value:"Synopsis",id:"synopsis",level:3},{value:"Options",id:"options",level:3},{value:"Options inherited from parent commands",id:"options-inherited-from-parent-commands",level:3},{value:"SEE ALSO",id:"see-also",level:3}],u={toc:c},p="wrapper";function d(e){let{components:t,...n}=e;return(0,i.kt)(p,(0,r.Z)({},u,n,{components:t,mdxType:"MDXLayout"}),(0,i.kt)("h2",{id:"cunicu-wg-show"},"cunicu wg show"),(0,i.kt)("p",null,"Shows current WireGuard configuration and runtime information of specified ","[interface]","."),(0,i.kt)("h3",{id:"synopsis"},"Synopsis"),(0,i.kt)("p",null,"Shows current WireGuard configuration and runtime information of specified ","[interface]","."),(0,i.kt)("p",null,"If no ","[interface]"," is specified, ","[interface]"," defaults to 'all'."),(0,i.kt)("p",null,"If 'interfaces' is specified, prints a list of all WireGuard interfaces, one per line, and quits."),(0,i.kt)("p",null,"If no options are given after the interface specification, then prints a list of all attributes in a visually pleasing way meant for the terminal.\nOtherwise, prints specified information grouped by newlines and tabs, meant to be used in scripts."),(0,i.kt)("p",null,"For this script-friendly display, if 'all' is specified, then the first field for all categories of information is the interface name."),(0,i.kt)("p",null,"If 'dump' is specified, then several lines are printed; the first contains in order separated by tab: private-key, public-key, listen-port, fwmark.\nSubsequent lines are printed for each peer and contain in order separated by tab: public-key, preshared-key, endpoint, allowed-ips, latest-handshake, transfer-rx, transfer-tx, persistent-keepalive."),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre"},"cunicu wg show { interface-name | all | interfaces } [{ public-key | private-key | listen-port | fwmark | peers | preshared-keys | endpoints | allowed-ips | latest-handshakes | transfer | persistent-keepalive | dump }] [flags]\n")),(0,i.kt)("h3",{id:"options"},"Options"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre"},'  -h, --help                help for show\n  -s, --rpc-socket string   Unix control and monitoring socket (default "/var/run/cunicu.sock")\n')),(0,i.kt)("h3",{id:"options-inherited-from-parent-commands"},"Options inherited from parent commands"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre"},'  -q, --color string            Enable colorization of output (one of: auto, always, never) (default "auto")\n  -l, --log-file stringArray    path of a file to write logs to\n  -d, --log-level stringArray   log level filter rule (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")\n')),(0,i.kt)("h3",{id:"see-also"},"SEE ALSO"),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"/docs/usage/man/wg"},"cunicu wg"),"\t - WireGuard commands")))}d.isMDXComponent=!0}}]);