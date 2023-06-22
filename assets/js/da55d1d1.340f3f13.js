"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[4398],{4137:(e,n,t)=>{t.d(n,{Zo:()=>s,kt:()=>f});var r=t(7294);function o(e,n,t){return n in e?Object.defineProperty(e,n,{value:t,enumerable:!0,configurable:!0,writable:!0}):e[n]=t,e}function a(e,n){var t=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);n&&(r=r.filter((function(n){return Object.getOwnPropertyDescriptor(e,n).enumerable}))),t.push.apply(t,r)}return t}function i(e){for(var n=1;n<arguments.length;n++){var t=null!=arguments[n]?arguments[n]:{};n%2?a(Object(t),!0).forEach((function(n){o(e,n,t[n])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(t)):a(Object(t)).forEach((function(n){Object.defineProperty(e,n,Object.getOwnPropertyDescriptor(t,n))}))}return e}function c(e,n){if(null==e)return{};var t,r,o=function(e,n){if(null==e)return{};var t,r,o={},a=Object.keys(e);for(r=0;r<a.length;r++)t=a[r],n.indexOf(t)>=0||(o[t]=e[t]);return o}(e,n);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);for(r=0;r<a.length;r++)t=a[r],n.indexOf(t)>=0||Object.prototype.propertyIsEnumerable.call(e,t)&&(o[t]=e[t])}return o}var u=r.createContext({}),l=function(e){var n=r.useContext(u),t=n;return e&&(t="function"==typeof e?e(n):i(i({},n),e)),t},s=function(e){var n=l(e.components);return r.createElement(u.Provider,{value:n},e.children)},m="mdxType",p={inlineCode:"code",wrapper:function(e){var n=e.children;return r.createElement(r.Fragment,{},n)}},d=r.forwardRef((function(e,n){var t=e.components,o=e.mdxType,a=e.originalType,u=e.parentName,s=c(e,["components","mdxType","originalType","parentName"]),m=l(t),d=o,f=m["".concat(u,".").concat(d)]||m[d]||p[d]||a;return t?r.createElement(f,i(i({ref:n},s),{},{components:t})):r.createElement(f,i({ref:n},s))}));function f(e,n){var t=arguments,o=n&&n.mdxType;if("string"==typeof e||o){var a=t.length,i=new Array(a);i[0]=d;var c={};for(var u in n)hasOwnProperty.call(n,u)&&(c[u]=n[u]);c.originalType=e,c[m]="string"==typeof e?e:o,i[1]=c;for(var l=2;l<a;l++)i[l]=t[l];return r.createElement.apply(null,i)}return r.createElement.apply(null,t)}d.displayName="MDXCreateElement"},5055:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>u,contentTitle:()=>i,default:()=>p,frontMatter:()=>a,metadata:()=>c,toc:()=>l});var r=t(7462),o=(t(7294),t(4137));const a={title:"cunicu monitor",sidebar_label:"monitor",sidebar_class_name:"command-name",slug:"/usage/man/monitor",hide_title:!0,keywords:["manpage"]},i=void 0,c={unversionedId:"usage/md/cunicu_monitor",id:"usage/md/cunicu_monitor",title:"cunicu monitor",description:"cunicu monitor",source:"@site/docs/usage/md/cunicu_monitor.md",sourceDirName:"usage/md",slug:"/usage/man/monitor",permalink:"/docs/usage/man/monitor",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/usage/md/cunicu_monitor.md",tags:[],version:"current",frontMatter:{title:"cunicu monitor",sidebar_label:"monitor",sidebar_class_name:"command-name",slug:"/usage/man/monitor",hide_title:!0,keywords:["manpage"]},sidebar:"tutorialSidebar",previous:{title:"invite",permalink:"/docs/usage/man/invite"},next:{title:"relay",permalink:"/docs/usage/man/relay"}},u={},l=[{value:"cunicu monitor",id:"cunicu-monitor",level:2},{value:"Options",id:"options",level:3},{value:"Options inherited from parent commands",id:"options-inherited-from-parent-commands",level:3},{value:"SEE ALSO",id:"see-also",level:3}],s={toc:l},m="wrapper";function p(e){let{components:n,...t}=e;return(0,o.kt)(m,(0,r.Z)({},s,t,{components:n,mdxType:"MDXLayout"}),(0,o.kt)("h2",{id:"cunicu-monitor"},"cunicu monitor"),(0,o.kt)("p",null,"Monitor the cun\u012bcu daemon for events"),(0,o.kt)("pre",null,(0,o.kt)("code",{parentName:"pre"},"cunicu monitor [flags]\n")),(0,o.kt)("h3",{id:"options"},"Options"),(0,o.kt)("pre",null,(0,o.kt)("code",{parentName:"pre"},'  -f, --format format       Output format (one of: json, logger, human) (default "human")\n  -h, --help                help for monitor\n  -s, --rpc-socket string   Unix control and monitoring socket (default "/var/run/cunicu.sock")\n')),(0,o.kt)("h3",{id:"options-inherited-from-parent-commands"},"Options inherited from parent commands"),(0,o.kt)("pre",null,(0,o.kt)("code",{parentName:"pre"},'  -q, --color string       Enable colorization of output (one of: auto, always, never) (default "auto")\n  -l, --log-file string    path of a file to write logs to\n  -d, --log-level string   log level filter rule (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")\n')),(0,o.kt)("h3",{id:"see-also"},"SEE ALSO"),(0,o.kt)("ul",null,(0,o.kt)("li",{parentName:"ul"},(0,o.kt)("a",{parentName:"li",href:"/docs/usage/man/"},"cunicu"),"\t - cun\u012bcu is a user-space daemon managing WireGuard\xae interfaces to establish peer-to-peer connections in harsh network environments.")))}p.isMDXComponent=!0}}]);