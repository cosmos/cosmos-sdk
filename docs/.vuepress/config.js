module.exports = {
  theme: "cosmos",
  title: "Cosmos SDK",
  markdown: {
    anchor: {
      permalinkSymbol: ""
    }
  },
  base: process.env.VUEPRESS_BASE || "/",
  themeConfig: {
    repo: "cosmos/cosmos-sdk",
    docsDir: "docs",
    editLinks: true,
    logo: "/logo.svg",
    label: "sdk",
    sidebar: [
      {
        title: "Resources",
        children: [
          {
            title: "Modules",
            directory: true,
            path: "/modules/"
          },
          {
            title: "Tutorials",
            path: "https://github.com/cosmos/sdk-application-tutorial"
          }
        ]
      }
    ],
    gutter: {
      title: "Help & Support",
      editLink: true,
      children: [
        {
          title: "Riot Chat",
          text: "Chat with Tendermint developers on Riot Chat.",
          highlighted: "500+ people chatting now"
        },
        {
          title: "Cosmos SDK Forum",
          text: "Found an Issue?",
          highlighted:
            "Help us improve this page by suggesting edits on GitHub."
        }
      ]
    },
    footer: {
      logo: "/logo-bw.svg",
      textLink: {
        text: "cosmos.network",
        url: "https://cosmos.network"
      },
      services: [
        {
          service: "medium",
          url: "https://blog.cosmos.network/"
        },
        {
          service: "twitter",
          url: "https://twitter.com/cosmos"
        },
        {
          service: "linkedin",
          url: "https://www.linkedin.com/company/tendermint/"
        },
        {
          service: "reddit",
          url: "https://reddit.com/r/cosmosnetwork"
        },
        {
          service: "telegram",
          url: "https://t.me/cosmosproject"
        },
        {
          service: "youtube",
          url: "https://www.youtube.com/c/CosmosProject"
        }
      ],
      smallprint:
        "The development of the Cosmos project is led primarily by Tendermint Inc., the for-profit entity which also maintains this website. Funding for this development comes primarily from the Interchain Foundation, a Swiss non-profit.",
      links: [
        {
          title: "Documentation",
          children: [
            {
              title: "Cosmos SDK",
              url: "https://cosmos.network/docs"
            },
            {
              title: "Cosmos Hub",
              url: "https://hub.cosmos.network/"
            }
          ]
        },
        {
          title: "Community",
          children: [
            {
              title: "Cosmos blog",
              url: "https://blog.cosmos.network/"
            },
            {
              title: "Forum",
              url: "https://forum.cosmos.network/"
            },
            {
              title: "Chat",
              url: "https://riot.im/app/#/room/#cosmos-sdk:matrix.org"
            }
          ]
        },
        {
          title: "Contributing",
          children: [
            {
              title: "Contributing to the docs",
              url: "https://github.com/cosmos/cosmos-sdk/tree/master/docs"
            },
            {
              title: "Source code on GitHub",
              url: "https://github.com/cosmos/cosmos-sdk/"
            }
          ]
        }
      ]
    }
  }
};
