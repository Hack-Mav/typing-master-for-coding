module.exports = {
  globDirectory: 'build/',
  globPatterns: [
    '**/*.{js,css,html,png,jpg,jpeg,svg,ico,woff,woff2,ttf,eot}'
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
  ],
  navigateFallback: '/index.html',
  navigateFallbackDenylist: [/^\/_/, /\/[^/?]+\.[^/]+$/],
};