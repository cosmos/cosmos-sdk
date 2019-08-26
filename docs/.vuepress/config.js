const glob = require("glob");
const markdownIt = require("markdown-it");
const meta = require("markdown-it-meta");
const fs = require("fs");
const _ = require("lodash");

const sidebar = (directory, array) => {
  return array.map(i => {
    const children = _.sortBy(
      glob
        .sync(`./${directory}/${i[1]}/*.md`)
        .map(path => {
          const md = new markdownIt();
          const file = fs.readFileSync(path, "utf8");
          md.use(meta);
          md.render(file);
          const order = md.meta.order;
          return { path, order };
        })
        .filter(f => f.order !== false),
      ["order", "path"]
    )
      .map(f => f.path)
      .filter(f => !f.match("readme"));
    return {
      title: i[0],
      children
    };
  });
};

module.exports = {
  title: "Cosmos SDK",
  base: process.env.VUEPRESS_BASE || "/",
  locales: {
    "/": {
      lang: "en-US"
    },
    "/ru/": {
      lang: "ru"
    },
  },
  themeConfig: {
    repo: "cosmos/cosmos-sdk",
    docsDir: "docs",
    editLinks: true,
    docsBranch: "master",
    locales: {
      "/": {
        label: "English",
        sidebar: sidebar("", [
          ["Intro", "intro"],
          ["Basics", "basics"],
          ["SDK Core", "core"],
          ["About Modules", "modules"],
          ["Using the SDK", "sdk"],
          ["Interfaces", "interfaces"]
        ])
      },
      "/ru/": {
        label: "Русский",
        sidebar: sidebar("ru", [
          ["Введение", "intro"],
          ["Основы", "basics"],
          ["SDK Core", "core"],
          ["Модули", "modules"],
          ["Используем SDK", "sdk"],
          ["Интерфейсы", "interfaces"]
        ])
      },
      '/kr/': {
        label: '한국어',
        sidebar: sidebar('kr', [
          ['소개', 'intro'],
          ['기초', 'basics'],
          ['SDK Core', 'core'],
          ['모듈들', 'modules'],
          ['프로그램 사용', 'sdk'],
          ['인터페이스', 'interfaces'],
        ]),
      },
      '/cn/': {
        label: '中文',
        sidebar: sidebar('cn', [
          ['介绍', 'intro'],
          ['基本', 'basics'],
          ['SDK Core', 'core'],
          ['模块', 'modules'],
          ['使用该程序', 'sdk'],
          ['接口', 'interfaces'],
        ]),
      },
    }
  }
};
