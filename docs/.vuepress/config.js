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
        title: "Introduction",
        collapsable: false,
        children: [
          "/introduction/cosmos-hub",
          "/introduction/tendermint-cosmos",
          "/introduction/tendermint"
        ]
      },
      {
        title: "Getting Started",
        collapsable: false,
        children: [
          "/getting-started/voyager",
          "/getting-started/installation",
          "/getting-started/join-testnet",
          "/getting-started/networks"
        ]
      },
      {
        title: "Cosmos SDK",
        collapsable: false,
        children: [
          ["/sdk/overview", "Overview"],
          ["/sdk/core/intro", "Core"],
          "/sdk/core/app1",
          "/sdk/core/app2",
          "/sdk/core/app3",
          "/sdk/core/app4",
          "/sdk/core/app5",
          // "/sdk/modules",
          "/sdk/clients"
        ]
      },
      // {
      //   title: "Specifications",
      //   collapsable: false,
      //   children: [
      //     ["/specs/overview", "Overview"],
      //     "/specs/governance",
      //     "/specs/ibc",
      //     "/specs/staking",
      //     "/specs/icts",
      //   ]
      // },
      {
        title: "SDK by Examples - Simple Governance",
        collapsable: false,
        children: [
          ["/sdk/sdk-by-examples/simple-governance/intro", "Intro"],
          "/sdk/sdk-by-examples/simple-governance/setup-and-design",
          "/sdk/sdk-by-examples/simple-governance/app-init",
          "/sdk/sdk-by-examples/simple-governance/simple-gov-module",
          "/sdk/sdk-by-examples/simple-governance/bridging-it-all",
          "/sdk/sdk-by-examples/simple-governance/running-the-application"
        ]
      },
      {
        title: "Light Client",
        collapsable: false,
        children: [
	  "/light/",
	  "/light/getting_started"
	]
      },
      {
        title: "Lotion JS",
        collapsable: false,
        children: [
	  ["/lotion/overview", "Overview"]
	]
      },
      {
        title: "Validators",
        collapsable: false,
        children: [
          ["/validators/overview", "Overview"],
          ["/validators/security", "Security"],
          ["/validators/validator-setup", "Validator Setup"],
          "/validators/validator-faq"
        ]
      },
      {
        title: "Clients",
        collapsable: false,
        children: [
          ["/clients/service-providers", "Service Providers"]
        ]
      },
      {
        title: "Resources",
        collapsable: false,
        children: [
          // ["/resources/faq" "General"],
          "/resources/delegator-faq",
          ["/resources/whitepaper", "Whitepaper - English"],
          ["/resources/whitepaper-ko", "Whitepaper - 한국어"],
          ["/resources/whitepaper-zh-CN", "Whitepaper - 中文"],
          ["/resources/whitepaper-pt", "Whitepaper - Português"]
        ]
      }
    ]
  }
}
