"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[9901],{3905:(e,n,t)=>{t.d(n,{Zo:()=>u,kt:()=>f});var r=t(67294);function a(e,n,t){return n in e?Object.defineProperty(e,n,{value:t,enumerable:!0,configurable:!0,writable:!0}):e[n]=t,e}function o(e,n){var t=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);n&&(r=r.filter((function(n){return Object.getOwnPropertyDescriptor(e,n).enumerable}))),t.push.apply(t,r)}return t}function i(e){for(var n=1;n<arguments.length;n++){var t=null!=arguments[n]?arguments[n]:{};n%2?o(Object(t),!0).forEach((function(n){a(e,n,t[n])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(t)):o(Object(t)).forEach((function(n){Object.defineProperty(e,n,Object.getOwnPropertyDescriptor(t,n))}))}return e}function c(e,n){if(null==e)return{};var t,r,a=function(e,n){if(null==e)return{};var t,r,a={},o=Object.keys(e);for(r=0;r<o.length;r++)t=o[r],n.indexOf(t)>=0||(a[t]=e[t]);return a}(e,n);if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(r=0;r<o.length;r++)t=o[r],n.indexOf(t)>=0||Object.prototype.propertyIsEnumerable.call(e,t)&&(a[t]=e[t])}return a}var s=r.createContext({}),l=function(e){var n=r.useContext(s),t=n;return e&&(t="function"==typeof e?e(n):i(i({},n),e)),t},u=function(e){var n=l(e.components);return r.createElement(s.Provider,{value:n},e.children)},d="mdxType",p={inlineCode:"code",wrapper:function(e){var n=e.children;return r.createElement(r.Fragment,{},n)}},m=r.forwardRef((function(e,n){var t=e.components,a=e.mdxType,o=e.originalType,s=e.parentName,u=c(e,["components","mdxType","originalType","parentName"]),d=l(t),m=a,f=d["".concat(s,".").concat(m)]||d[m]||p[m]||o;return t?r.createElement(f,i(i({ref:n},u),{},{components:t})):r.createElement(f,i({ref:n},u))}));function f(e,n){var t=arguments,a=n&&n.mdxType;if("string"==typeof e||a){var o=t.length,i=new Array(o);i[0]=m;var c={};for(var s in n)hasOwnProperty.call(n,s)&&(c[s]=n[s]);c.originalType=e,c[d]="string"==typeof e?e:a,i[1]=c;for(var l=2;l<o;l++)i[l]=t[l];return r.createElement.apply(null,i)}return r.createElement.apply(null,t)}m.displayName="MDXCreateElement"},70630:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>s,contentTitle:()=>i,default:()=>p,frontMatter:()=>o,metadata:()=>c,toc:()=>l});var r=t(87462),a=(t(67294),t(3905));const o={title:"cunicu daemon",sidebar_label:"daemon",sidebar_class_name:"command-name",slug:"/usage/man/daemon",hide_title:!0,keywords:["manpage"]},i=void 0,c={unversionedId:"usage/md/cunicu_daemon",id:"usage/md/cunicu_daemon",title:"cunicu daemon",description:"cunicu daemon",source:"@site/docs/usage/md/cunicu_daemon.md",sourceDirName:"usage/md",slug:"/usage/man/daemon",permalink:"/docs/usage/man/daemon",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/usage/md/cunicu_daemon.md",tags:[],version:"current",frontMatter:{title:"cunicu daemon",sidebar_label:"daemon",sidebar_class_name:"command-name",slug:"/usage/man/daemon",hide_title:!0,keywords:["manpage"]},sidebar:"tutorialSidebar",previous:{title:"config set",permalink:"/docs/usage/man/config/set"},next:{title:"invite",permalink:"/docs/usage/man/invite"}},s={},l=[{value:"cunicu daemon",id:"cunicu-daemon",level:2},{value:"Synopsis",id:"synopsis",level:3},{value:"Examples",id:"examples",level:3},{value:"Options",id:"options",level:3},{value:"Options inherited from parent commands",id:"options-inherited-from-parent-commands",level:3},{value:"SEE ALSO",id:"see-also",level:3}],u={toc:l},d="wrapper";function p(e){let{components:n,...t}=e;return(0,a.kt)(d,(0,r.Z)({},u,t,{components:n,mdxType:"MDXLayout"}),(0,a.kt)("h2",{id:"cunicu-daemon"},"cunicu daemon"),(0,a.kt)("p",null,"Start the main daemon"),(0,a.kt)("h3",{id:"synopsis"},"Synopsis"),(0,a.kt)("p",null,"Starts the main cunicu agent."),(0,a.kt)("p",null,"Sending a SIGUSR1 signal to the daemon will trigger an immediate synchronization of all WireGuard interfaces."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"cunicu daemon [interface-names...] [flags]\n")),(0,a.kt)("h3",{id:"examples"},"Examples"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"$ cunicu daemon -U -x mysecretpass wg0\n")),(0,a.kt)("h3",{id:"options"},"Options"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},"  -b, --backend URL               One or more URLs to signaling backends\n  -x, --community passphrase      A passphrase shared with other peers in the same community\n  -c, --config filename           One or more filenames of configuration files\n  -E, --discover-endpoints        Enable ICE endpoint discovery (default true)\n  -P, --discover-peers            Enable peer discovery (default true)\n  -D, --domain domain             A DNS domain name used for DNS auto-configuration\n  -n, --hostname name             A name which identifies this peer\n  -o, --option stringArray        Set arbitrary options (example: --option watch_interval=5s)\n  -F, --port-forwarding           Enabled in-kernel port-forwarding (default true)\n  -T, --routing-table int         Kernel routing table to use (default 254)\n  -s, --rpc-socket path           The path of the unix socket used by other cunicu commands\n      --rpc-wait                  Wait until first client connected to control socket before continuing start\n  -C, --sync-config               Enable synchronization of configuration files (default true)\n  -H, --sync-hosts                Enable synchronization of /etc/hosts file (default true)\n  -R, --sync-routes               Enable synchronization of AllowedIPs with Kernel routes (default true)\n  -w, --watch-config              Watch configuration for changes and apply changes at runtime.\n  -i, --watch-interval duration   An interval at which we are periodically polling the kernel for updates on WireGuard interfaces\n  -U, --wg-userspace              Use user-space WireGuard implementation for newly created interfaces\n  -h, --help                      help for daemon\n")),(0,a.kt)("h3",{id:"options-inherited-from-parent-commands"},"Options inherited from parent commands"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre"},'  -q, --color string            Enable colorization of output (one of: auto, always, never) (default "auto")\n  -l, --log-file stringArray    path of a file to write logs to\n  -d, --log-level stringArray   log level filter rule (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")\n')),(0,a.kt)("h3",{id:"see-also"},"SEE ALSO"),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/docs/usage/man/"},"cunicu"),"\t - cun\u012bcu is a user-space daemon managing WireGuard\xae interfaces to establish peer-to-peer connections in harsh network environments.")))}p.isMDXComponent=!0}}]);