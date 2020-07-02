module.exports = {
  title: "GraphQL Compiler",
  description: "gqlc is a compiler for the GraphQL IDL",
  plugins: [
    "@vuepress/nprogress",
    "@vuepress/pwa"
  ],
  themeConfig: {
    repo: "gqlc/gqlc",
    nav: [
      { text: "Guide", link: "/guide/" },
      { text: "Generators", link: "/generators/" }
    ],
    sidebar: {
      "/guide/": [
        {
          title: "Guide",
          collapsable: false,
          children: [
            "",
            "getting-started",
            "config"
          ]
        },
        {
          title: "Advanced",
          collapsable: false,
          children: [
            "importing-types",
            "remote-service-as-a-source"
          ]
        }
      ],
      "/generators/": [
        {
          title: "Generators",
          collapsable: false,
          children: [
            "documentation",
            "go",
            "javascript"
          ]
        },
        {
          title: "Community Maintained",
          collapsable: false,
          children: [
            "community"
          ]
        }
      ]
    }
  }
};
