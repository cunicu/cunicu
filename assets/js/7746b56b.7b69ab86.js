"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[5015],{3905:(e,n,r)=>{r.d(n,{Zo:()=>u,kt:()=>f});var t=r(67294);function o(e,n,r){return n in e?Object.defineProperty(e,n,{value:r,enumerable:!0,configurable:!0,writable:!0}):e[n]=r,e}function a(e,n){var r=Object.keys(e);if(Object.getOwnPropertySymbols){var t=Object.getOwnPropertySymbols(e);n&&(t=t.filter((function(n){return Object.getOwnPropertyDescriptor(e,n).enumerable}))),r.push.apply(r,t)}return r}function i(e){for(var n=1;n<arguments.length;n++){var r=null!=arguments[n]?arguments[n]:{};n%2?a(Object(r),!0).forEach((function(n){o(e,n,r[n])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(r)):a(Object(r)).forEach((function(n){Object.defineProperty(e,n,Object.getOwnPropertyDescriptor(r,n))}))}return e}function s(e,n){if(null==e)return{};var r,t,o=function(e,n){if(null==e)return{};var r,t,o={},a=Object.keys(e);for(t=0;t<a.length;t++)r=a[t],n.indexOf(r)>=0||(o[r]=e[r]);return o}(e,n);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);for(t=0;t<a.length;t++)r=a[t],n.indexOf(r)>=0||Object.prototype.propertyIsEnumerable.call(e,r)&&(o[r]=e[r])}return o}var l=t.createContext({}),c=function(e){var n=t.useContext(l),r=n;return e&&(r="function"==typeof e?e(n):i(i({},n),e)),r},u=function(e){var n=c(e.components);return t.createElement(l.Provider,{value:n},e.children)},p="mdxType",m={inlineCode:"code",wrapper:function(e){var n=e.children;return t.createElement(t.Fragment,{},n)}},d=t.forwardRef((function(e,n){var r=e.components,o=e.mdxType,a=e.originalType,l=e.parentName,u=s(e,["components","mdxType","originalType","parentName"]),p=c(r),d=o,f=p["".concat(l,".").concat(d)]||p[d]||m[d]||a;return r?t.createElement(f,i(i({ref:n},u),{},{components:r})):t.createElement(f,i({ref:n},u))}));function f(e,n){var r=arguments,o=n&&n.mdxType;if("string"==typeof e||o){var a=r.length,i=new Array(a);i[0]=d;var s={};for(var l in n)hasOwnProperty.call(n,l)&&(s[l]=n[l]);s.originalType=e,s[p]="string"==typeof e?e:o,i[1]=s;for(var c=2;c<a;c++)i[c]=r[c];return t.createElement.apply(null,i)}return t.createElement.apply(null,r)}d.displayName="MDXCreateElement"},67700:(e,n,r)=>{r.r(n),r.d(n,{assets:()=>l,contentTitle:()=>i,default:()=>m,frontMatter:()=>a,metadata:()=>s,toc:()=>c});var t=r(87462),o=(r(67294),r(3905));const a={title:"cunicu version",sidebar_label:"version",sidebar_class_name:"command-name",slug:"/usage/man/version",hide_title:!0,keywords:["manpage"]},i=void 0,s={unversionedId:"usage/md/cunicu_version",id:"usage/md/cunicu_version",title:"cunicu version",description:"cunicu version",source:"@site/docs/usage/md/cunicu_version.md",sourceDirName:"usage/md",slug:"/usage/man/version",permalink:"/docs/usage/man/version",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/usage/md/cunicu_version.md",tags:[],version:"current",frontMatter:{title:"cunicu version",sidebar_label:"version",sidebar_class_name:"command-name",slug:"/usage/man/version",hide_title:!0,keywords:["manpage"]},sidebar:"tutorialSidebar",previous:{title:"sync",permalink:"/docs/usage/man/sync"},next:{title:"wg",permalink:"/docs/usage/man/wg"}},l={},c=[{value:"cunicu version",id:"cunicu-version",level:2},{value:"Examples",id:"examples",level:3},{value:"Options",id:"options",level:3},{value:"Options inherited from parent commands",id:"options-inherited-from-parent-commands",level:3},{value:"SEE ALSO",id:"see-also",level:3}],u={toc:c},p="wrapper";function m(e){let{components:n,...r}=e;return(0,o.kt)(p,(0,t.Z)({},u,r,{components:n,mdxType:"MDXLayout"}),(0,o.kt)("h2",{id:"cunicu-version"},"cunicu version"),(0,o.kt)("p",null,"Show version of the cun\u012bcu binary and optionally also a running daemon"),(0,o.kt)("pre",null,(0,o.kt)("code",{parentName:"pre"},"cunicu version [flags]\n")),(0,o.kt)("h3",{id:"examples"},"Examples"),(0,o.kt)("pre",null,(0,o.kt)("code",{parentName:"pre"},"$ sudo cunicu version\nclient: v0.1.2 (os=linux, arch=arm64, commit=b22ee3e7, branch=master, built-at=2022-09-09T13:44:22+02:00, built-by=goreleaser)\ndaemon: v0.1.2 (os=linux, arch=arm64, commit=b22ee3e7, branch=master, built-at=2022-09-09T13:44:22+02:00, built-by=goreleaser)\n")),(0,o.kt)("h3",{id:"options"},"Options"),(0,o.kt)("pre",null,(0,o.kt)("code",{parentName:"pre"},'  -f, --format format   Output format (one of: human, json) (default "human")\n  -h, --help            help for version\n  -s, --short           Only show version and nothing else\n')),(0,o.kt)("h3",{id:"options-inherited-from-parent-commands"},"Options inherited from parent commands"),(0,o.kt)("pre",null,(0,o.kt)("code",{parentName:"pre"},'  -q, --color string            Enable colorization of output (one of: auto, always, never) (default "auto")\n  -l, --log-file stringArray    path of a file to write logs to\n  -d, --log-level stringArray   log level filter rule (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")\n')),(0,o.kt)("h3",{id:"see-also"},"SEE ALSO"),(0,o.kt)("ul",null,(0,o.kt)("li",{parentName:"ul"},(0,o.kt)("a",{parentName:"li",href:"/docs/usage/man/"},"cunicu"),"\t - cun\u012bcu is a user-space daemon managing WireGuard\xae interfaces to establish peer-to-peer connections in harsh network environments.")))}m.isMDXComponent=!0}}]);