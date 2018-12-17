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
    docsBranch: "develop",
    editLinkText: 'Edit this page on Github',
    lastUpdated: true,
    algolia: {
      apiKey: 'a6e2f64347bb826b732e118c1366819a',
      indexName: 'cosmos_network',
      debug: false
    },
    nav: [
      { text: "Back to Cosmos", link: "https://cosmos.network" },
      { text: "RPC", link: "../rpc/" }
    ],
    sidebar: [
      {
        title: "Overview",
        collapsable: false,
        children: [
          "/intro/",
          "/intro/sdk-app-architecture",
          "/intro/ocap"
        ]
      },
      {
        title: "Gaia",
        collapsable: false,
        children: [
          "/gaia/installation",
          "/gaia/join-testnet",
          "/gaia/validators/validator-setup",
          "/gaia/validators/overview",
          "/gaia/validators/security",
          "/gaia/validators/validator-faq",
          "/gaia/deploy-testnet",
          "/gaia/ledger",
          "/gaia/gaiacli"
        ]
      },
      {
        title: "Tutorial",
        collapsable: false,
        children: [
          "/tutorial/",
          "/tutorial/app-design",
          "/tutorial/app-init",
          "/tutorial/keeper",
          "/tutorial/msgs-handlers",
          "/tutorial/set-name",
          "/tutorial/buy-name",
          "/tutorial/queriers",
          "/tutorial/codec",
          "/tutorial/cli",
          "/tutorial/rest",
          "/tutorial/app-complete",
          "/tutorial/entrypoint",
          "/tutorial/dep",
          "/tutorial/build-run",
          "/tutorial/run-rest"
        ]
      },
      {
        title: "Clients",
        collapsable: false,
        children: [
      	  "/clients/",
          "/clients/cli",
          "/clients/service-providers",
      	  "/clients/lite/", // this renders the readme
      	  "/clients/lite/getting_started",
      	  "/clients/lite/specification"
      	]
      }
    ]
  }
}
