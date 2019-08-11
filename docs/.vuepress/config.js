module.exports = {
  title: "Cosmos SDK Documentation",
  description: "Documentation for the Cosmos SDK and Gaia.",
  ga: "UA-51029217-2",
  dest: "./dist/docs",
  base: "/docs/",
  markdown: {
    lineNumbers: true
  },
  themeConfig: {
    repo: "cosmos/cosmos-sdk",
    editLinks: true,
    docsDir: "docs",
    docsBranch: "master",
    editLinkText: "Edit this page on Github",
    lastUpdated: true,
    algolia: {
      apiKey: "a6e2f64347bb826b732e118c1366819a",
      indexName: "cosmos_network",
      debug: false
    },
    nav: [
      { text: "Back to Cosmos", link: "https://cosmos.network" },
      { text: "RPC", link: "https://cosmos.network/rpc/" }
    ],
    sidebar: [
      {
        title: "Overview",
        collapsable: true,
        children: [
          "/intro/",
          "/intro/why-app-specific",
          "/intro/sdk-app-architecture",
          "/intro/sdk-design"
        ]
      },
      {
        title: "Tutorial",
        collapsable: true,
        children: [
          "/tutorial/",
          "/tutorial/app-design",
          "/tutorial/app-init",
          "/tutorial/types",
          "/tutorial/key",
          "/tutorial/keeper",
          "/tutorial/msgs-handlers",
          "/tutorial/set-name",
          "/tutorial/buy-name",
          "/tutorial/queriers",
          "/tutorial/alias",
          "/tutorial/codec",
          "/tutorial/cli",
          "/tutorial/rest",
          "/tutorial/module",
          "/tutorial/genesis",
          "/tutorial/app-complete",
          "/tutorial/entrypoint",
          "/tutorial/gomod",
          "/tutorial/build-run",
          "/tutorial/run-rest"
        ]
      },
      {
        title: "Interfaces",
        collapsable: true,
        children: [
          "/interfaces/",
          "/interfaces/cli",
          "/interfaces/service-providers",
          "/interfaces/lite/", // this renders the readme
          "/interfaces/lite/getting_started",
          "/interfaces/lite/specification"
        ]
      }
    ]
  }
};
