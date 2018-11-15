module.exports = {
  title: "Cosmos Documentation",
  description: "Documentation for the Cosmos Network.",
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
          "/gaia/networks",
          "/gaia/validators/overview",
          "/gaia/validators/security",
          "/gaia/validators/validator-faq",
          "/gaia/validators/validator-setup",
          "/gaia/ledger"
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
