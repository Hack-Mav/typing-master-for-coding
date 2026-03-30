/**
 * Lazy Loading Utilities
 * Provides utilities for code splitting and lazy loading
 */

import { lazy, ComponentType, LazyExoticComponent } from 'react';

/**
 * Lazy load a component with retry logic
 */
export function lazyWithRetry<T extends ComponentType<any>>(
  componentImport: () => Promise<{ default: T }>,
  retries = 3,
  interval = 1000
): LazyExoticComponent<T> {
  return lazy(() => {
    return new Promise<{ default: T }>((resolve, reject) => {
      const attemptImport = (retriesLeft: number) => {
        componentImport()
          .then(resolve)
          .catch(error => {
            if (retriesLeft === 0) {
              reject(error);
              return;
            }

            console.warn(
              `Failed to load component, retrying... (${retriesLeft} attempts left)`
            );

            setTimeout(() => {
              attemptImport(retriesLeft - 1);
            }, interval);
          });
      };

      attemptImport(retries);
    });
  });
}

/**
 * Preload a lazy component
 */
export function preloadComponent<T extends ComponentType<any>>(
  lazyComponent: LazyExoticComponent<T>
): void {
  const component = lazyComponent as any;
  if (component._payload && component._payload._status === -1) {
    component._payload._result();
  }
}

/**
 * Lazy load parser WASM file
 */
export async function lazyLoadParser(
  language: string,
  retries = 3
): Promise<ArrayBuffer> {
  const languageMap: Record<string, string> = {
    python: 'tree-sitter-python.wasm',
    javascript: 'tree-sitter-javascript.wasm',
    typescript: 'tree-sitter-typescript.wasm',
    cpp: 'tree-sitter-cpp.wasm',
    rust: 'tree-sitter-rust.wasm',
    yaml: 'tree-sitter-yaml.wasm',
  };

  const wasmFile = languageMap[language.toLowerCase()];
  if (!wasmFile) {
    throw new Error(`Unsupported language: ${language}`);
  }

  const url = `/${wasmFile}`;

  for (let attempt = 0; attempt <= retries; attempt++) {
    try {
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      return await response.arrayBuffer();
    } catch (error) {
      if (attempt === retries) {
        throw new Error(
          `Failed to load parser for ${language} after ${retries + 1} attempts: ${error}`
        );
      }
      console.warn(
        `Failed to load parser, retrying... (attempt ${attempt + 1}/${retries + 1})`
      );
      await new Promise(resolve => setTimeout(resolve, 1000 * (attempt + 1)));
    }
  }

  throw new Error(`Failed to load parser for ${language}`);
}

/**
 * Prefetch resources
 */
export function prefetchResource(
  url: string,
  type: 'script' | 'style' | 'font' | 'fetch' = 'fetch'
): void {
  if (typeof document === 'undefined') return;

  const link = document.createElement('link');
  link.rel = 'prefetch';
  link.as = type;
  link.href = url;
  document.head.appendChild(link);
}

/**
 * Preload critical resources
 */
export function preloadResource(
  url: string,
  type: 'script' | 'style' | 'font' | 'fetch' = 'fetch'
): void {
  if (typeof document === 'undefined') return;

  const link = document.createElement('link');
  link.rel = 'preload';
  link.as = type;
  link.href = url;
  document.head.appendChild(link);
}

/**
 * Dynamically import a module with retry
 */
export async function dynamicImportWithRetry<T>(
  importFn: () => Promise<T>,
  retries = 3,
  interval = 1000
): Promise<T> {
  for (let attempt = 0; attempt <= retries; attempt++) {
    try {
      return await importFn();
    } catch (error) {
      if (attempt === retries) {
        throw error;
      }
      console.warn(
        `Dynamic import failed, retrying... (attempt ${attempt + 1}/${retries + 1})`
      );
      await new Promise(resolve =>
        setTimeout(resolve, interval * (attempt + 1))
      );
    }
  }

  throw new Error('Dynamic import failed');
}

/**
 * Check if a resource is cached
 */
export async function isResourceCached(url: string): Promise<boolean> {
  if (!('caches' in window)) return false;

  try {
    const cache = await caches.open('workbox-precache-v2');
    const response = await cache.match(url);
    return !!response;
  } catch (error) {
    console.warn('Failed to check cache:', error);
    return false;
  }
}

/**
 * Preload parsers for specific languages
 */
export async function preloadParsers(languages: string[]): Promise<void> {
  const promises = languages.map(async language => {
    try {
      const isCached = await isResourceCached(`/tree-sitter-${language}.wasm`);
      if (!isCached) {
        prefetchResource(`/tree-sitter-${language}.wasm`, 'fetch');
      }
    } catch (error) {
      console.warn(`Failed to preload parser for ${language}:`, error);
    }
  });

  await Promise.allSettled(promises);
}

const lazyLoadUtils = {
  lazyWithRetry,
  preloadComponent,
  lazyLoadParser,
  prefetchResource,
  preloadResource,
  dynamicImportWithRetry,
  isResourceCached,
  preloadParsers,
};

export default lazyLoadUtils;
