"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[6812],{4137:(e,n,t)=>{t.d(n,{Zo:()=>s,kt:()=>d});var r=t(7294);function o(e,n,t){return n in e?Object.defineProperty(e,n,{value:t,enumerable:!0,configurable:!0,writable:!0}):e[n]=t,e}function i(e,n){var t=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);n&&(r=r.filter((function(n){return Object.getOwnPropertyDescriptor(e,n).enumerable}))),t.push.apply(t,r)}return t}function a(e){for(var n=1;n<arguments.length;n++){var t=null!=arguments[n]?arguments[n]:{};n%2?i(Object(t),!0).forEach((function(n){o(e,n,t[n])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(t)):i(Object(t)).forEach((function(n){Object.defineProperty(e,n,Object.getOwnPropertyDescriptor(t,n))}))}return e}function c(e,n){if(null==e)return{};var t,r,o=function(e,n){if(null==e)return{};var t,r,o={},i=Object.keys(e);for(r=0;r<i.length;r++)t=i[r],n.indexOf(t)>=0||(o[t]=e[t]);return o}(e,n);if(Object.getOwnPropertySymbols){var i=Object.getOwnPropertySymbols(e);for(r=0;r<i.length;r++)t=i[r],n.indexOf(t)>=0||Object.prototype.propertyIsEnumerable.call(e,t)&&(o[t]=e[t])}return o}var u=r.createContext({}),l=function(e){var n=r.useContext(u),t=n;return e&&(t="function"==typeof e?e(n):a(a({},n),e)),t},s=function(e){var n=l(e.components);return r.createElement(u.Provider,{value:n},e.children)},p="mdxType",f={inlineCode:"code",wrapper:function(e){var n=e.children;return r.createElement(r.Fragment,{},n)}},m=r.forwardRef((function(e,n){var t=e.components,o=e.mdxType,i=e.originalType,u=e.parentName,s=c(e,["components","mdxType","originalType","parentName"]),p=l(t),m=o,d=p["".concat(u,".").concat(m)]||p[m]||f[m]||i;return t?r.createElement(d,a(a({ref:n},s),{},{components:t})):r.createElement(d,a({ref:n},s))}));function d(e,n){var t=arguments,o=n&&n.mdxType;if("string"==typeof e||o){var i=t.length,a=new Array(i);a[0]=m;var c={};for(var u in n)hasOwnProperty.call(n,u)&&(c[u]=n[u]);c.originalType=e,c[p]="string"==typeof e?e:o,a[1]=c;for(var l=2;l<i;l++)a[l]=t[l];return r.createElement.apply(null,a)}return r.createElement.apply(null,t)}m.displayName="MDXCreateElement"},5490:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>u,contentTitle:()=>a,default:()=>f,frontMatter:()=>i,metadata:()=>c,toc:()=>l});var r=t(7462),o=(t(7294),t(4137));const i={title:"cunicu config",sidebar_label:"config",sidebar_class_name:"command-name",slug:"/usage/man/config",hide_title:!0,keywords:["manpage"]},a=void 0,c={unversionedId:"usage/md/cunicu_config",id:"usage/md/cunicu_config",title:"cunicu config",description:"cunicu config",source:"@site/docs/usage/md/cunicu_config.md",sourceDirName:"usage/md",slug:"/usage/man/config",permalink:"/docs/usage/man/config",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/usage/md/cunicu_config.md",tags:[],version:"current",frontMatter:{title:"cunicu config",sidebar_label:"config",sidebar_class_name:"command-name",slug:"/usage/man/config",hide_title:!0,keywords:["manpage"]},sidebar:"tutorialSidebar",previous:{title:"completion zsh",permalink:"/docs/usage/man/completion/zsh"},next:{title:"config get",permalink:"/docs/usage/man/config/get"}},u={},l=[{value:"cunicu config",id:"cunicu-config",level:2},{value:"Synopsis",id:"synopsis",level:3},{value:"Options",id:"options",level:3},{value:"Options inherited from parent commands",id:"options-inherited-from-parent-commands",level:3},{value:"SEE ALSO",id:"see-also",level:3}],s={toc:l},p="wrapper";function f(e){let{components:n,...t}=e;return(0,o.kt)(p,(0,r.Z)({},s,t,{components:n,mdxType:"MDXLayout"}),(0,o.kt)("h2",{id:"cunicu-config"},"cunicu config"),(0,o.kt)("p",null,"Manage configuration of a running cun\u012bcu daemon."),(0,o.kt)("h3",{id:"synopsis"},"Synopsis"),(0,o.kt)("h3",{id:"options"},"Options"),(0,o.kt)("pre",null,(0,o.kt)("code",{parentName:"pre"},'  -h, --help                help for config\n  -s, --rpc-socket string   Unix control and monitoring socket (default "/var/run/cunicu.sock")\n')),(0,o.kt)("h3",{id:"options-inherited-from-parent-commands"},"Options inherited from parent commands"),(0,o.kt)("pre",null,(0,o.kt)("code",{parentName:"pre"},'  -q, --color string       Enable colorization of output (one of: auto, always, never) (default "auto")\n  -l, --log-file string    path of a file to write logs to\n  -d, --log-level string   log level filter rule (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")\n')),(0,o.kt)("h3",{id:"see-also"},"SEE ALSO"),(0,o.kt)("ul",null,(0,o.kt)("li",{parentName:"ul"},(0,o.kt)("a",{parentName:"li",href:"/docs/usage/man/"},"cunicu"),"\t - cun\u012bcu is a user-space daemon managing WireGuard\xae interfaces to establish peer-to-peer connections in harsh network environments."),(0,o.kt)("li",{parentName:"ul"},(0,o.kt)("a",{parentName:"li",href:"/docs/usage/man/config/get"},"cunicu config get"),"\t - Get current value of a configuration setting"),(0,o.kt)("li",{parentName:"ul"},(0,o.kt)("a",{parentName:"li",href:"/docs/usage/man/config/reload"},"cunicu config reload"),"\t - Reload the configuration of the cun\u012bcu daemon"),(0,o.kt)("li",{parentName:"ul"},(0,o.kt)("a",{parentName:"li",href:"/docs/usage/man/config/set"},"cunicu config set"),"\t - Update the value of a configuration setting")))}f.isMDXComponent=!0}}]);