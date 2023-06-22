"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[155],{4137:(e,t,n)=>{n.d(t,{Zo:()=>p,kt:()=>d});var o=n(7294);function i(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function r(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);t&&(o=o.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,o)}return n}function l(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?r(Object(n),!0).forEach((function(t){i(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):r(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function c(e,t){if(null==e)return{};var n,o,i=function(e,t){if(null==e)return{};var n,o,i={},r=Object.keys(e);for(o=0;o<r.length;o++)n=r[o],t.indexOf(n)>=0||(i[n]=e[n]);return i}(e,t);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);for(o=0;o<r.length;o++)n=r[o],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(i[n]=e[n])}return i}var a=o.createContext({}),s=function(e){var t=o.useContext(a),n=t;return e&&(n="function"==typeof e?e(t):l(l({},t),e)),n},p=function(e){var t=s(e.components);return o.createElement(a.Provider,{value:t},e.children)},u="mdxType",m={inlineCode:"code",wrapper:function(e){var t=e.children;return o.createElement(o.Fragment,{},t)}},f=o.forwardRef((function(e,t){var n=e.components,i=e.mdxType,r=e.originalType,a=e.parentName,p=c(e,["components","mdxType","originalType","parentName"]),u=s(n),f=i,d=u["".concat(a,".").concat(f)]||u[f]||m[f]||r;return n?o.createElement(d,l(l({ref:t},p),{},{components:n})):o.createElement(d,l({ref:t},p))}));function d(e,t){var n=arguments,i=t&&t.mdxType;if("string"==typeof e||i){var r=n.length,l=new Array(r);l[0]=f;var c={};for(var a in t)hasOwnProperty.call(t,a)&&(c[a]=t[a]);c.originalType=e,c[u]="string"==typeof e?e:i,l[1]=c;for(var s=2;s<r;s++)l[s]=n[s];return o.createElement.apply(null,l)}return o.createElement.apply(null,n)}f.displayName="MDXCreateElement"},1272:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>a,contentTitle:()=>l,default:()=>m,frontMatter:()=>r,metadata:()=>c,toc:()=>s});var o=n(7462),i=(n(7294),n(4137));const r={title:"cunicu completion fish",sidebar_label:"completion fish",sidebar_class_name:"command-name",slug:"/usage/man/completion/fish",hide_title:!0,keywords:["manpage"]},l=void 0,c={unversionedId:"usage/md/cunicu_completion_fish",id:"usage/md/cunicu_completion_fish",title:"cunicu completion fish",description:"cunicu completion fish",source:"@site/docs/usage/md/cunicu_completion_fish.md",sourceDirName:"usage/md",slug:"/usage/man/completion/fish",permalink:"/docs/usage/man/completion/fish",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/usage/md/cunicu_completion_fish.md",tags:[],version:"current",frontMatter:{title:"cunicu completion fish",sidebar_label:"completion fish",sidebar_class_name:"command-name",slug:"/usage/man/completion/fish",hide_title:!0,keywords:["manpage"]},sidebar:"tutorialSidebar",previous:{title:"completion bash",permalink:"/docs/usage/man/completion/bash"},next:{title:"completion powershell",permalink:"/docs/usage/man/completion/powershell"}},a={},s=[{value:"cunicu completion fish",id:"cunicu-completion-fish",level:2},{value:"Synopsis",id:"synopsis",level:3},{value:"Options",id:"options",level:3},{value:"Options inherited from parent commands",id:"options-inherited-from-parent-commands",level:3},{value:"SEE ALSO",id:"see-also",level:3}],p={toc:s},u="wrapper";function m(e){let{components:t,...n}=e;return(0,i.kt)(u,(0,o.Z)({},p,n,{components:t,mdxType:"MDXLayout"}),(0,i.kt)("h2",{id:"cunicu-completion-fish"},"cunicu completion fish"),(0,i.kt)("p",null,"Generate the autocompletion script for fish"),(0,i.kt)("h3",{id:"synopsis"},"Synopsis"),(0,i.kt)("p",null,"Generate the autocompletion script for the fish shell."),(0,i.kt)("p",null,"To load completions in your current shell session:"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre"},"cunicu completion fish | source\n")),(0,i.kt)("p",null,"To load completions for every new session, execute once:"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre"},"cunicu completion fish > ~/.config/fish/completions/cunicu.fish\n")),(0,i.kt)("p",null,"You will need to start a new shell for this setup to take effect."),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre"},"cunicu completion fish [flags]\n")),(0,i.kt)("h3",{id:"options"},"Options"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre"},"  -h, --help              help for fish\n      --no-descriptions   disable completion descriptions\n")),(0,i.kt)("h3",{id:"options-inherited-from-parent-commands"},"Options inherited from parent commands"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre"},'  -q, --color string       Enable colorization of output (one of: auto, always, never) (default "auto")\n  -l, --log-file string    path of a file to write logs to\n  -d, --log-level string   log level filter rule (one of: debug, info, warn, error, dpanic, panic, and fatal) (default "info")\n')),(0,i.kt)("h3",{id:"see-also"},"SEE ALSO"),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"/docs/usage/man/completion"},"cunicu completion"),"\t - Generate the autocompletion script for the specified shell")))}m.isMDXComponent=!0}}]);