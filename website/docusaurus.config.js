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
      {
        
      }
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
  organizationName: "stv0g", // Usually your GitHub org/user name.
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
      "classic",
      /** @type {import("@docusaurus/preset-classic").Options} */
      ({
        docs: {
          sidebarPath: require.resolve("./sidebars.js"),
          editUrl:
            "https://github.com/stv0g/cunicu/edit/master/",
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
            "https://github.com/stv0g/cunicu/master/website",
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
            href: "https://github.com/stv0g/cunicu",
            label: "GitHub",
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
                href: "https://gophers.slack.com/archives/C036CTEGJFN",
              },
              {
                label: "Twitter",
                href: "https://twitter.com/cunicuVPN",
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
                label: "GitHub",
                href: "https://github.com/stv0g/cunicu",
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
    }),

  markdown: {
    mermaid: true,
  },
  themes: ['@docusaurus/theme-mermaid'],
};
