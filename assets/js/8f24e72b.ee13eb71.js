"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[5734],{4137:(e,n,t)=>{t.d(n,{Zo:()=>d,kt:()=>f});var r=t(7294);function a(e,n,t){return n in e?Object.defineProperty(e,n,{value:t,enumerable:!0,configurable:!0,writable:!0}):e[n]=t,e}function s(e,n){var t=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);n&&(r=r.filter((function(n){return Object.getOwnPropertyDescriptor(e,n).enumerable}))),t.push.apply(t,r)}return t}function o(e){for(var n=1;n<arguments.length;n++){var t=null!=arguments[n]?arguments[n]:{};n%2?s(Object(t),!0).forEach((function(n){a(e,n,t[n])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(t)):s(Object(t)).forEach((function(n){Object.defineProperty(e,n,Object.getOwnPropertyDescriptor(t,n))}))}return e}function i(e,n){if(null==e)return{};var t,r,a=function(e,n){if(null==e)return{};var t,r,a={},s=Object.keys(e);for(r=0;r<s.length;r++)t=s[r],n.indexOf(t)>=0||(a[t]=e[t]);return a}(e,n);if(Object.getOwnPropertySymbols){var s=Object.getOwnPropertySymbols(e);for(r=0;r<s.length;r++)t=s[r],n.indexOf(t)>=0||Object.prototype.propertyIsEnumerable.call(e,t)&&(a[t]=e[t])}return a}var u=r.createContext({}),c=function(e){var n=r.useContext(u),t=n;return e&&(t="function"==typeof e?e(n):o(o({},n),e)),t},d=function(e){var n=c(e.components);return r.createElement(u.Provider,{value:n},e.children)},l="mdxType",p={inlineCode:"code",wrapper:function(e){var n=e.children;return r.createElement(r.Fragment,{},n)}},m=r.forwardRef((function(e,n){var t=e.components,a=e.mdxType,s=e.originalType,u=e.parentName,d=i(e,["components","mdxType","originalType","parentName"]),l=c(t),m=a,f=l["".concat(u,".").concat(m)]||l[m]||p[m]||s;return t?r.createElement(f,o(o({ref:n},d),{},{components:t})):r.createElement(f,o({ref:n},d))}));function f(e,n){var t=arguments,a=n&&n.mdxType;if("string"==typeof e||a){var s=t.length,o=new Array(s);o[0]=m;var i={};for(var u in n)hasOwnProperty.call(n,u)&&(i[u]=n[u]);i.originalType=e,i[l]="string"==typeof e?e:a,o[1]=i;for(var c=2;c<s;c++)o[c]=t[c];return r.createElement.apply(null,o)}return r.createElement.apply(null,t)}m.displayName="MDXCreateElement"},2449:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>u,contentTitle:()=>o,default:()=>p,frontMatter:()=>s,metadata:()=>i,toc:()=>c});var r=t(7462),a=(t(7294),t(4137));const s={title:"cunicu addresses",sidebar_label:"addresses",sidebar_class_name:"command-name",slug:"/usage/man/addresses",hide_title:!0,keywords:["manpage"]},o=void 0,i={unversionedId:"usage/md/cunicu_addresses",id:"usage/md/cunicu_addresses",title:"cunicu addresses",description:"cunicu addresses",source:"@site/docs/usage/md/cunicu_addresses.md",sourceDirName:"usage/md",slug:"/usage/man/addresses",permalink:"/docs/usage/man/addresses",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/usage/md/cunicu_addresses.md",tags:[],version:"current",frontMatter:{title:"cunicu addresses",sidebar_label:"addresses",sidebar_class_name:"command-name",slug:"/usage/man/addresses",hide_title:!0,keywords:["manpage"]},sidebar:"tutorialSidebar",previous:{title:"cunicu",permalink:"/docs/usage/man/"},next:{title:"completion",permalink:"/docs/usage/man/completion"}},u={},c=[{value:"cunicu addresses",id:"cunicu-addresses",level:2},{value:"Synopsis",id:"synopsis",level:3},{value:"Examples",id:"examples",level:3},{value:"Options",id:"options",level:3},{value:"Options inherited from parent commands",id:"options-inherited-from-parent-commands",level:3},{value:"SEE ALSO",id:"see-also",level:3}],d={toc:c},l="wrapper";function p(e){let{components:n,...t}=e;return(0,a.kt)(l,(0,r.Z)({},d,t,{components:n,mdxType:"MDXLayout"}),(0,a.kt)("h2",{id:"cunicu-addresses"},"cunicu addresses"),(0,a.kt)("p",null,"Derive IPv4 and IPv6 addresses from a WireGuard X25519 public key"),(0,a.kt)("h3",{id:"synopsis"},"Synopsis"),(0,a.kt)("p",null,"cun\u012bcu auto-configuration feature derives and assigns IPv4 and IPv6 addresses based on the public key of the WireGuard interface.\nThis sub-command accepts a WireGuard public key on the standard input and prints out the calculated IP addresses on the standard output."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"cunicu addresses [flags]\n")),(0,a.kt)("h3",{id:"examples"},"Examples"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"$ wg genkey | wg pubkey | cunicu addresses\nfc2f:9a4d:777f:7a97:8197:4a5d:1d1b:ed79\n10.237.119.127\n")),(0,a.kt)("h3",{id:"options"},"Options"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"  -h, --help   help for addresses\n  -m, --mask   Print CIDR mask\n")),(0,a.kt)("h3",{id:"options-inherited-from-parent-commands"},"Options inherited from parent commands"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},'  -q, --color string       Enable colorization of output (one of: auto, always, never) (default "auto")\n  -l, --log-file string    path of a file to write logs to\n  -d, --log-level string   log level filter rule (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")\n')),(0,a.kt)("h3",{id:"see-also"},"SEE ALSO"),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/docs/usage/man/"},"cunicu"),"\t - cun\u012bcu is a user-space daemon managing WireGuard\xae interfaces to establish peer-to-peer connections in harsh network environments.")))}p.isMDXComponent=!0}}]);