module.exports = {
  theme: "cosmos",
  title: "Cosmos SDK",
  markdown: {
    anchor: {
      permalinkSymbol: ""
    }
  },
  head: [
    [
      "link",
      {
        rel: "stylesheet",
        type: "text/css",
        href: "https://cloud.typography.com/6138116/7255612/css/fonts.css"
      }
    ],
  ],
  locales: {
    "/": {
      lang: "en-US"
    },
    kr: {
      lang: "kr"
    },
    kr: {
      lang: "kr"
    },
    cn: {
      lang: "cn"
    },
    ru: {
      lang: "ru"
    }
  },
  base: process.env.VUEPRESS_BASE || "/",
  themeConfig: {
    repo: "cosmos/cosmos-sdk",
    docsRepo: "cosmos/cosmos-sdk",
    docsDir: "docs",
    editLinks: true,
    label: "sdk",
    sidebar: [
      {
        title: "Using the SDK",
        children: [
          {
            title: "Modules",
            directory: true,
            path: "/modules"
          }
        ]
      },
      {
        title: "Resources",
        children: [
          {
            title: "Tutorials",
            path: "https://tutorials.cosmos.network"
          },
          {
            title: "SDK API Reference",
            path: "https://godoc.org/github.com/cosmos/cosmos-sdk"
          },
          {
            title: "REST API Spec",
            path: "https://cosmos.network/rpc/"
          }
        ]
      }
    ],
    gutter: {
      title: "Help & Support",
      editLink: true,
      chat: {
        title: "Riot Chat",
        text: "Chat with Cosmos developers on Riot Chat.",
        url: "https://riot.im/app/#/room/#cosmos-sdk:matrix.org",
        bg: "linear-gradient(225.11deg, #2E3148 0%, #161931 95.68%)"
      },
      forum: {
        title: "Cosmos SDK Forum",
        text: "Join the SDK Developer Forum to learn more.",
        url: "https://forum.cosmos.network/",
        bg: "linear-gradient(225deg, #46509F -1.08%, #2F3564 95.88%)",
        logo: "cosmos"
      },
      github: {
        title: "Found an Issue?",
        text: "Help us improve this page by suggesting edits on GitHub."
      }
    },
    footer: {
      logo: "/logo-bw.svg",
      textLink: {
        text: "cosmos.network",
        url: "/"
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
        "This website is maintained by Tendermint Inc. The contents and opinions of this website are those of Tendermint Inc.",
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
            },
            {
              title: "Tendermint Core",
              url: "https://docs.tendermint.com/"
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
              url:
                "https://github.com/cosmos/cosmos-sdk/blob/master/docs/DOCS_README.md"
            },
            {
              title: "Source code on GitHub",
              url: "https://github.com/cosmos/cosmos-sdk/"
            }
          ]
        }
      ]
    }
  },
  plugins: [
    [
      "@vuepress/google-analytics",
      {
        ga: "UA-51029217-12"
      }
    ],
    [
      "sitemap",
      {
        hostname: "https://docs.cosmos.network"
      }
    ]
  ]
};
