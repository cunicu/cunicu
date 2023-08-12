(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[6435],{47326:(e,t,n)=>{"use strict";n.r(t),n.d(t,{assets:()=>u,contentTitle:()=>o,default:()=>h,frontMatter:()=>a,metadata:()=>l,toc:()=>c});var i=n(87462),r=(n(67294),n(3905)),s=n(68363);const a={},o="Route Synchronization",l={unversionedId:"features/rtsync",id:"features/rtsync",title:"Route Synchronization",description:"The route synchronization feature keeps the kernel routing table in sync with WireGuard's AllowedIPs setting.",source:"@site/docs/features/rtsync.md",sourceDirName:"features",slug:"/features/rtsync",permalink:"/docs/features/rtsync",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/features/rtsync.md",tags:[],version:"current",frontMatter:{},sidebar:"tutorialSidebar",previous:{title:"Pre-shared Key Establishment",permalink:"/docs/features/pske"},next:{title:"Usage",permalink:"/docs/usage/"}},u={},c=[{value:"Configuration",id:"configuration",level:2}],d={toc:c},p="wrapper";function h(e){let{components:t,...n}=e;return(0,r.kt)(p,(0,i.Z)({},d,n,{components:t,mdxType:"MDXLayout"}),(0,r.kt)("h1",{id:"route-synchronization"},"Route Synchronization"),(0,r.kt)("p",null,"The route synchronization feature keeps the kernel routing table in sync with WireGuard's ",(0,r.kt)("em",{parentName:"p"},"AllowedIPs")," setting."),(0,r.kt)("p",null,"This synchronization is bi-directional:"),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},"Networks with are found in a Peers AllowedIP list will be installed as a kernel route."),(0,r.kt)("li",{parentName:"ul"},"Kernel routes with the peers link-local IP address as next-hop will be added to the Peers ",(0,r.kt)("em",{parentName:"li"},"AllowedIPs")," list.")),(0,r.kt)("p",null,"This rather simple feature allows user to pair cunicu with a software routing daemon like ",(0,r.kt)("a",{parentName:"p",href:"https://bird.network.cz/"},"Bird2")," while using a single WireGuard interface with multiple peer-to-peer links."),(0,r.kt)("h2",{id:"configuration"},"Configuration"),(0,r.kt)("p",null,"The following settings can be used in the main section of the ",(0,r.kt)("a",{parentName:"p",href:"../config/"},"configuration file")," or with-in the ",(0,r.kt)("inlineCode",{parentName:"p"},"interfaces")," section to customize settings of an individual interface."),(0,r.kt)(s.Z,{pointer:"#/components/schemas/RouteSyncSettings",mdxType:"ApiSchema"}))}h.isMDXComponent=!0},26242:()=>{},11314:()=>{},67251:()=>{},99018:()=>{},43044:()=>{},3408:()=>{},35126:()=>{}}]);