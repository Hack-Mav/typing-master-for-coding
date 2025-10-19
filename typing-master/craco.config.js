module.exports = {
  webpack: {
    configure: (webpackConfig) => {
      const isProduction = process.env.NODE_ENV === 'production';

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

      // Use faster source maps in development
      if (!isProduction) {
        webpackConfig.devtool = 'eval-cheap-module-source-map';
      }

      // Optimize bundle splitting (only in production)
      if (isProduction) {
        webpackConfig.optimization = {
          ...webpackConfig.optimization,
          splitChunks: {
          chunks: 'all',
          cacheGroups: {
            // Vendor libraries
            vendor: {
              test: /[\\/]node_modules[\\/]/,
              name: 'vendors',
              priority: 10,
              reuseExistingChunk: true,
            },
            // Tree-sitter parsers
            parsers: {
              test: /[\\/]node_modules[\\/].*tree-sitter.*[\\/]/,
              name: 'parsers',
              priority: 20,
              reuseExistingChunk: true,
            },
            // Monaco editor
            monaco: {
              test: /[\\/]node_modules[\\/]@monaco-editor[\\/]/,
              name: 'monaco',
              priority: 20,
              reuseExistingChunk: true,
            },
            // React libraries
            react: {
              test: /[\\/]node_modules[\\/](react|react-dom|react-router-dom)[\\/]/,
              name: 'react-vendor',
              priority: 15,
              reuseExistingChunk: true,
            },
            // Common code
            common: {
              minChunks: 2,
              priority: 5,
              reuseExistingChunk: true,
              enforce: true,
            },
          },
          maxInitialRequests: 25,
          maxAsyncRequests: 25,
          minSize: 20000,
          maxSize: 244000,
        },
        runtimeChunk: {
          name: 'runtime',
        },
        usedExports: true,
        sideEffects: true,
        };

        // Add performance hints (production only)
        webpackConfig.performance = {
          maxEntrypointSize: 512000,
          maxAssetSize: 512000,
          hints: 'warning',
        };
      } else {
        // Development: disable performance hints
        webpackConfig.performance = {
          hints: false,
        };
      }

      return webpackConfig;
    },
  },
};