"use strict";(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[3989],{3905:(e,a,t)=>{t.d(a,{Zo:()=>o,kt:()=>g});var n=t(67294);function s(e,a,t){return a in e?Object.defineProperty(e,a,{value:t,enumerable:!0,configurable:!0,writable:!0}):e[a]=t,e}function r(e,a){var t=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);a&&(n=n.filter((function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable}))),t.push.apply(t,n)}return t}function m(e){for(var a=1;a<arguments.length;a++){var t=null!=arguments[a]?arguments[a]:{};a%2?r(Object(t),!0).forEach((function(a){s(e,a,t[a])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(t)):r(Object(t)).forEach((function(a){Object.defineProperty(e,a,Object.getOwnPropertyDescriptor(t,a))}))}return e}function p(e,a){if(null==e)return{};var t,n,s=function(e,a){if(null==e)return{};var t,n,s={},r=Object.keys(e);for(n=0;n<r.length;n++)t=r[n],a.indexOf(t)>=0||(s[t]=e[t]);return s}(e,a);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);for(n=0;n<r.length;n++)t=r[n],a.indexOf(t)>=0||Object.prototype.propertyIsEnumerable.call(e,t)&&(s[t]=e[t])}return s}var i=n.createContext({}),l=function(e){var a=n.useContext(i),t=a;return e&&(t="function"==typeof e?e(a):m(m({},a),e)),t},o=function(e){var a=l(e.components);return n.createElement(i.Provider,{value:a},e.children)},c="mdxType",N={inlineCode:"code",wrapper:function(e){var a=e.children;return n.createElement(n.Fragment,{},a)}},k=n.forwardRef((function(e,a){var t=e.components,s=e.mdxType,r=e.originalType,i=e.parentName,o=p(e,["components","mdxType","originalType","parentName"]),c=l(t),k=s,g=c["".concat(i,".").concat(k)]||c[k]||N[k]||r;return t?n.createElement(g,m(m({ref:a},o),{},{components:t})):n.createElement(g,m({ref:a},o))}));function g(e,a){var t=arguments,s=a&&a.mdxType;if("string"==typeof e||s){var r=t.length,m=new Array(r);m[0]=k;var p={};for(var i in a)hasOwnProperty.call(a,i)&&(p[i]=a[i]);p.originalType=e,p[c]="string"==typeof e?e:s,m[1]=p;for(var l=2;l<r;l++)m[l]=t[l];return n.createElement.apply(null,m)}return n.createElement.apply(null,t)}k.displayName="MDXCreateElement"},72655:(e,a,t)=>{t.r(a),t.d(a,{assets:()=>i,contentTitle:()=>m,default:()=>N,frontMatter:()=>r,metadata:()=>p,toc:()=>l});var n=t(87462),s=(t(67294),t(3905));const r={},m="Session Signaling",p={unversionedId:"development/signaling",id:"development/signaling",title:"Session Signaling",description:"Lets assume two peers $Pa$ & $Pb$ are seeking to establish a ICE session.",source:"@site/docs/development/signaling.md",sourceDirName:"development",slug:"/development/signaling",permalink:"/docs/development/signaling",draft:!1,editUrl:"https://github.com/cunicu/cunicu/edit/main/docs/development/signaling.md",tags:[],version:"current",frontMatter:{},sidebar:"tutorialSidebar",previous:{title:"Proxying",permalink:"/docs/development/proxying"}},i={},l=[{value:"Session Description",id:"session-description",level:2},{value:"Backends",id:"backends",level:2},{value:"Available backends",id:"available-backends",level:3},{value:"Semantics",id:"semantics",level:3},{value:"Interface",id:"interface",level:3}],o={toc:l},c="wrapper";function N(e){let{components:a,...t}=e;return(0,s.kt)(c,(0,n.Z)({},o,t,{components:a,mdxType:"MDXLayout"}),(0,s.kt)("h1",{id:"session-signaling"},"Session Signaling"),(0,s.kt)("p",null,"Lets assume two peers ",(0,s.kt)("span",{parentName:"p",className:"math math-inline"},(0,s.kt)("span",{parentName:"span",className:"katex"},(0,s.kt)("span",{parentName:"span",className:"katex-mathml"},(0,s.kt)("math",{parentName:"span",xmlns:"http://www.w3.org/1998/Math/MathML"},(0,s.kt)("semantics",{parentName:"math"},(0,s.kt)("mrow",{parentName:"semantics"},(0,s.kt)("msub",{parentName:"mrow"},(0,s.kt)("mi",{parentName:"msub"},"P"),(0,s.kt)("mi",{parentName:"msub"},"a"))),(0,s.kt)("annotation",{parentName:"semantics",encoding:"application/x-tex"},"P_a")))),(0,s.kt)("span",{parentName:"span",className:"katex-html","aria-hidden":"true"},(0,s.kt)("span",{parentName:"span",className:"base"},(0,s.kt)("span",{parentName:"span",className:"strut",style:{height:"0.8333em",verticalAlign:"-0.15em"}}),(0,s.kt)("span",{parentName:"span",className:"mord"},(0,s.kt)("span",{parentName:"span",className:"mord mathnormal",style:{marginRight:"0.13889em"}},"P"),(0,s.kt)("span",{parentName:"span",className:"msupsub"},(0,s.kt)("span",{parentName:"span",className:"vlist-t vlist-t2"},(0,s.kt)("span",{parentName:"span",className:"vlist-r"},(0,s.kt)("span",{parentName:"span",className:"vlist",style:{height:"0.1514em"}},(0,s.kt)("span",{parentName:"span",style:{top:"-2.55em",marginLeft:"-0.1389em",marginRight:"0.05em"}},(0,s.kt)("span",{parentName:"span",className:"pstrut",style:{height:"2.7em"}}),(0,s.kt)("span",{parentName:"span",className:"sizing reset-size6 size3 mtight"},(0,s.kt)("span",{parentName:"span",className:"mord mathnormal mtight"},"a")))),(0,s.kt)("span",{parentName:"span",className:"vlist-s"},"\u200b")),(0,s.kt)("span",{parentName:"span",className:"vlist-r"},(0,s.kt)("span",{parentName:"span",className:"vlist",style:{height:"0.15em"}},(0,s.kt)("span",{parentName:"span"}))))))))))," & ",(0,s.kt)("span",{parentName:"p",className:"math math-inline"},(0,s.kt)("span",{parentName:"span",className:"katex"},(0,s.kt)("span",{parentName:"span",className:"katex-mathml"},(0,s.kt)("math",{parentName:"span",xmlns:"http://www.w3.org/1998/Math/MathML"},(0,s.kt)("semantics",{parentName:"math"},(0,s.kt)("mrow",{parentName:"semantics"},(0,s.kt)("msub",{parentName:"mrow"},(0,s.kt)("mi",{parentName:"msub"},"P"),(0,s.kt)("mi",{parentName:"msub"},"b"))),(0,s.kt)("annotation",{parentName:"semantics",encoding:"application/x-tex"},"P_b")))),(0,s.kt)("span",{parentName:"span",className:"katex-html","aria-hidden":"true"},(0,s.kt)("span",{parentName:"span",className:"base"},(0,s.kt)("span",{parentName:"span",className:"strut",style:{height:"0.8333em",verticalAlign:"-0.15em"}}),(0,s.kt)("span",{parentName:"span",className:"mord"},(0,s.kt)("span",{parentName:"span",className:"mord mathnormal",style:{marginRight:"0.13889em"}},"P"),(0,s.kt)("span",{parentName:"span",className:"msupsub"},(0,s.kt)("span",{parentName:"span",className:"vlist-t vlist-t2"},(0,s.kt)("span",{parentName:"span",className:"vlist-r"},(0,s.kt)("span",{parentName:"span",className:"vlist",style:{height:"0.3361em"}},(0,s.kt)("span",{parentName:"span",style:{top:"-2.55em",marginLeft:"-0.1389em",marginRight:"0.05em"}},(0,s.kt)("span",{parentName:"span",className:"pstrut",style:{height:"2.7em"}}),(0,s.kt)("span",{parentName:"span",className:"sizing reset-size6 size3 mtight"},(0,s.kt)("span",{parentName:"span",className:"mord mathnormal mtight"},"b")))),(0,s.kt)("span",{parentName:"span",className:"vlist-s"},"\u200b")),(0,s.kt)("span",{parentName:"span",className:"vlist-r"},(0,s.kt)("span",{parentName:"span",className:"vlist",style:{height:"0.15em"}},(0,s.kt)("span",{parentName:"span"}))))))))))," are seeking to establish a ICE session."),(0,s.kt)("p",null,"The smaller public key (PK) of the two peers takes the role of the controlling agent.\nIn this example PA has the role of the controlling agent as: ",(0,s.kt)("span",{parentName:"p",className:"math math-inline"},(0,s.kt)("span",{parentName:"span",className:"katex"},(0,s.kt)("span",{parentName:"span",className:"katex-mathml"},(0,s.kt)("math",{parentName:"span",xmlns:"http://www.w3.org/1998/Math/MathML"},(0,s.kt)("semantics",{parentName:"math"},(0,s.kt)("mrow",{parentName:"semantics"},(0,s.kt)("mi",{parentName:"mrow"},"P"),(0,s.kt)("mi",{parentName:"mrow"},"K"),(0,s.kt)("mo",{parentName:"mrow",stretchy:"false"},"("),(0,s.kt)("msub",{parentName:"mrow"},(0,s.kt)("mi",{parentName:"msub"},"P"),(0,s.kt)("mi",{parentName:"msub"},"a")),(0,s.kt)("mo",{parentName:"mrow",stretchy:"false"},")"),(0,s.kt)("mo",{parentName:"mrow"},"<"),(0,s.kt)("mi",{parentName:"mrow"},"P"),(0,s.kt)("mi",{parentName:"mrow"},"K"),(0,s.kt)("mo",{parentName:"mrow",stretchy:"false"},"("),(0,s.kt)("msub",{parentName:"mrow"},(0,s.kt)("mi",{parentName:"msub"},"P"),(0,s.kt)("mi",{parentName:"msub"},"b")),(0,s.kt)("mo",{parentName:"mrow",stretchy:"false"},")")),(0,s.kt)("annotation",{parentName:"semantics",encoding:"application/x-tex"},"PK(P_a) < PK(P_b)")))),(0,s.kt)("span",{parentName:"span",className:"katex-html","aria-hidden":"true"},(0,s.kt)("span",{parentName:"span",className:"base"},(0,s.kt)("span",{parentName:"span",className:"strut",style:{height:"1em",verticalAlign:"-0.25em"}}),(0,s.kt)("span",{parentName:"span",className:"mord mathnormal",style:{marginRight:"0.13889em"}},"P"),(0,s.kt)("span",{parentName:"span",className:"mord mathnormal",style:{marginRight:"0.07153em"}},"K"),(0,s.kt)("span",{parentName:"span",className:"mopen"},"("),(0,s.kt)("span",{parentName:"span",className:"mord"},(0,s.kt)("span",{parentName:"span",className:"mord mathnormal",style:{marginRight:"0.13889em"}},"P"),(0,s.kt)("span",{parentName:"span",className:"msupsub"},(0,s.kt)("span",{parentName:"span",className:"vlist-t vlist-t2"},(0,s.kt)("span",{parentName:"span",className:"vlist-r"},(0,s.kt)("span",{parentName:"span",className:"vlist",style:{height:"0.1514em"}},(0,s.kt)("span",{parentName:"span",style:{top:"-2.55em",marginLeft:"-0.1389em",marginRight:"0.05em"}},(0,s.kt)("span",{parentName:"span",className:"pstrut",style:{height:"2.7em"}}),(0,s.kt)("span",{parentName:"span",className:"sizing reset-size6 size3 mtight"},(0,s.kt)("span",{parentName:"span",className:"mord mathnormal mtight"},"a")))),(0,s.kt)("span",{parentName:"span",className:"vlist-s"},"\u200b")),(0,s.kt)("span",{parentName:"span",className:"vlist-r"},(0,s.kt)("span",{parentName:"span",className:"vlist",style:{height:"0.15em"}},(0,s.kt)("span",{parentName:"span"})))))),(0,s.kt)("span",{parentName:"span",className:"mclose"},")"),(0,s.kt)("span",{parentName:"span",className:"mspace",style:{marginRight:"0.2778em"}}),(0,s.kt)("span",{parentName:"span",className:"mrel"},"<"),(0,s.kt)("span",{parentName:"span",className:"mspace",style:{marginRight:"0.2778em"}})),(0,s.kt)("span",{parentName:"span",className:"base"},(0,s.kt)("span",{parentName:"span",className:"strut",style:{height:"1em",verticalAlign:"-0.25em"}}),(0,s.kt)("span",{parentName:"span",className:"mord mathnormal",style:{marginRight:"0.13889em"}},"P"),(0,s.kt)("span",{parentName:"span",className:"mord mathnormal",style:{marginRight:"0.07153em"}},"K"),(0,s.kt)("span",{parentName:"span",className:"mopen"},"("),(0,s.kt)("span",{parentName:"span",className:"mord"},(0,s.kt)("span",{parentName:"span",className:"mord mathnormal",style:{marginRight:"0.13889em"}},"P"),(0,s.kt)("span",{parentName:"span",className:"msupsub"},(0,s.kt)("span",{parentName:"span",className:"vlist-t vlist-t2"},(0,s.kt)("span",{parentName:"span",className:"vlist-r"},(0,s.kt)("span",{parentName:"span",className:"vlist",style:{height:"0.3361em"}},(0,s.kt)("span",{parentName:"span",style:{top:"-2.55em",marginLeft:"-0.1389em",marginRight:"0.05em"}},(0,s.kt)("span",{parentName:"span",className:"pstrut",style:{height:"2.7em"}}),(0,s.kt)("span",{parentName:"span",className:"sizing reset-size6 size3 mtight"},(0,s.kt)("span",{parentName:"span",className:"mord mathnormal mtight"},"b")))),(0,s.kt)("span",{parentName:"span",className:"vlist-s"},"\u200b")),(0,s.kt)("span",{parentName:"span",className:"vlist-r"},(0,s.kt)("span",{parentName:"span",className:"vlist",style:{height:"0.15em"}},(0,s.kt)("span",{parentName:"span"})))))),(0,s.kt)("span",{parentName:"span",className:"mclose"},")"))))),"."),(0,s.kt)("mermaid",{value:"sequenceDiagram\n    autonumber\n\n    actor Pa as Peer A\n    actor Pb as Peer B\n    participant b as Backend\n\n    Pa ->> b: SessionDescription(Pa -> Pb)\n    b ->> Pb: SessionDescription(Pa -> Pb)\n\n    Pb ->> b: SessionDescription(Pb -> Pa)\n    b ->> Pa: SessionDescription(Pb -> Pa)"}),(0,s.kt)("mermaid",{value:"stateDiagram-v2 \n    [*] --\x3e Unknown\n    \n    note right of Unknown\n        No agent exists\n    end note\n\n    Unknown --\x3e Idle: 1. Create new agent<br>2. Send local credentials\n\n    Idle --\x3e New: On remote credentials<br>1. Start gathering local candidates\n    Idle --\x3e Idle: Repeatedly send local credentials with back-off\n\n    New --\x3e Connecting: On remote candidate<br>1. Connect\n    Connecting --\x3e Checking\n    Checking --\x3e Connected\n    Checking --\x3e Failed\n    Completed --\x3e Disconnected\n    Connected --\x3e Disconnected\n    Connected --\x3e Completed\n    Completed --\x3e Closed\n    Disconnected --\x3e Closed\n    Closed --\x3e Idle: 1. Create new agent<br>2. Send local credentials\n    Failed --\x3e Closed"}),(0,s.kt)("h2",{id:"session-description"},"Session Description"),(0,s.kt)("p",null,"Session descriptions are exchanged by one or more the signaling backends via signaling ",(0,s.kt)("em",{parentName:"p"},"envelopes")," which contain signaling ",(0,s.kt)("em",{parentName:"p"},"messages"),".\nThe ",(0,s.kt)("em",{parentName:"p"},"envelopes")," are containers which encrypt the carried ",(0,s.kt)("em",{parentName:"p"},"message")," via asymmetric cryptography using the public key of the recipient."),(0,s.kt)("p",null,"Both the ",(0,s.kt)("em",{parentName:"p"},"envelope")," and the ",(0,s.kt)("em",{parentName:"p"},"message")," are serialized using Protobuf."),(0,s.kt)("p",null,"Checkout the ",(0,s.kt)("a",{parentName:"p",href:"https://github.com/cunicu/cunicu/blob/main/proto/signaling/signaling.proto"},(0,s.kt)("inlineCode",{parentName:"a"},"pkg/pb/signaling.proto"))," for details."),(0,s.kt)("h2",{id:"backends"},"Backends"),(0,s.kt)("p",null,"cun\u012bcu can support multiple backends for signaling session information such as session IDs, ICE candidates, public keys and STUN credentials."),(0,s.kt)("h3",{id:"available-backends"},"Available backends"),(0,s.kt)("ul",null,(0,s.kt)("li",{parentName:"ul"},"gRPC"),(0,s.kt)("li",{parentName:"ul"},"Kubernetes API server")),(0,s.kt)("p",null,"For the use within a Kubernetes cluster also a dedicated backend using the Kubernetes api-server is available.\nCheckout the ",(0,s.kt)("a",{parentName:"p",href:"https://github.com/cunicu/cunicu/blob/main/pkg/signaling/backend.go"},(0,s.kt)("inlineCode",{parentName:"a"},"Backend"))," interface for implementing your own backend."),(0,s.kt)("h3",{id:"semantics"},"Semantics"),(0,s.kt)("p",null,"A backend must:"),(0,s.kt)("ul",null,(0,s.kt)("li",{parentName:"ul"},"Must facilitate a reliable delivery ",(0,s.kt)("em",{parentName:"li"},"envelopes")," between peers using their public keys as addresses."),(0,s.kt)("li",{parentName:"ul"},"Must support delivery of ",(0,s.kt)("em",{parentName:"li"},"envelopes")," to a group of recipients (e.g. multicast)."),(0,s.kt)("li",{parentName:"ul"},"May deliver the ",(0,s.kt)("em",{parentName:"li"},"envelopes")," out-of-order."),(0,s.kt)("li",{parentName:"ul"},"May discard ",(0,s.kt)("em",{parentName:"li"},"envelopes")," if the recipient is not yet known or reachable."),(0,s.kt)("li",{parentName:"ul"},"Shall be stateless. It shall not buffer or record any ",(0,s.kt)("em",{parentName:"li"},"envelopes"),".")),(0,s.kt)("h3",{id:"interface"},"Interface"),(0,s.kt)("p",null,"All signaling backends implement the rather simple ",(0,s.kt)("a",{parentName:"p",href:"https://github.com/cunicu/cunicu/blob/main/pkg/signaling/backend.go"},(0,s.kt)("inlineCode",{parentName:"a"},"signaling.Backend")," interface"),":"),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-go"},"type Message = pb.SignalingMessage\n\ntype MessageHandler interface {\n    OnSignalingMessage(*crypto.PublicKeyPair, *Message)\n}\n\ntype Backend interface {\n    io.Closer\n\n    // Publish a signaling message to a specific peer\n    Publish(ctx context.Context, kp *crypto.KeyPair, msg *Message) error\n\n    // Subscribe to messages send by a specific peer\n    Subscribe(ctx context.Context, kp *crypto.KeyPair, h MessageHandler) (bool, error)\n\n    // Unsubscribe from messages send by a specific peer\n    Unsubscribe(ctx context.Context, kp *crypto.KeyPair, h MessageHandler) (bool, error)\n\n    // Returns the backends type identifier\n    Type() signalingproto.BackendType\n}\n")))}N.isMDXComponent=!0}}]);