import { precacheAndRoute, cleanupOutdatedCaches } from 'workbox-precaching';
import { registerRoute } from 'workbox-routing';
import { StaleWhileRevalidate, CacheFirst } from 'workbox-strategies';

// Precache all static assets
precacheAndRoute(self.__WB_MANIFEST);
cleanupOutdatedCaches();

// Cache API responses with stale-while-revalidate strategy
registerRoute(
  ({ url }) => url.pathname.startsWith('/api/'),
  new StaleWhileRevalidate({
    cacheName: 'api-cache',
    plugins: [
      {
        cacheKeyWillBeUsed: async ({ request }) => {
          return `${request.url}?timestamp=${Date.now()}`;
        },
      },
    ],
  })
);

// Cache lesson content and snippets
registerRoute(
  ({ url }) =>
    url.pathname.includes('/lessons/') || url.pathname.includes('/snippets/'),
  new CacheFirst({
    cacheName: 'content-cache',
    plugins: [
      {
        cacheWillUpdate: async ({ response }) => {
          return response.status === 200 ? response : null;
        },
      },
    ],
  })
);

// Handle offline fallback
self.addEventListener('fetch', event => {
  if (event.request.mode === 'navigate') {
    event.respondWith(
      fetch(event.request).catch(() => {
        return caches.match('/offline.html');
      })
    );
  }
});

// Handle background sync for session data
self.addEventListener('sync', event => {
  if (event.tag === 'session-sync') {
    event.waitUntil(syncSessionData());
  }
});

async function syncSessionData() {
  // Implementation for syncing offline session data
  console.log('Syncing session data...');
}
