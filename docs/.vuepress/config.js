module.exports = {
  theme: "cosmos",
  title: "Cosmos SDK",
  markdown: {
    anchor: {
      permalinkSymbol: ""
    }
  },
  base: process.env.VUEPRESS_BASE || "/",
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
    logo: "/logo.svg",
    sidebar: [
      // {
      //   title: "Using the SDK",
      //   children: [
      //     {
      //       title: "Scaffolding",
      //       children: [
      //         {
      //           title: "Scaffolding 1",
      //           path: "/scaffolding1"
      //         },
      //         {
      //           title: "Scaffolding 2",
      //           path: "/scaffolding2"
      //         }
      //       ]
      //     },
      //     {
      //       title: "Modules",
      //       children: [
      //         {
      //           title: "Auth",
      //           path: "/auth",
      //         },
      //         {
      //           title: "Bank",
      //           path: "/bank"
      //         }
      //       ]
      //     }
      //   ]
      // },
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
          }
        ]
      }
    ]
  }
};
