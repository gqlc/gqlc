module.exports = {
  plugins: [
    "@vuepress/nprogress",
    "@vuepress/pwa"
  ],
  themeConfig: {
    repo: "gqlc/gqlc",
    nav: [
      { text: "Guide", link: "/guide/" }
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
      ]
    }
  }
};
