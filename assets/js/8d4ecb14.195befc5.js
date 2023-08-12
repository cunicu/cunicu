"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[7335],{3905:(e,n,t)=>{t.d(n,{Zo:()=>s,kt:()=>g});var r=t(67294);function o(e,n,t){return n in e?Object.defineProperty(e,n,{value:t,enumerable:!0,configurable:!0,writable:!0}):e[n]=t,e}function a(e,n){var t=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);n&&(r=r.filter((function(n){return Object.getOwnPropertyDescriptor(e,n).enumerable}))),t.push.apply(t,r)}return t}function i(e){for(var n=1;n<arguments.length;n++){var t=null!=arguments[n]?arguments[n]:{};n%2?a(Object(t),!0).forEach((function(n){o(e,n,t[n])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(t)):a(Object(t)).forEach((function(n){Object.defineProperty(e,n,Object.getOwnPropertyDescriptor(t,n))}))}return e}function c(e,n){if(null==e)return{};var t,r,o=function(e,n){if(null==e)return{};var t,r,o={},a=Object.keys(e);for(r=0;r<a.length;r++)t=a[r],n.indexOf(t)>=0||(o[t]=e[t]);return o}(e,n);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);for(r=0;r<a.length;r++)t=a[r],n.indexOf(t)>=0||Object.prototype.propertyIsEnumerable.call(e,t)&&(o[t]=e[t])}return o}var l=r.createContext({}),u=function(e){var n=r.useContext(l),t=n;return e&&(t="function"==typeof e?e(n):i(i({},n),e)),t},s=function(e){var n=u(e.components);return r.createElement(l.Provider,{value:n},e.children)},d="mdxType",f={inlineCode:"code",wrapper:function(e){var n=e.children;return r.createElement(r.Fragment,{},n)}},p=r.forwardRef((function(e,n){var t=e.components,o=e.mdxType,a=e.originalType,l=e.parentName,s=c(e,["components","mdxType","originalType","parentName"]),d=u(t),p=o,g=d["".concat(l,".").concat(p)]||d[p]||f[p]||a;return t?r.createElement(g,i(i({ref:n},s),{},{components:t})):r.createElement(g,i({ref:n},s))}));function g(e,n){var t=arguments,o=n&&n.mdxType;if("string"==typeof e||o){var a=t.length,i=new Array(a);i[0]=p;var c={};for(var l in n)hasOwnProperty.call(n,l)&&(c[l]=n[l]);c.originalType=e,c[d]="string"==typeof e?e:o,i[1]=c;for(var u=2;u<a;u++)i[u]=t[u];return r.createElement.apply(null,i)}return r.createElement.apply(null,t)}p.displayName="MDXCreateElement"},73798:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>l,contentTitle:()=>i,default:()=>f,frontMatter:()=>a,metadata:()=>c,toc:()=>u});var r=t(87462),o=(t(67294),t(3905));const a={title:"cunicu config reload",sidebar_label:"config reload",sidebar_class_name:"command-name",slug:"/usage/man/config/reload",hide_title:!0,keywords:["manpage"]},i=void 0,c={unversionedId:"usage/md/cunicu_config_reload",id:"usage/md/cunicu_config_reload",title:"cunicu config reload",description:"cunicu config reload",source:"@site/docs/usage/md/cunicu_config_reload.md",sourceDirName:"usage/md",slug:"/usage/man/config/reload",permalink:"/docs/usage/man/config/reload",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/usage/md/cunicu_config_reload.md",tags:[],version:"current",frontMatter:{title:"cunicu config reload",sidebar_label:"config reload",sidebar_class_name:"command-name",slug:"/usage/man/config/reload",hide_title:!0,keywords:["manpage"]},sidebar:"tutorialSidebar",previous:{title:"config get",permalink:"/docs/usage/man/config/get"},next:{title:"config set",permalink:"/docs/usage/man/config/set"}},l={},u=[{value:"cunicu config reload",id:"cunicu-config-reload",level:2},{value:"Options",id:"options",level:3},{value:"Options inherited from parent commands",id:"options-inherited-from-parent-commands",level:3},{value:"SEE ALSO",id:"see-also",level:3}],s={toc:u},d="wrapper";function f(e){let{components:n,...t}=e;return(0,o.kt)(d,(0,r.Z)({},s,t,{components:n,mdxType:"MDXLayout"}),(0,o.kt)("h2",{id:"cunicu-config-reload"},"cunicu config reload"),(0,o.kt)("p",null,"Reload the configuration of the cun\u012bcu daemon"),(0,o.kt)("pre",null,(0,o.kt)("code",{parentName:"pre"},"cunicu config reload [flags]\n")),(0,o.kt)("h3",{id:"options"},"Options"),(0,o.kt)("pre",null,(0,o.kt)("code",{parentName:"pre"},"  -h, --help   help for reload\n")),(0,o.kt)("h3",{id:"options-inherited-from-parent-commands"},"Options inherited from parent commands"),(0,o.kt)("pre",null,(0,o.kt)("code",{parentName:"pre"},'  -q, --color string            Enable colorization of output (one of: auto, always, never) (default "auto")\n  -l, --log-file stringArray    path of a file to write logs to\n  -d, --log-level stringArray   log level filter rule (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")\n  -s, --rpc-socket string       Unix control and monitoring socket (default "/var/run/cunicu.sock")\n')),(0,o.kt)("h3",{id:"see-also"},"SEE ALSO"),(0,o.kt)("ul",null,(0,o.kt)("li",{parentName:"ul"},(0,o.kt)("a",{parentName:"li",href:"/docs/usage/man/config"},"cunicu config"),"\t - Manage configuration of a running cun\u012bcu daemon.")))}f.isMDXComponent=!0}}]);