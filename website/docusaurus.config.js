// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const prism = require("prism-react-renderer");
const math = require("remark-math");
const katex = require("rehype-katex");

/** @type {import("@docusaurus/types").Config} */
module.exports = {
  title: "cunīcu",
  tagline: "zeroconf • p2p • mesh • vpn",
  url: "https://cunicu.li",
  baseUrl: "/",
  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",
  favicon: "img/favicon.png",
  trailingSlash: false,

  stylesheets: [
    {
      href: "https://cdn.jsdelivr.net/npm/katex@0.13.24/dist/katex.min.css",
      type: "text/css",
      integrity:
        "sha384-odtC+0UGzzFL/6PNoE8rX/SPcQDXBJ+uRepguP4QkPCm2LBxH3FA3y+fKSiJ+AmM",
      crossorigin: "anonymous",
    },
  ],

  plugins: [
    [
      require.resolve("@cmfcmf/docusaurus-search-local"),
      { }
    ],
    [
      require.resolve("@gabrielcsapo/docusaurus-plugin-matomo"),
      {
        siteId: "5",
        matomoUrl: "https://matomo.0l.de",
        siteUrl: "https://cunicu.li",
      }
    ],
    function (context, options) {
      return {
        name: "webpack-configuration-plugin",
        configureWebpack(config, isServer, utils) {
          return {
            resolve: {
              symlinks: false,
            }
          };
        }
      };
    },
  ],

  // GitHub pages deployment config.
  // If you aren"t using GitHub pages, you don't need these.
  organizationName: "cunicu", // Usually your GitHub org/user name.
  projectName: "cunicu", // Usually your repo name.

  // Even if you don"t use internalization, you can use this field to set useful
  // metadata like html lang. For example, if your site is Chinese, you may want
  // to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  presets: [
    [
      "redocusaurus",
      {
        // Plugin Options for loading OpenAPI files
        specs: [
          {
            spec: "openapi/openapi.yaml",
          },
        ],
        theme: {
          primaryColor: '#d4aa01',
          options: {
            showObjectSchemaExamples: true,
            expandResponses: "all",
            schemaExpansionLevel: "all",
          }
        },
      },
    ],
    [
      "classic",
      /** @type {import("@docusaurus/preset-classic").Options} */
      ({
        docs: {
          sidebarPath: require.resolve("./sidebars.js"),
          editUrl:
            "https://github.com/cunicu/cunicu/edit/main/",
          remarkPlugins: [
            math
          ],
          rehypePlugins: [
            katex
          ],
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            "https://github.com/cunicu/cunicu/main/website",
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import("@docusaurus/preset-classic").ThemeConfig} */
    ({
      colorMode: {
        disableSwitch: true
      },
      navbar: {
        title: "cunīcu",
        logo: {
          alt: "cunīcu logo",
          src: "img/cunicu_icon.svg",
        },
        items: [
          {
            type: "doc",
            docId: "index",
            position: "left",
            label: "Docs",
          },
          {
            to: "/blog",
            label: "Blog",
            position: "left"
          },
          {
            href: "https://github.com/cunicu",
            label: "Sourcecode",
            position: "right",
          },
        ],
      },
      footer: {
        style: "dark",
        links: [
          {
            title: "Docs",
            items: [
              {
                label: "Tutorial",
                to: "/docs/",
              },
            ],
          },
          {
            title: "Community",
            items: [
              {
                label: "Slack",
                href: "https://join.slack.com/t/gophers/shared_invite/zt-1447h1rgj-s9W5BcyRzBxUwNAZJUKmaQ)",
              },
              {
                label: "Fediverse",
                href: "https://fosstodon.org/@cunicu",
              },
            ],
          },
          {
            title: "More",
            items: [
              {
                label: "Blog",
                to: "/blog",
              },
              {
                label: "Sourcecode",
                href: "https://github.com/cunicu",
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} Steffen Vogel.`,
      },
      prism: {
        theme: prism.themes.github,
        darkTheme: prism.themes.dracula,
      },

      metadata: [
        {
          name: "keywords",
          content: "go, golang, iot, networking, nat-traversal, vpn, vpn-manager, mesh, ice, multi-agent-systems, wireguard, edge-cloud, wireguard-vpn"
        },
        {
          name: "description",
          content: "A zeroconf peer-to-peer mesh VPN using Wireguard® and Interactive Connectivity Establishment (ICE)"
        },
        {
          name: "twitter:creator",
          content: "@stv0g"
        }
      ]
    }),

  markdown: {
    mermaid: true,
  },
  themes: ['@docusaurus/theme-mermaid'],

  headTags: [
    {
      tagName: "meta",
      attributes: {
        
      }
    }
  ],
};
