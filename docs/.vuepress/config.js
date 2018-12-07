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
    lastUpdated: "Last Updated",
    nav: [{ text: "Back to Cosmos", link: "https://cosmos.network" }],
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
          "/tutorial/app-complete",
          "/tutorial/entrypoint",
          "/tutorial/dep",
          "/tutorial/build-run"
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
