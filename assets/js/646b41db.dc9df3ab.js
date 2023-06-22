"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[209],{4137:(e,t,r)=>{r.d(t,{Zo:()=>c,kt:()=>f});var n=r(7294);function a(e,t,r){return t in e?Object.defineProperty(e,t,{value:r,enumerable:!0,configurable:!0,writable:!0}):e[t]=r,e}function l(e,t){var r=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);t&&(n=n.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),r.push.apply(r,n)}return r}function i(e){for(var t=1;t<arguments.length;t++){var r=null!=arguments[t]?arguments[t]:{};t%2?l(Object(r),!0).forEach((function(t){a(e,t,r[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(r)):l(Object(r)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(r,t))}))}return e}function o(e,t){if(null==e)return{};var r,n,a=function(e,t){if(null==e)return{};var r,n,a={},l=Object.keys(e);for(n=0;n<l.length;n++)r=l[n],t.indexOf(r)>=0||(a[r]=e[r]);return a}(e,t);if(Object.getOwnPropertySymbols){var l=Object.getOwnPropertySymbols(e);for(n=0;n<l.length;n++)r=l[n],t.indexOf(r)>=0||Object.prototype.propertyIsEnumerable.call(e,r)&&(a[r]=e[r])}return a}var s=n.createContext({}),u=function(e){var t=n.useContext(s),r=t;return e&&(r="function"==typeof e?e(t):i(i({},t),e)),r},c=function(e){var t=u(e.components);return n.createElement(s.Provider,{value:t},e.children)},p="mdxType",m={inlineCode:"code",wrapper:function(e){var t=e.children;return n.createElement(n.Fragment,{},t)}},d=n.forwardRef((function(e,t){var r=e.components,a=e.mdxType,l=e.originalType,s=e.parentName,c=o(e,["components","mdxType","originalType","parentName"]),p=u(r),d=a,f=p["".concat(s,".").concat(d)]||p[d]||m[d]||l;return r?n.createElement(f,i(i({ref:t},c),{},{components:r})):n.createElement(f,i({ref:t},c))}));function f(e,t){var r=arguments,a=t&&t.mdxType;if("string"==typeof e||a){var l=r.length,i=new Array(l);i[0]=d;var o={};for(var s in t)hasOwnProperty.call(t,s)&&(o[s]=t[s]);o.originalType=e,o[p]="string"==typeof e?e:a,i[1]=o;for(var u=2;u<l;u++)i[u]=r[u];return n.createElement.apply(null,i)}return n.createElement.apply(null,r)}d.displayName="MDXCreateElement"},8925:(e,t,r)=>{r.r(t),r.d(t,{assets:()=>s,contentTitle:()=>i,default:()=>m,frontMatter:()=>l,metadata:()=>o,toc:()=>u});var n=r(7462),a=(r(7294),r(4137));const l={title:"cunicu relay",sidebar_label:"relay",sidebar_class_name:"command-name",slug:"/usage/man/relay",hide_title:!0,keywords:["manpage"]},i=void 0,o={unversionedId:"usage/md/cunicu_relay",id:"usage/md/cunicu_relay",title:"cunicu relay",description:"cunicu relay",source:"@site/docs/usage/md/cunicu_relay.md",sourceDirName:"usage/md",slug:"/usage/man/relay",permalink:"/docs/usage/man/relay",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/usage/md/cunicu_relay.md",tags:[],version:"current",frontMatter:{title:"cunicu relay",sidebar_label:"relay",sidebar_class_name:"command-name",slug:"/usage/man/relay",hide_title:!0,keywords:["manpage"]},sidebar:"tutorialSidebar",previous:{title:"monitor",permalink:"/docs/usage/man/monitor"},next:{title:"restart",permalink:"/docs/usage/man/restart"}},s={},u=[{value:"cunicu relay",id:"cunicu-relay",level:2},{value:"Synopsis",id:"synopsis",level:3},{value:"Examples",id:"examples",level:3},{value:"Options",id:"options",level:3},{value:"Options inherited from parent commands",id:"options-inherited-from-parent-commands",level:3},{value:"SEE ALSO",id:"see-also",level:3}],c={toc:u},p="wrapper";function m(e){let{components:t,...r}=e;return(0,a.kt)(p,(0,n.Z)({},c,r,{components:t,mdxType:"MDXLayout"}),(0,a.kt)("h2",{id:"cunicu-relay"},"cunicu relay"),(0,a.kt)("p",null,"Start relay API server"),(0,a.kt)("h3",{id:"synopsis"},"Synopsis"),(0,a.kt)("p",null,"This command starts a gRPC server providing cunicu agents with a list of available STUN and TURN servers."),(0,a.kt)("p",null,(0,a.kt)("strong",{parentName:"p"},"Note:")," Currently this command does not run a TURN server itself. But relies on an external server like Coturn."),(0,a.kt)("p",null,"With this feature you can distribute a list of available STUN/TURN servers easily to a fleet of agents.\nIt also allows to issue short-lived HMAC-SHA1 credentials based the proposed TURN REST API and thereby static long term credentials."),(0,a.kt)("p",null,"The command expects a list of STUN or TURN URLs according to RFC7065/RFC7064 with a few extensions:"),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("p",{parentName:"li"},"A secret for the TURN REST API can be provided by the 'secret' query parameter"),(0,a.kt)("ul",{parentName:"li"},(0,a.kt)("li",{parentName:"ul"},"Example: turn:server.com?secret=rest-api-secret"))),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("p",{parentName:"li"},"A time-to-live to the TURN REST API secrets can be provided by the 'ttl' query parameter"),(0,a.kt)("ul",{parentName:"li"},(0,a.kt)("li",{parentName:"ul"},"Example: turn:server.com?ttl=1h"))),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("p",{parentName:"li"},"Static TURN credentials can be provided by the URIs user info"),(0,a.kt)("ul",{parentName:"li"},(0,a.kt)("li",{parentName:"ul"},"Example: turn:user1:",(0,a.kt)("a",{parentName:"li",href:"mailto:pass1@server.com"},"pass1@server.com"))))),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"cunicu relay URL... [flags]\n")),(0,a.kt)("h3",{id:"examples"},"Examples"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"relay turn:server.com?secret=rest-api-secret&ttl=1h\n")),(0,a.kt)("h3",{id:"options"},"Options"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},'  -h, --help            help for relay\n  -L, --listen string   listen address (default ":8080")\n  -S, --secure          listen with TLS\n')),(0,a.kt)("h3",{id:"options-inherited-from-parent-commands"},"Options inherited from parent commands"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},'  -q, --color string       Enable colorization of output (one of: auto, always, never) (default "auto")\n  -l, --log-file string    path of a file to write logs to\n  -d, --log-level string   log level filter rule (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")\n')),(0,a.kt)("h3",{id:"see-also"},"SEE ALSO"),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/docs/usage/man/"},"cunicu"),"\t - cun\u012bcu is a user-space daemon managing WireGuard\xae interfaces to establish peer-to-peer connections in harsh network environments.")))}m.isMDXComponent=!0}}]);