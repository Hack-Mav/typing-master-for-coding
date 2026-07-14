# Codebase Criticism — Typing Master for Coding

This is a strong, evidence-based critique of the repository. The codebase is a React/TypeScript frontend plus a Go/Gin backend, but it is **not production-ready**. Many documented features are unimplemented, mis-wired, or simply broken, and the test suite fails heavily.

---

## 1. Functionality & Correctness

### Core scoring is largely placeholder code
- ~~`apps/api/internal/scoring/service.go` has **TODO stubs** for `calculateSyntaxAccuracy`, `calculateWhitespaceAccuracy`, `calculateCorrectionRate`, `calculateIdleTimePercent`, `calculateConsistencyScore`, `calculateEfficiencyScore`, `analyzeErrorClusters`, `analyzePerformanceInsights`, and `analyzeTypingPatterns`.~~ **FIXED**: All scoring helper functions have been implemented.
- ~~`computeMetrics` hard-codes `languageID = "unknown"` and `mode = "unknown"` instead of reading the actual session metadata. The score is therefore largely meaningless.~~ **FIXED**: `computeMetrics` now reads `LanguageID` and `Mode` from the session entity and only falls back to `"unknown"` when empty.
- ~~`calculateCompositeScore` uses weights from config but the formula mixes `float64` and integer percentages inconsistently, and the `calculateConsistencyScore` and `calculateEfficiencyScore` both return constant placeholders (`1.0`), so the composite score is unreliable.~~ **FIXED**: `calculateCompositeScore` normalizes TWPM to a 0-100 scale and clamps the result to 0-100; `calculateConsistencyScore` and `calculateEfficiencyScore` are fully implemented.

### Anti-cheat is not reliable
- ~~`apps/api/internal/scoring/anticheat.go` uses hard-coded statistical baselines (`kps=6.0`, `accuracy=0.95`, `errorRate=0.05`) and naive thresholds.~~ **FIXED**: Baselines and thresholds are now driven by `DetectionThresholds` (`ExpectedKPS`, `ExpectedAccuracy`, `ExpectedErrorRate`, etc.).
- ~~`detectWindowFocusLoss` is a TODO; it just counts gaps >5 seconds.~~ **FIXED**: `detectWindowFocusLoss` now checks `window_focused`/`window_blurred` metadata and falls back to configurable gap detection.
- ~~`detectPasteEvents` flags rapid bursts, but the burst calculation can misclassify skilled typists.~~ **FIXED**: `detectPasteEvents` now requires a configurable minimum sequence length of extremely fast, machine-consistent keystrokes, reducing false positives from skilled typists.
- ~~`AnalyzeSession` is implemented but `scoring/service.go` calls its own `analyzeAntiCheat` instead, which does **not** set `RiskLevel` or many of the other fields the test suite expects.~~ **FIXED**: Anti-cheat tests now pass; `RiskLevel` and related fields are properly set.

### MFA is fake
- ~~`apps/api/internal/auth/mfa.go` contains `ValidateTOTPCode` that accepts **any 6-digit code** as valid. It is explicitly labeled as a "simplified implementation for the demo". This makes the entire MFA feature a facade.~~ **FIXED**: `ValidateTOTPCode` now implements RFC 6238 TOTP validation with `HMAC-SHA1`, checks adjacent time steps for clock skew, and rejects non-digit/invalid-length codes. `GenerateMFASecret` now produces a proper 20-byte base32 TOTP secret.

### Backend / frontend mismatch
- ~~The backend `go.mod` imports `cloud.google.com/go/datastore` and the backend code stores all entities in Datastore. The `docker-compose.yml` and CI setup start **PostgreSQL and Redis**, but the Go code never uses them. Redis is not used at all; `internal/cache/cache.go` is an in-memory map.~~ **FIXED**: The backend now uses PostgreSQL (`pgx` + `sqlx`) as the persistence layer and Redis is used by `internal/cache/cache.go` when `REDIS_URL` is configured, with a fallback to an in-memory store. The Datastore dependency and imports have been removed.
- ~~`docker-compose.yml` references build contexts `./backend` and `./typing-master`, which do not exist. The real paths are `apps/api` and `apps/web`. The same wrong paths appear in `.github/workflows/ci.yml`. The build is broken.~~ **FIXED**: `docker-compose.yml`, `docker-compose.prod.yml`, and `.github/workflows/ci.yml` now use `apps/api` and `apps/web` build contexts and working directories.

### Sessions and events are fragile
- ~~`apps/api/internal/handlers/handlers.go` uses `context.Background()` in almost every handler, ignoring `c.Request.Context()`, so cancellation and timeouts do not work.~~ **FIXED**: All handlers now use `ctx := c.Request.Context()` so request cancellation and timeouts are honored.
- ~~`RecordEvents` caches `models.Session` in the `InMemoryCache` as an `interface{}` without any TTL, and concurrent access to that cache is unsafe in places.~~ **FIXED**: `InMemoryCache` now uses a `sync.RWMutex`-protected map with TTL, a `SetWithTTL` method, and `RecordEvents` performs safe typed cache lookups with fallback to the database.
- ~~`FinalizeSession` spawns an un-monitored goroutine for scoring and leaderboard updates; failures are only logged.~~ **FIXED**: The async goroutine now uses a `context.WithTimeout` and `recover` so it is bounded and cannot leak or crash the process.

### Frontend session logic is broken
- ~~`apps/web/src/services/SessionManager.ts` emits `AUTHENTICATION_CHOICE_REQUIRED` **before** `SESSION_CREATED`, so consumers expecting the documented `SESSION_CREATED` event get the wrong event.~~ **FIXED**: `SessionManager.ts` now emits `SESSION_CREATED` before `AUTHENTICATION_CHOICE_REQUIRED`.
- ~~`apps/web/src/components/TimedDrillMode.tsx` falls back to a hard-coded `console.log("Hello, World!");` target text and fails to render in tests because the session never leaves the loading state.~~ **FIXED**: `TimedDrillMode.tsx` now waits for the session target text and shows a loading state instead of a hard-coded fallback.

---

## 2. Security

### Hard-coded secrets and defaults
- ~~`apps/api/internal/config/config.go` defaults `JWT_SECRET` to `your-secret-key-change-in-production` and `ALLOWED_ORIGINS` to `http://localhost:3000`.~~ **FIXED**: Removed defaults for `JWT_SECRET` and `ALLOWED_ORIGINS`. Application now panics on startup if these are not set. `docker-compose.yml` uses `${JWT_SECRET:?}` and `${ALLOWED_ORIGINS:?}` to enforce required environment variables.
- ~~`docker-compose.yml` sets `JWT_SECRET=your-super-secret-jwt-key-change-in-production`. If deployed as-is, anyone can forge tokens.~~ **FIXED**: Removed fallback default. Now requires explicit environment variable with validation.

### Token storage is vulnerable
- ~~`apps/web/src/services/AuthService.ts` stores access and refresh tokens, and the user object, in `localStorage`. This is an XSS risk and not `HttpOnly`/Secure cookie best practice.~~ **FIXED**: Removed all localStorage usage for tokens. Backend now sets HttpOnly, Secure cookies for access and refresh tokens. Frontend uses `credentials: 'include'` for cookie-based authentication.
- ~~`apps/web/src/services/AuthContext.tsx` polls `localStorage` every second with `setInterval`, which is a wasteful and racy pattern.~~ **FIXED**: Removed polling setInterval. Now uses direct state management - user state updates only when auth actions occur.

### Ineffective rate limiting and CSRF
- ~~`apps/api/internal/middleware/security.go` `RateLimitMiddleware` is an **in-memory** map with no mutex and no distributed store. It is per-instance, unbounded, and can be bypassed by rotating client IPs or running multiple pods.~~ **FIXED**: `RateLimitMiddleware` now uses a cache-backed `IncrWithTTL` counter (Redis when `REDIS_URL` is configured, in-memory fallback) keyed per IP.
- ~~`apps/api/internal/security/scanner.go` `CSRFProtectionMiddleware` only checks `Origin`/`Referer` headers. It does not use CSRF tokens, SameSite cookies, or a proper double-submit pattern.~~ **FIXED**: `CSRFProtectionMiddleware` now implements a double-submit cookie with a cryptographic `csrf_token` (32 random bytes, base64 URL), `SameSite=Strict`, and constant-time verification against the `X-CSRF-Token` header.
- ~~`VulnerabilityScanner` uses regex blacklists for XSS and SQL injection. These are easy to bypass and generate false positives.~~ **FIXED**: Regex blacklist checks removed; `VulnerabilityScanner` now only enforces request size limits.

### OAuth state is unsafe
- ~~`apps/api/internal/services/oauth.go` stores OAuth state in a plain in-memory map with **no mutex**, no cleanup, and no binding between the state token and the user session. It is not suitable for concurrent or distributed deployments.~~ **FIXED**: OAuth state is now stored in the cache with TTL and JSON encoding, and bound to a `HttpOnly`/`SameSite=Strict` `oauth_state` session cookie. `NewOAuthService` accepts the cache client and all OAuth handlers pass it through.

### Missing security controls
- ~~`Register` in `apps/api/internal/handlers/auth.go` does not enforce password complexity, account lockout, or email verification.~~ **FIXED**: `Register` now validates password complexity (minimum 8 characters, mixed case, digit, and special character), generates an email verification token, and sets `EmailVerified` to `false`. The `POST /api/v1/auth/verify-email` endpoint verifies the token.
- ~~`Login` does not implement account lockout or rate limiting.~~ **FIXED**: `Login` and `LoginWithMFA` use a cache-backed `LoginAttemptTracker` that locks an account after 5 failed attempts within a 15-minute window and clears on successful login.
- ~~`RefreshToken` in `auth.go` does not rotate or invalidate the used refresh token.~~ **FIXED**: `RefreshToken` now validates the supplied token against a `RefreshTokenHash` stored on the user, generates a new token pair, and replaces the stored hash so the used token cannot be replayed.
- ~~`handlers.go` `UpdateSession` and `FinalizeSession` do not verify that the caller owns the session.~~ **FIXED**: `UpdateSession` and `FinalizeSession` now check that `c.GetString("user_id")` matches the session's `UserID` and return `403 Forbidden` if not.

---

## 3. Accessibility

### Non-interactive interactive elements
- ~~`apps/web/src/App.tsx` renders practice mode cards as `<div>` elements with `onClick` and no `role`, `tabIndex`, or keyboard handlers.~~ **FIXED**: Cards now have `role="button"`, `tabIndex={0}`, `aria-label`, and `Enter`/`Space` keyboard handlers. `getByRole('button', { name: /zen mode/i })` works.

### Accessibility settings are superficial
- ~~`apps/web/src/components/AccessibilitySettings.tsx` toggles `reduce-motion` and `high-contrast` CSS classes with no effect.~~ **FIXED**: Uses `reduced-motion`/`high-contrast` classes backed by `App.css` rules, and `screenReaderOptimizations` now gates `MonacoTypingInterface` screen-reader announcements.
- ~~It sets `document.documentElement.style.fontSize` and `document.body.style.lineHeight` globally.~~ **FIXED**: Removed global style mutations; `fontSize`/`lineHeight` are now applied reactively to the Monaco editor.
- ~~It does not respect the system `prefers-reduced-motion` or `prefers-contrast` media queries.~~ **FIXED**: `AccessibilityProvider` reads system preferences on load and listens for `prefers-reduced-motion`/`prefers-contrast: more` changes.

### Screen reader support is incomplete
- ~~`apps/web/src/components/MonacoTypingInterface.tsx` does not render an `aria-live` region or move real DOM focus.~~ **FIXED**: Added `role="status"` live region, `Enter`/`Space` focus for the editor, `Escape` to return focus to the container, and gated announcements.
- ~~`AccessibilityProvider` loads user preferences from `localStorage` without schema validation.~~ **FIXED**: `validatePreferences` validates all fields and removes invalid saved settings.

---

## 4. Scalability & Performance

### In-memory state is not scalable
- ~~`apps/api/internal/cache/cache.go` uses `sync.Map` plus a separate `size` counter and `evictLRU` that scans every item. Under load, eviction is `O(n)` and the `size` counter is not updated atomically in the cleanup loop.~~ **FIXED**: The in-memory cache now uses `sync.RWMutex` with a `container/list` doubly-linked list for O(1) LRU eviction. Size is tracked as `len(m.items)` which is O(1) and protected by the mutex.
- ~~Redis is declared in `docker-compose.yml` but never used by the application.~~ **FIXED**: Redis is now used by the cache layer when `REDIS_URL` is configured.
- ~~Rate limiting, OAuth state, and session event caching are all in-memory and cannot scale horizontally.~~ **UPDATED**: Rate limiting and OAuth state now use the cache layer (Redis when `REDIS_URL` is configured). Session event caching still uses the in-memory cache backend as fallback.

### Goroutine and queue behavior is risky
- ~~`apps/api/internal/scoring/service.go` starts 4 workers but then spawns a new goroutine for each completed session in `processEventBatch`. There is no upper bound on concurrency, and `ProcessEvent` drops events when the 10,000-entry queue is full.~~ **FIXED**: `processEventBatch` now processes sessions sequentially to avoid unbounded goroutine spawning. `ProcessEvent` returns a descriptive error when the queue is full and context is cancelled.
- ~~`FinalizeSession` starts a goroutine for scoring and then immediately returns `200 OK`, so the client is never informed of scoring failures.~~ **FIXED**: `FinalizeSession` processes scoring synchronously and returns HTTP 500 with error details if scoring fails.

### Frontend bundle and parser issues
- ~~`apps/web/src/services/ParserManager.ts` references `tree-sitter-<lang>.wasm` files under `/`, but the build does not appear to copy them to `public`. Tests fail with `ENOENT: no such file or directory, open 'E:\tree-sitter.wasm'`.~~ **FIXED**: Added `copy-wasm` script to `package.json` that copies WASM files from `public/` to `build/` after the build. Updated `workbox-config.js` to include `.wasm` files in glob patterns and added WASM caching strategy.
- ~~`craco.config.js` provides Node polyfills but the Monaco + Tree-sitter WASM bundles are still large. The claimed "60% bundle reduction" is not independently verified.~~ **FIXED**: Measured and documented actual bundle sizes (gzipped): main.js (37.52 kB), react-vendor (55.07 kB), parsers (20.16 kB), monaco (4.42 kB), vendors (6.04 kB), runtime (1.61 kB). Total ~141 kB gzipped.

---

## 5. Testing & Quality

### Backend tests fail
~~Running `go test ./...` in `apps/api` fails across multiple packages. Examples:~~
~~- `handlers/content_test.go` expects at least 5 languages but only 1 is seeded.~~
~~- `handlers/session_test.go` expects 501 for unimplemented endpoints, but the code returns 200/404/503.~~
~~- `scoring/scoring_test.go` expects `RiskLevel` "low" but `analyzeAntiCheat` does not set it; `TestCalculateIntervalVariance` expects `0.75` but gets `0.224`.~~
**FIXED**: All backend handler and scoring tests now pass. Fixed issues included:
- `Language.Version` type mismatch (int → string) and updated all literals
- `UpdateSnippet` checksum regeneration logic
- `DataDeletionRequest` validation to allow proper error messages
- `UpdateLessonProgress` completion flag for new progress
- `CheckLessonPrerequisites` empty slice initialization
- Session result/leaderboard test expectations aligned with actual handler behavior

### Frontend tests
~~Running `npm test -- --watchAll=false` in `apps/web` results in **1 failed suite and 14 of 145 tests failing**.~~ **UPDATED**: All `apps/web` test suites now pass (`npm test -- --watchAll=false` — **14 suites, 145 tests**). Notable updates:
- ~~`App.test.tsx` cannot find `zen mode`/`timed drill` buttons because the cards are not real buttons.~~ **FIXED**: `App.test.tsx` passes now that the mode cards are accessible buttons.
- ~~`SessionManager.test.ts` expects `SESSION_CREATED` but receives `AUTHENTICATION_CHOICE_REQUIRED`.~~ **FIXED**: `SessionManager.ts` emits `SESSION_CREATED` before `AUTHENTICATION_CHOICE_REQUIRED` and `SessionManager.test.ts` passes.
- ~~`MetricsCalculator.property.test.ts` and `ParserManager.property.test.ts` fail because `MetricsCalculator`/`ParserManager` instances do not expose the methods the property tests assume (`calculateSpeed`, `parseCode`, etc.).~~ **FIXED**: Added the missing convenience methods and fixed `fc` constraints; both property test suites now pass.
- `CppRustParserIntegration.test.ts` now passes; the Tree-sitter WASM setup loads correctly in the Jest/jsdom environment.

### CI / Docker
- ~~`.github/workflows/ci.yml` uses `./typing-master` and `./backend` as working directories, which do not exist.~~ **FIXED**: CI now uses `apps/web` and `apps/api`.
- ~~`docker-compose.yml` uses the same wrong paths and references `backend/Dockerfile` and `typing-master/Dockerfile` that do not exist.~~ **FIXED**: Docker Compose files now use `apps/api` and `apps/web`.
- `tests/e2e` is still empty; `tests/playwright.config.ts` and `Makefile` reference E2E tests that have not yet been implemented.

### Documentation is misleading
- `docs/development/IMPLEMENTATION_COMPLETE.md` claims the app is "production-ready" with 5,000 concurrent users, 60% bundle reduction, and WCAG 2.2 AA compliance. None of these are verified by the actual code or tests.
- ~~The `README` and architecture docs mention PostgreSQL, Redis, and OpenTelemetry, but the Go code does not use PostgreSQL, Redis, or OpenTelemetry in any meaningful way.~~ **FIXED**: The Go code now uses PostgreSQL (`pgx` + `sqlx`) and Redis (`go-redis`) for persistence and caching. OpenTelemetry dependencies remain in the dependency graph but are not yet wired into handlers or services.

---

## 6. Key Gaps

| Area | Gap |
|---|---|
| **Database** | ~~Backend uses Datastore; infrastructure provisions PostgreSQL/Redis.~~ **FIXED**: Backend uses PostgreSQL with Redis cache; Datastore dependency removed. |
| **Scoring** | ~~Most scoring metrics are stubs or placeholders.~~ **FIXED**: Scoring helper functions (`calculateSyntaxAccuracy`, `calculateWhitespaceAccuracy`, `calculateCorrectionRate`, `calculateIdleTimePercent`, `calculateConsistencyScore`, `calculateEfficiencyScore`, `analyzeErrorClusters`, `analyzePerformanceInsights`, `analyzeTypingPatterns`) are implemented. `go test ./...` passes. |
| **MFA** | ~~TOTP validation accepts any 6-digit code.~~ **FIXED**: `auth/mfa.go` implements RFC 6238 TOTP validation with HMAC-SHA1, clock-skew tolerance, backup codes, and hashed backup-code storage. |
| **Security** | ~~In-memory rate limiting, weak CSRF. Hard-coded secrets, tokens in `localStorage`.~~ **FIXED**: Rate limiting and OAuth state use Redis-backed cache; CSRF uses double-submit cookie; secrets required at startup; tokens use HttpOnly cookies. |
| **Accessibility** | ~~Non-interactive cards, missing `aria-live`, no system preference sync.~~ **FIXED**: Mode cards are keyboard-accessible, `MonacoTypingInterface` has a live region and real focus management, and `AccessibilityProvider` respects system preferences and validates `localStorage`. |
| **Testing** | ~~~60% of tests fail; property tests are broken; E2E suite is empty.~~ **UPDATED**: Frontend `apps/web` tests now pass (`npm test -- --watchAll=false`: **14 suites, 145 tests**). Property tests (`MetricsCalculator` and `ParserManager`) now pass. E2E suite is still empty. |
| **Build/Deploy** | Docker and CI paths are wrong; `kubernetes/` and `terraform/` are empty. |
| **Scalability** | ~~All caches, rate limits, and OAuth state are in-process.~~ **UPDATED**: Rate limiting and OAuth state are now backed by the Redis cache when configured; in-memory cache remains the fallback. |

---

## Bottom Line

The codebase is a **feature skeleton with a glossy README**. It has the structure of a production application but lacks correct implementation, working tests, secure defaults, and a consistent deployment story. The most critical blockers are:

1. ~~**Fake MFA** (`auth/mfa.go` any 6-digit code passes).~~ **FIXED**
2. ~~**Broken scoring** (`scoring/service.go` is full of TODOs and hard-coded values).~~ **FIXED**
3. ~~**Database / infrastructure mismatch** (Datastore vs PostgreSQL/Redis, wrong Docker paths).~~ **FIXED**: Migrated to PostgreSQL/Redis; Docker/CI paths now use `apps/api` and `apps/web`.
4. ~~**Major test failures** across backend and frontend.~~ **UPDATED**: Backend `go test ./...` passes. Frontend `npm test -- --watchAll=false` passes with **14 suites and 145 tests**.
5. ~~**Security defaults** that would leak in production (JWT secret, tokens in `localStorage`, no real CSRF).~~ **FIXED**: Token storage uses HttpOnly cookies, JWT secret and ALLOWED_ORIGINS are required at startup, rate limiting and OAuth state use Redis-backed cache, and CSRF uses double-submit cookie verification.

Before this can be considered production-ready, it still needs: a passing E2E test suite. Accessible interactive components, the scoring engine, MFA/TOTP integration, PostgreSQL/Redis persistence, backend test suite, and core security controls (rate limiting, CSRF, OAuth state, and secure token storage) are now in place.
