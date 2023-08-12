"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[8671],{3905:(e,t,n)=>{n.d(t,{Zo:()=>c,kt:()=>g});var r=n(67294);function a(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function o(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,r)}return n}function i(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?o(Object(n),!0).forEach((function(t){a(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):o(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function u(e,t){if(null==e)return{};var n,r,a=function(e,t){if(null==e)return{};var n,r,a={},o=Object.keys(e);for(r=0;r<o.length;r++)n=o[r],t.indexOf(n)>=0||(a[n]=e[n]);return a}(e,t);if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(r=0;r<o.length;r++)n=o[r],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(a[n]=e[n])}return a}var l=r.createContext({}),p=function(e){var t=r.useContext(l),n=t;return e&&(n="function"==typeof e?e(t):i(i({},t),e)),n},c=function(e){var t=p(e.components);return r.createElement(l.Provider,{value:t},e.children)},s="mdxType",d={inlineCode:"code",wrapper:function(e){var t=e.children;return r.createElement(r.Fragment,{},t)}},m=r.forwardRef((function(e,t){var n=e.components,a=e.mdxType,o=e.originalType,l=e.parentName,c=u(e,["components","mdxType","originalType","parentName"]),s=p(n),m=a,g=s["".concat(l,".").concat(m)]||s[m]||d[m]||o;return n?r.createElement(g,i(i({ref:t},c),{},{components:n})):r.createElement(g,i({ref:t},c))}));function g(e,t){var n=arguments,a=t&&t.mdxType;if("string"==typeof e||a){var o=n.length,i=new Array(o);i[0]=m;var u={};for(var l in t)hasOwnProperty.call(t,l)&&(u[l]=t[l]);u.originalType=e,u[s]="string"==typeof e?e:a,i[1]=u;for(var p=2;p<o;p++)i[p]=n[p];return r.createElement.apply(null,i)}return r.createElement.apply(null,n)}m.displayName="MDXCreateElement"},52940:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>l,contentTitle:()=>i,default:()=>d,frontMatter:()=>o,metadata:()=>u,toc:()=>p});var r=n(87462),a=(n(67294),n(3905));const o={title:"cunicu wg pubkey",sidebar_label:"wg pubkey",sidebar_class_name:"command-name",slug:"/usage/man/wg/pubkey",hide_title:!0,keywords:["manpage"]},i=void 0,u={unversionedId:"usage/md/cunicu_wg_pubkey",id:"usage/md/cunicu_wg_pubkey",title:"cunicu wg pubkey",description:"cunicu wg pubkey",source:"@site/docs/usage/md/cunicu_wg_pubkey.md",sourceDirName:"usage/md",slug:"/usage/man/wg/pubkey",permalink:"/docs/usage/man/wg/pubkey",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/usage/md/cunicu_wg_pubkey.md",tags:[],version:"current",frontMatter:{title:"cunicu wg pubkey",sidebar_label:"wg pubkey",sidebar_class_name:"command-name",slug:"/usage/man/wg/pubkey",hide_title:!0,keywords:["manpage"]},sidebar:"tutorialSidebar",previous:{title:"wg genpsk",permalink:"/docs/usage/man/wg/genpsk"},next:{title:"wg show",permalink:"/docs/usage/man/wg/show"}},l={},p=[{value:"cunicu wg pubkey",id:"cunicu-wg-pubkey",level:2},{value:"Synopsis",id:"synopsis",level:3},{value:"Examples",id:"examples",level:3},{value:"Options",id:"options",level:3},{value:"Options inherited from parent commands",id:"options-inherited-from-parent-commands",level:3},{value:"SEE ALSO",id:"see-also",level:3}],c={toc:p},s="wrapper";function d(e){let{components:t,...n}=e;return(0,a.kt)(s,(0,r.Z)({},c,n,{components:t,mdxType:"MDXLayout"}),(0,a.kt)("h2",{id:"cunicu-wg-pubkey"},"cunicu wg pubkey"),(0,a.kt)("p",null,"Calculates a public key and prints it in base64 to standard output."),(0,a.kt)("h3",{id:"synopsis"},"Synopsis"),(0,a.kt)("p",null,"Calculates a public key and prints it in base64 to standard output from a corresponding private key (generated with genkey) given in base64 on standard input."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"cunicu wg pubkey [flags]\n")),(0,a.kt)("h3",{id:"examples"},"Examples"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"# A private key and a corresponding public key may be generated at once by calling:\n$ umask 077\n$ wg genkey | tee private.key | wg pubkey > public.key\n")),(0,a.kt)("h3",{id:"options"},"Options"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"  -h, --help   help for pubkey\n")),(0,a.kt)("h3",{id:"options-inherited-from-parent-commands"},"Options inherited from parent commands"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},'  -q, --color string            Enable colorization of output (one of: auto, always, never) (default "auto")\n  -l, --log-file stringArray    path of a file to write logs to\n  -d, --log-level stringArray   log level filter rule (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")\n')),(0,a.kt)("h3",{id:"see-also"},"SEE ALSO"),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/docs/usage/man/wg"},"cunicu wg"),"\t - WireGuard commands")))}d.isMDXComponent=!0}}]);