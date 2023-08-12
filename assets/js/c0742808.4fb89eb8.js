"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[4015],{3905:(e,t,r)=>{r.d(t,{Zo:()=>u,kt:()=>m});var n=r(67294);function a(e,t,r){return t in e?Object.defineProperty(e,t,{value:r,enumerable:!0,configurable:!0,writable:!0}):e[t]=r,e}function o(e,t){var r=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);t&&(n=n.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),r.push.apply(r,n)}return r}function i(e){for(var t=1;t<arguments.length;t++){var r=null!=arguments[t]?arguments[t]:{};t%2?o(Object(r),!0).forEach((function(t){a(e,t,r[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(r)):o(Object(r)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(r,t))}))}return e}function s(e,t){if(null==e)return{};var r,n,a=function(e,t){if(null==e)return{};var r,n,a={},o=Object.keys(e);for(n=0;n<o.length;n++)r=o[n],t.indexOf(r)>=0||(a[r]=e[r]);return a}(e,t);if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(n=0;n<o.length;n++)r=o[n],t.indexOf(r)>=0||Object.prototype.propertyIsEnumerable.call(e,r)&&(a[r]=e[r])}return a}var c=n.createContext({}),l=function(e){var t=n.useContext(c),r=t;return e&&(r="function"==typeof e?e(t):i(i({},t),e)),r},u=function(e){var t=l(e.components);return n.createElement(c.Provider,{value:t},e.children)},p="mdxType",f={inlineCode:"code",wrapper:function(e){var t=e.children;return n.createElement(n.Fragment,{},t)}},d=n.forwardRef((function(e,t){var r=e.components,a=e.mdxType,o=e.originalType,c=e.parentName,u=s(e,["components","mdxType","originalType","parentName"]),p=l(r),d=a,m=p["".concat(c,".").concat(d)]||p[d]||f[d]||o;return r?n.createElement(m,i(i({ref:t},u),{},{components:r})):n.createElement(m,i({ref:t},u))}));function m(e,t){var r=arguments,a=t&&t.mdxType;if("string"==typeof e||a){var o=r.length,i=new Array(o);i[0]=d;var s={};for(var c in t)hasOwnProperty.call(t,c)&&(s[c]=t[c]);s.originalType=e,s[p]="string"==typeof e?e:a,i[1]=s;for(var l=2;l<o;l++)i[l]=r[l];return n.createElement.apply(null,i)}return n.createElement.apply(null,r)}d.displayName="MDXCreateElement"},41320:(e,t,r)=>{r.r(t),r.d(t,{assets:()=>c,contentTitle:()=>i,default:()=>f,frontMatter:()=>o,metadata:()=>s,toc:()=>l});var n=r(87462),a=(r(67294),r(3905));const o={},i="Features",s={unversionedId:"features/index",id:"features/index",title:"Features",description:"The cun\u012bcu daemon supports many features which are implemented by separate software modules/packages.",source:"@site/docs/features/index.md",sourceDirName:"features",slug:"/features/",permalink:"/docs/features/",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/features/index.md",tags:[],version:"current",frontMatter:{},sidebar:"tutorialSidebar",previous:{title:"Installation",permalink:"/docs/install"},next:{title:"Auto-configuration",permalink:"/docs/features/autocfg"}},c={},l=[],u={toc:l},p="wrapper";function f(e){let{components:t,...r}=e;return(0,a.kt)(p,(0,n.Z)({},u,r,{components:t,mdxType:"MDXLayout"}),(0,a.kt)("h1",{id:"features"},"Features"),(0,a.kt)("p",null,"The cun\u012bcu daemon supports many features which are implemented by separate software modules/packages.\nThis structure promotes the ",(0,a.kt)("a",{parentName:"p",href:"https://en.wikipedia.org/wiki/Separation_of_concerns"},"separation of concerns")," within the code-base and allows for use-cases in which only subsets of features are used.\nE.g. we can use cun\u012bcu for the post-quantum safe exchange of pre-shared keys without any of the other features like peer or endpoint discovery. With very few exceptions all of the features listed below can be used separately."),(0,a.kt)("p",null,"Currently, the following features are implemented as separate modules:"),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/docs/features/autocfg"},"Auto-configuration of missing interface settings and link-local IP addresses")," (",(0,a.kt)("inlineCode",{parentName:"li"},"autocfg"),")"),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/docs/features/cfgsync"},"Config Synchronization")," (",(0,a.kt)("inlineCode",{parentName:"li"},"cfgsync"),")"),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/docs/features/pdisc"},"Peer Discovery")," (",(0,a.kt)("inlineCode",{parentName:"li"},"pdisc"),")"),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/docs/features/epdisc"},"Endpoint Discovery")," (",(0,a.kt)("inlineCode",{parentName:"li"},"epdisc"),")"),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/docs/features/hooks"},"Hooks")," (",(0,a.kt)("inlineCode",{parentName:"li"},"hooks"),")"),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/docs/features/hsync"},"Hosts-file Synchronization")," (",(0,a.kt)("inlineCode",{parentName:"li"},"hsync"),")"),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/docs/features/pske"},"Pre-shared Key Establishment")," (",(0,a.kt)("inlineCode",{parentName:"li"},"pske"),")"),(0,a.kt)("li",{parentName:"ul"},(0,a.kt)("a",{parentName:"li",href:"/docs/features/rtsync"},"Route Synchronization")," (",(0,a.kt)("inlineCode",{parentName:"li"},"rtsync"),")")))}f.isMDXComponent=!0}}]);