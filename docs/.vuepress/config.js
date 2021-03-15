module.exports = {
  theme: "cosmos",
  title: "Cosmos SDK",
  locales: {
    "/": {
      lang: "en-US"
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
  head: [
    ['link', { rel: "apple-touch-icon", sizes: "180x180", href: "/apple-touch-icon.png" }],
    ['link', { rel: "icon", type: "image/png", sizes: "32x32", href: "/favicon-32x32.png" }],
    ['link', { rel: "icon", type: "image/png", sizes: "16x16", href: "/favicon-16x16.png" }],
    ['link', { rel: "manifest", href: "/site.webmanifest" }],
    ['meta', { name: "msapplication-TileColor", content: "#2e3148" }],
    ['meta', { name: "theme-color", content: "#ffffff" }],
    ['link', { rel: "icon", type: "image/svg+xml", href: "/favicon-svg.svg" }],
    ['link', { rel: "apple-touch-icon-precomposed", href: "/apple-touch-icon-precomposed.png" }],
  ],
  themeConfig: {
    repo: "cosmos/cosmos-sdk",
    docsRepo: "cosmos/cosmos-sdk",
    docsDir: "docs",
    editLinks: true,
    label: "sdk",
    algolia: {
      id: "BH4D9OD16A",
      key: "ac317234e6a42074175369b2f42e9754",
      index: "cosmos-sdk"
    },
    versions: [
      {
        "label": "v0.39",
        "key": "v0.39"
      },
      {
        "label": "v0.42",
        "key": "v0.42"
      },
      {
        "label": "master",
        "key": "master"
      }
    ],
    topbar: {
      banner: true
    },
    sidebar: { 
      auto: true,
      nav: [
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
      ]
    },
    gutter: {
      title: "Help & Support",
      editLink: true,
      chat: {
        title: "Discord",
        text: "Chat with Cosmos developers on Discord.",
        url: "https://discordapp.com/channels/669268347736686612",
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
      question: {
        text: "Chat with Cosmos developers in <a href='https://discord.gg/W8trcGV' target='_blank'>Discord</a> or reach out on the <a href='https://forum.cosmos.network/c/tendermint' target='_blank'>SDK Developer Forum</a> to learn more."
      },
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
        "This website is maintained by Tendermint Inc. The contents and opinions of this website are those of Tendermint Inc.",
      links: [
        {
          title: "Documentation",
          children: [
            {
              title: "Cosmos SDK",
              url: "https://docs.cosmos.network"
            },
            {
              title: "Cosmos Hub",
              url: "https://hub.cosmos.network"
            },
            {
              title: "Tendermint Core",
              url: "https://docs.tendermint.com"
            }
          ]
        },
        {
          title: "Community",
          children: [
            {
              title: "Cosmos blog",
              url: "https://blog.cosmos.network"
            },
            {
              title: "Forum",
              url: "https://forum.cosmos.network"
            },
            {
              title: "Chat",
              url: "https://discord.gg/W8trcGV"
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
