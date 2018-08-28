module.exports = {
  title: "Cosmos Network",
  description: "Documentation for the Cosmos Network.",
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
          "/introduction/tendermint",
        ]
      },
      {
        title: "Getting Started",
        collapsable: false,
        children: [
          "/getting-started/voyager",
          "/getting-started/installation",
          "/getting-started/full-node",
          "/getting-started/create-testnet"
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
        title: "Lotion JS",
        collapsable: false,
        children: [["/lotion/overview", "Overview"], "/lotion/building-an-app"]
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
