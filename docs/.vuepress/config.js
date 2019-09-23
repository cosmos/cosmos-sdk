const glob = require("glob");
const markdownIt = require("markdown-it");
const meta = require("markdown-it-meta");
const ascii = require("./markdown-it-ascii.js");
const fs = require("fs");
const _ = require("lodash");
var path = require("path");

module.exports = {
  title: "Cosmos SDK",
  base: process.env.VUEPRESS_BASE || "/",
  plugins: [
    [
      "@vuepress/search",
      {
        searchMaxSuggestions: 10
      }
    ]
  ],
  markdown: {
    anchor: {
      permalinkSymbol: ""
    }
    // extendMarkdown: md => {
    //   md.use(ascii);
    // }
  },
  locales: {
    "/": {
      lang: "en-US"
    },
    "/ru/": {
      lang: "ru"
    },
    "/kr/": {
      lang: "kr"
    },
    "/cn/": {
      lang: "cn"
    }
  },
  themeConfig: {
    repo: "cosmos/cosmos-sdk",
    docsDir: "docs",
    editLinks: true,
    sidebar: [
      {
        title: "Using the SDK",
        children: [
          {
            title: "Scaffolding",
            children: [
              {
                title: "Scaffolding 1",
                path: "/scaffolding1"
              },
              {
                title: "Scaffolding 2",
                path: "/scaffolding2"
              }
            ]
          },
          {
            title: "Modules",
            children: [
              {
                title: "Auth",
                path: "/auth",
              },
              {
                title: "Bank",
                path: "/bank"
              }
            ]
          }
        ]
      },
      {
        title: "Resources",
        children: [
          {
            title: "Tutorials",
            path: "https://github.com/cosmos/sdk-application-tutorial"
          },
          {
            title: "SDK API Reference",
            path: "https://github.com/cosmos/sdk-application-tutorial"
          },
          {
            title: "REST API Spec",
            path: "https://github.com/cosmos/sdk-application-tutorial"
          },
        ]
      }
    ]
  }
};
