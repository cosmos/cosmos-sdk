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
  }
};
