module.exports = {
  globDirectory: 'build/',
  globPatterns: [
    '**/*.{js,css,html,png,jpg,jpeg,svg,ico,woff,woff2,ttf,eot,wasm}'
  ],
  swDest: 'build/service-worker.js',
  clientsClaim: true,
  skipWaiting: true,
  runtimeCaching: [
    {
      urlPattern: /^https:\/\/fonts\.googleapis\.com/,
      handler: 'StaleWhileRevalidate',
      options: {
        cacheName: 'google-fonts-stylesheets',
      },
    },
    {
      urlPattern: /^https:\/\/fonts\.gstatic\.com/,
      handler: 'CacheFirst',
      options: {
        cacheName: 'google-fonts-webfonts',
      },
    },
    {
      urlPattern: /\.(?:png|jpg|jpeg|svg|gif|webp)$/,
      handler: 'CacheFirst',
      options: {
        cacheName: 'images',
      },
    },
    {
      urlPattern: /\/api\//,
      handler: 'NetworkFirst',
      options: {
        cacheName: 'api-cache',
        networkTimeoutSeconds: 3,
      },
    },
    {
      urlPattern: /\/lessons\//,
      handler: 'StaleWhileRevalidate',
      options: {
        cacheName: 'lessons-cache',
      },
    },
    {
      urlPattern: /\.wasm$/,
      handler: 'CacheFirst',
      options: {
        cacheName: 'wasm-cache',
        expiration: {
          maxEntries: 10,
          maxAgeSeconds: 60 * 60 * 24 * 365, // 1 year
        },
      },
    },
  ],
  navigateFallback: '/index.html',
  navigateFallbackDenylist: [/^\/_/, /\/[^/?]+\.[^/]+$/],
};