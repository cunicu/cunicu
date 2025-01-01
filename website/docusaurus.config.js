// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

import {themes as prismThemes} from 'prism-react-renderer';
import remarkMath from "remark-math";
import rehypeKatex from "rehype-katex";

export default {
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
    require.resolve('docusaurus-lunr-search'),
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
  // If you aren't using GitHub pages, you don't need these.
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
      "@docusaurus/preset-classic",
      ({
        docs: {
          sidebarPath: require.resolve("./sidebars.js"),
          editUrl:
            "https://github.com/cunicu/cunicu/edit/main/",
          remarkPlugins: [
            remarkMath
          ],
          rehypePlugins: [
            rehypeKatex
          ],
            exclude: ["/docs/usage/man/**"],
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
    {
      mermaid: {
        theme: {
          light: 'neutral',
          dark: 'forest'
        },
      },
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
            to: "/blog",
            label: "📰 Blog",
            position: "left"
          },
          {
            to: "https://discuss.cunicu.li",
            label: "👋 Community",
            position: "left"
          },
          {
            type: "doc",
            docId: "index",
            position: "left",
            label: "📚 Documentation",
          },
          {
            href: "https://github.com/cunicu",
            position: "right",
            className: "header-github-link",
            "aria-label": "GitHub repository",
          },
          {
            href: "https://codeberg.org/cunicu",
            position: "right",
            className: "header-codeberg-link",
            "aria-label": "Codeberg repository",
          },
        ],
      },
      footer: {
        style: "dark",
        links: [
          {
            title: "Documentation",
            items: [
              {
                label: "Tutorial",
                to: "/docs/",
              },
              {
                label: "Legal",
                to: "/docs/legal/",
              },
            ],
          },
          {
            title: "Community",
            items: [
              {
                label: "Forum",
                href: "https://discuss.cunicu.li",
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
                label: "Contact",
                href: "/docs/contact",
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} Steffen Vogel.`,
      },

      prism: {
        theme: prismThemes.github,
        darkTheme: prismThemes.dracula,
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
    },

  themes: ["@docusaurus/theme-mermaid"],
  markdown: {
    mermaid: true,
  }
};
