// IndexedDB utilities for offline data storage
import { openDB, IDBPDatabase } from 'idb';

/**
 * In-memory fallback for IndexedDB used in test environments or when
 * IndexedDB is not available. Exposes the same API surface as idb's IDBPDatabase.
 */
class MemoryDatabase {
  private stores: Map<string, Map<string, any>> = new Map();

  private getStore(name: string): Map<string, any> {
    if (!this.stores.has(name)) {
      this.stores.set(name, new Map());
    }
    return this.stores.get(name)!;
  }

  private getKey(value: any, key?: any): string {
    if (key !== undefined) return String(key);
    return String(value.id ?? value.key ?? Date.now());
  }

  async get(storeName: string, key: any): Promise<any> {
    return this.getStore(storeName).get(String(key));
  }

  async put(storeName: string, value: any, key?: any): Promise<void> {
    this.getStore(storeName).set(this.getKey(value, key), value);
  }

  async getAll(storeName: string): Promise<any[]> {
    return Array.from(this.getStore(storeName).values());
  }

  async delete(storeName: string, key: any): Promise<void> {
    this.getStore(storeName).delete(String(key));
  }

  async clear(storeName: string): Promise<void> {
    this.getStore(storeName).clear();
  }
}

interface SessionEvent {
  timestampMs: number;
  keyPressed: string;
  action: 'down' | 'up';
  cursorPosition: number;
  errorFlag: boolean;
  expectedToken?: string;
  metadata?: Record<string, any>;
}

interface SessionData {
  id: string;
  userId?: string;
  mode: string;
  languageId: string;
  lessonId?: string;
  snippetId?: string;
  startedAt: number;
  endedAt?: number;
  events: SessionEvent[];
  settings: Record<string, any>;
  synced: boolean;
}

interface LessonData {
  id: string;
  languageId: string;
  title: string;
  difficulty: number;
  content: string;
  version: number;
  cachedAt: number;
}

interface SnippetData {
  id: string;
  languageId: string;
  title: string;
  sourceCode: string;
  tags: string[];
  difficulty: number;
  cachedAt: number;
}

class IndexedDBManager {
  private db: IDBPDatabase | null = null;
  private readonly DB_NAME = 'TypingMasterDB';
  private readonly DB_VERSION = 1;

  async init(): Promise<void> {
    if (this.db) return;

    try {
      if (typeof indexedDB === 'undefined') {
        throw new Error('IndexedDB is not available');
      }

      this.db = await openDB(this.DB_NAME, this.DB_VERSION, {
        upgrade(db) {
          // Sessions store
          const sessionsStore = db.createObjectStore('sessions', {
            keyPath: 'id',
          });
          sessionsStore.createIndex('by-user', 'userId');
          sessionsStore.createIndex('by-sync-status', 'synced');

          // Lessons store
          db.createObjectStore('lessons', { keyPath: 'id' });

          // Snippets store
          db.createObjectStore('snippets', { keyPath: 'id' });

          // User settings store
          db.createObjectStore('userSettings', { keyPath: 'key' });
        },
      });
    } catch (error) {
      console.warn('IndexedDB unavailable, using in-memory fallback:', error);
      this.db = new MemoryDatabase() as unknown as IDBPDatabase;
    }
  }

  // Session management
  async saveSession(session: SessionData): Promise<void> {
    await this.init();
    await this.db!.put('sessions', session);
  }

  async getSession(id: string): Promise<SessionData | undefined> {
    await this.init();
    return await this.db!.get('sessions', id);
  }

  async getUnsyncedSessions(): Promise<SessionData[]> {
    await this.init();
    const allSessions = await this.db!.getAll('sessions');
    return allSessions.filter((session: SessionData) => !session.synced);
  }

  async markSessionSynced(id: string): Promise<void> {
    await this.init();
    const session = await this.db!.get('sessions', id);
    if (session) {
      session.synced = true;
      await this.db!.put('sessions', session);
    }
  }

  // Lesson management
  async saveLesson(lesson: LessonData): Promise<void> {
    await this.init();
    await this.db!.put('lessons', { ...lesson, cachedAt: Date.now() });
  }

  async getLesson(id: string): Promise<LessonData | undefined> {
    await this.init();
    return await this.db!.get('lessons', id);
  }

  async getLessonsByLanguage(languageId: string): Promise<LessonData[]> {
    await this.init();
    const allLessons = await this.db!.getAll('lessons');
    return allLessons.filter(lesson => lesson.languageId === languageId);
  }

  // Snippet management
  async saveSnippet(snippet: SnippetData): Promise<void> {
    await this.init();
    await this.db!.put('snippets', { ...snippet, cachedAt: Date.now() });
  }

  async getSnippet(id: string): Promise<SnippetData | undefined> {
    await this.init();
    return await this.db!.get('snippets', id);
  }

  async getSnippetsByLanguage(languageId: string): Promise<SnippetData[]> {
    await this.init();
    const allSnippets = await this.db!.getAll('snippets');
    return allSnippets.filter(snippet => snippet.languageId === languageId);
  }

  // User settings management
  async saveSetting(key: string, value: any): Promise<void> {
    await this.init();
    await this.db!.put('userSettings', { key, value, updatedAt: Date.now() });
  }

  async getSetting(key: string): Promise<any> {
    await this.init();
    const setting = await this.db!.get('userSettings', key);
    return setting?.value;
  }

  async getAllSettings(): Promise<Record<string, any>> {
    await this.init();
    const settings = await this.db!.getAll('userSettings');
    return settings.reduce((acc: Record<string, any>, setting: any) => {
      acc[setting.key] = setting.value;
      return acc;
    }, {});
  }

  // Cleanup methods
  async clearOldCache(maxAge: number = 7 * 24 * 60 * 60 * 1000): Promise<void> {
    await this.init();
    const cutoff = Date.now() - maxAge;

    // Clear old lessons
    const lessons = await this.db!.getAll('lessons');
    for (const lesson of lessons) {
      if ((lesson as any).cachedAt < cutoff) {
        await this.db!.delete('lessons', (lesson as any).id);
      }
    }

    // Clear old snippets
    const snippets = await this.db!.getAll('snippets');
    for (const snippet of snippets) {
      if ((snippet as any).cachedAt < cutoff) {
        await this.db!.delete('snippets', (snippet as any).id);
      }
    }
  }

  async clearAllData(): Promise<void> {
    await this.init();
    await this.db!.clear('sessions');
    await this.db!.clear('lessons');
    await this.db!.clear('snippets');
    await this.db!.clear('userSettings');
  }
}

export const indexedDBManager = new IndexedDBManager();
export type { SessionEvent, SessionData, LessonData, SnippetData };
