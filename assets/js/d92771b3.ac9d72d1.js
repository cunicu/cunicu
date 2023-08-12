(self.webpackChunkcunicu=self.webpackChunkcunicu||[]).push([[7189],{37977:(e,a,t)=>{"use strict";t.r(a),t.d(a,{assets:()=>r,contentTitle:()=>c,default:()=>p,frontMatter:()=>o,metadata:()=>l,toc:()=>u});var n=t(87462),i=(t(67294),t(3905)),s=t(68363);const o={sidebar_position:4},c="JSON Schema",l={unversionedId:"config/schema",id:"config/schema",title:"JSON Schema",description:"JSON Schema is a declarative language that allows you to annotate and validate JSON documents.",source:"@site/docs/config/schema.md",sourceDirName:"config",slug:"/config/schema",permalink:"/docs/config/schema",draft:!1,editUrl:"https://github.com/stv0g/cunicu/edit/master/docs/config/schema.md",tags:[],version:"current",sidebarPosition:4,frontMatter:{sidebar_position:4},sidebar:"tutorialSidebar",previous:{title:"Advanced Example",permalink:"/docs/config/example-advanced"},next:{title:"Design",permalink:"/docs/design"}},r={},u=[{value:"Editor / Language Server support",id:"editor--language-server-support",level:2},{value:"Reference",id:"reference",level:2}],d={toc:u},m="wrapper";function p(e){let{components:a,...t}=e;return(0,i.kt)(m,(0,n.Z)({},d,t,{components:a,mdxType:"MDXLayout"}),(0,i.kt)("h1",{id:"json-schema"},"JSON Schema"),(0,i.kt)("p",null,(0,i.kt)("a",{parentName:"p",href:"https://json-schema.org/"},"JSON Schema")," is a declarative language that allows you to annotate and validate JSON documents."),(0,i.kt)("p",null,"JSON Schema can also be used to validate YAML documents and as such cun\u012bcu's configuration file.\nYAML Ain't Markup Language (YAML) is a powerful data serialization language that aims to be human friendly."),(0,i.kt)("p",null,"Most JSON is syntactically valid YAML, but idiomatic YAML follows very different conventions.\nWhile YAML has advanced features that cannot be directly mapped to JSON, most YAML files use features that can be validated by JSON Schema.\nJSON Schema is the most portable and broadly supported choice for YAML validation."),(0,i.kt)("p",null,"The schema of cun\u012bcu's configuration file is available at:"),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://github.com/stv0g/cunicu/blob/master/etc/cunicu.schema.yaml"},(0,i.kt)("inlineCode",{parentName:"a"},"etc/cunicu.schema.yaml"))),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"https://cunicu.li/schemas/config.yaml"},"https://cunicu.li/schemas/config.yaml"))),(0,i.kt)("h2",{id:"editor--language-server-support"},"Editor / Language Server support"),(0,i.kt)("p",null,"Redhat's ",(0,i.kt)("a",{parentName:"p",href:"https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml"},"YAML Visual Studio Code extension")," provides comprehensive YAML Language support, via the ",(0,i.kt)("a",{parentName:"p",href:"https://github.com/redhat-developer/yaml-language-server"},"yaml-language-server"),"."),(0,i.kt)("p",null,"It provides completion, validation and code lenses based on JSON Schemas."),(0,i.kt)("p",null,"To make use of it, you need to associate your config file with the JSON Schema by adding the following line into your config:"),(0,i.kt)("pre",null,(0,i.kt)("code",{parentName:"pre",className:"language-yaml"},"# yaml-language-server: $schema=https://cunicu.li/schemas/config.yaml\n---\n\nwatch_interval: 1s\n")),(0,i.kt)("h2",{id:"reference"},"Reference"),(0,i.kt)("p",null,"Here is a rendered reference based on this schema:"),(0,i.kt)(s.Z,{pointer:"#/components/schemas/Config",mdxType:"ApiSchema"}))}p.isMDXComponent=!0},26242:()=>{},11314:()=>{},67251:()=>{},99018:()=>{},43044:()=>{},3408:()=>{},35126:()=>{}}]);