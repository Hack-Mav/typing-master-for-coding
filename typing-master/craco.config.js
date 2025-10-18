module.exports = {
  webpack: {
    configure: (webpackConfig) => {
      // Add fallbacks for Node.js modules that aren't available in the browser
      webpackConfig.resolve.fallback = {
        ...webpackConfig.resolve.fallback,
        "fs": false,
        "fs/promises": false,
        "module": false,
        "path": require.resolve("path-browserify"),
        "util": require.resolve("util/"),
      };

      // Ignore Node.js modules in web-tree-sitter
      webpackConfig.resolve.alias = {
        ...webpackConfig.resolve.alias,
      };

      return webpackConfig;
    },
  },
};