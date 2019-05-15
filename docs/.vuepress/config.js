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
      { text: "RPC", link: "https://cosmos.network/rpc/" }
    ],
    sidebar: [
      {
        title: "Overview",
        collapsable: true,
        children: [
          "/intro/",
          "/intro/sdk-design",
          "/intro/ocap"
        ]
      },
      {
        title: "Cosmos Hub",
        collapsable: true,
        children: [
          "/cosmos-hub/what-is-gaia",
          "/cosmos-hub/installation",
          "/cosmos-hub/join-mainnet",
          "/cosmos-hub/validators/validator-setup",
          "/cosmos-hub/validators/overview",
          "/cosmos-hub/validators/security",
          "/cosmos-hub/validators/validator-faq",
          "/cosmos-hub/delegator-guide-cli",
          "/cosmos-hub/genesis",
          "/cosmos-hub/hd-wallets",
          "/cosmos-hub/ledger",
          "/cosmos-hub/gaiacli",
          "/cosmos-hub/join-testnet",
          "/cosmos-hub/deploy-testnet"
        ]
      },
      {
        title: "Tutorial",
        collapsable: true,
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
        collapsable: true,
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
