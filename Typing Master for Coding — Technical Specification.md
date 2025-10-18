# **Typing Master for Coding — Technical Specification**

## **1\. Overview**

A desktop-first, web-enabled application to help developers practice typing real code with syntax-aware guidance. It supports full syntax tutorials and practice content for **C++**, **Rust**, **Python**, **JavaScript**, and **YAML**. The app measures speed and accuracy with code-aware metrics, shows attractive scoreboards, and also offers a **stress‑free practice mode** without metrics.

## **2\. Goals & Non‑Goals**

**Goals**

* Build accurate, code‑aware typing practice across 5 languages.

* Offer multiple modes: Tutorials, Timed Drills, Accuracy Mode, Zen (no metrics), Custom Snippets, and Assessments.

* Provide measurable progress via robust metrics, history, and leaderboards.

* Keep privacy and accessibility as first‑class concerns.

**Non‑Goals**

* Executing user code or judging semantic correctness of program output.

* Full IDE features (debugger, intellisense). Only lightweight syntax highlighting and parsing.

## **3\. Personas**

* **Beginner Coder**: learns syntax via guided lessons and gentle feedback.

* **Professional Developer**: wants speed drills on idiomatic patterns, shortcuts, and punctuation accuracy.

* **Competitive Typer**: targets high scores, leaderboards, and badges.

* **Wellness‑Focused User**: prefers Zen Mode, low‑distraction, no pressure.

## **4\. Modes & User Journeys**

1. **Syntax Tutorials** (per language)

   * Structured lessons: tokens, operators, control flow, data structures, idioms.

   * Progression: Intro → Core Syntax → Idioms → Advanced Patterns → Mixed Review.

   * Hints: token tooltips, inline explanations, reference links.

2. **Timed Drill**

   * Fixed durations (1/3/5/10 minutes) or snippet length target.

   * Real‑time HUD: CPM, token WPM, accuracy, mistakes, time remaining.

3. **Accuracy Mode**

   * Focus on error‑free typing. Score penalizes corrections and syntax mismatches.

4. **Zen Mode (Stress‑Free)**

   * No metrics or HUD. Minimal UI, ambient focus. Optional post‑session summary disabled by default.

5. **Custom Snippets & Playlists**

   * Users import code or YAML; app normalizes whitespace and tabs per language conventions.

   * Playlist builder sequences snippets by theme/difficulty.

6. **Assessments**

   * Periodic standardized tests to benchmark progress and award badges.

## **5\. Functional Requirements**

### **5.1 Provided Requirements (mapped)**

* **FR‑1**: Support complete syntax tutorials for **C++, Rust, Python, JavaScript, YAML**.

* **FR‑2**: Provide **metrics** for typing efficiency, errors, and performance.

* **FR‑3**: **Attractive scoreboard** showing outcome of each practice session.

* **FR‑4**: **Practice session without metrics** (Zen Mode) for stress‑free typing.

### **5.2 Additional Aligned Requirements**

* **FR‑5**: Code‑aware tokenization using language grammars for precise error checks.

* **FR‑6**: AST‑shape conformity check (structure vs raw text) for advanced accuracy scoring.

* **FR‑7**: Multi‑language lesson catalog with difficulty tiers and prerequisites.

* **FR‑8**: Real‑time feedback toggles (per‑character, per‑token, or end‑of‑line).

* **FR‑9**: Keyboard layout support (QWERTY, AZERTY, QWERTZ, Colemak, Dvorak); user‑selectable tab width and indentation style.

* **FR‑10**: Offline‑first PWA support with local caching of lessons and sessions.

* **FR‑11**: Session history, trends, and insights (weekly summary, heatmaps).

* **FR‑12**: Leaderboards (global, friends, organization), anti‑cheat heuristics.

* **FR‑13**: Accessibility (WCAG 2.2 AA): screen reader labels, high‑contrast, font ligatures toggle, adjustable font size/line height.

* **FR‑14**: Internationalization of UI; locale‑aware number/date formatting; RTL support where applicable.

* **FR‑15**: Admin CMS for lesson creation, snippet curation, tagging, and A/B tests.

* **FR‑16**: Export summaries (PDF/CSV) and shareable result cards.

* **FR‑17**: Theming (light/dark/solarized), distraction‑free full‑screen mode.

* **FR‑18**: Privacy controls (anonymous mode); local‑only sessions if user declines account.

* **FR‑19**: API for importing curated snippets from Git or Gists (read‑only, metadata only).

* **FR‑20**: Versioned lesson content with deprecation and migration paths.

## **6\. Metrics & Scoring (Definitions)**

**Keystroke Metrics**

* **CPM** (Characters per minute): raw typed characters / minute.

* **Token WPM (tWPM)**: tokens per minute using language tokenization.

* **KPS** (Keystrokes per second) time series.

* **Backspace Rate**: backspaces / 100 characters.

* **Correction Latency**: mean time between mistake and correction.

**Accuracy Metrics**

* **Raw Accuracy**: correct characters / total characters.

* **Token Accuracy**: correctly typed tokens / total tokens.

* **Syntax Accuracy**: proportion of lines whose token sequence parses to expected AST (tolerant whitespace rules).

* **Structural Penalty**: missing/extra delimiters (braces, commas, colons) weight \> normal chars.

* **Whitespace Conformance**: matches indentation rules & trailing whitespace policy.

**Quality & Consistency**

* **Burst Consistency**: std. dev. of KPS across session.

* **Error Clusters**: contiguous mistake windows (hotspots).

* **Modifier Accuracy**: correct use of Shift/Alt for symbols.

**Composite Score** (for scoreboard)

* **Score S** \= α·tWPM \+ β·RawAccuracy \+ γ·SyntaxAccuracy − δ·ErrorWeight − ε·BackspaceRate − ζ·IdleTimePenalty

  * Default weights: α=2.0, β=50, γ=60, δ=0.3, ε=5, ζ=0.1 (tunable via A/B tests).

## **7\. Content Model (Lessons & Snippets)**

* **Language**: {id, name, version, parserId}

* **Lesson**: {id, languageId, title, difficulty, objectives\[\], prerequisites\[\], estimatedMins, tokensCovered\[\], snippetIds\[\], version}

* **Snippet**: {id, languageId, title, source, text, tags\[\], difficulty, estimatedTime, checksum}

* **Playlist**: {id, title, description, snippetIds\[\], authorId}

* **Assessment**: {id, languageId, blueprint: {sections\[\], timing, scoringWeights}}

**Content Rules**

* Snippet length bands (XS \<200 chars, S 200–400, M 400–800, L 800–1500).

* YAML snippets validated against a schema to ensure meaningful structure.

* Accessibility tags: complexity, camelCase density, punctuation‑heavy, Unicode.

## **8\. Error Checking & Parsing**

* Tokenization via **Tree‑sitter** grammars for C++, Rust, Python, JavaScript; YAML via a robust YAML parser.

* Real‑time comparison engine aligns user input with expected token stream.

* Parenthesis/brace/quote balance tracker; delimiter prediction cues (optional).

* AST sketch match: derive simplified AST (operator nodes, blocks, declarations) and compare shape.

* Whitespace policy layer per language (tabs vs spaces, indentation width, trailing newline).

## **9\. Architecture**

**High Level**

* **Client (PWA/SPA)**: React/TypeScript, Service Worker for offline cache, Web Workers for parsing & metrics, WebAssembly for Tree‑sitter parsers.

* **Backend API**: Go (gRPC/REST), stateless, JWT auth.

* **Content Service**: lesson/snippet CMS, versioning, tagging.

* **Scoring Service**: receives session events → computes metrics → stores results.

* **Leaderboard Service**: time‑windowed ranks with anti‑cheat filters.

* **Storage**: Postgres (core relational), Redis (leaderboard & sessions queue), S3/GCS (content exports), ClickHouse/BigQuery (telemetry analytics, optional).

**Event Flow**

1. Client emits keystroke events (batched) → Ingest API.

2. Scoring Service computes metrics (online for HUD; final recompute on submit).

3. Results persisted; Leaderboard updated asynchronously.

4. Client pulls summary & renders scoreboard.

## **10\. Data Model (Relational)**

**users**(id, handle, email?, locale, keyboard\_layout, privacy\_mode, created\_at)

**sessions**(id, user\_id?, mode, language\_id, lesson\_id?, snippet\_id?, started\_at, ended\_at, duration\_ms, settings\_json)

**session\_events**(id, session\_id, t\_ms, key, action, cursor\_pos, error\_flag, expected\_token?, meta\_json)

**results**(session\_id PK, cpm, twpm, raw\_accuracy, token\_accuracy, syntax\_accuracy, backspace\_rate, score, breakdown\_json)

**leaderboards**(id, scope, language\_id?, mode?, period, rank\_method)

**leaderboard\_entries**(board\_id, session\_id, user\_id?, score, rank, created\_at)

**lessons/snippets/playlists** per content model above.

## **11\. APIs (REST/JSON; gRPC mirrors)**

**Auth**

* POST /v1/auth/anonymous → {token}

* POST /v1/auth/signup → {user}

**Content**

* GET /v1/languages → list

* GET /v1/lessons?language=rust\&difficulty=2

* GET /v1/snippets/{id}

* POST /v1/snippets: import custom snippet (optional moderation)

**Sessions**

* POST /v1/sessions → create

* POST /v1/sessions/{id}/events → batch append

* POST /v1/sessions/{id}/finish → finalize & compute

* GET /v1/sessions/{id}/result → {metrics}

* GET /v1/users/{id}/history?period=30d

**Leaderboards**

* GET /v1/leaderboards?scope=global\&language=python\&period=weekly

* GET /v1/leaderboards/{id}/entries

**Admin**

* POST /v1/lessons (CMS)

* PUT /v1/lessons/{id}

* POST /v1/abtests

## **12\. UI/UX Specification**

**General**

* Monospace font with ligatures toggle, adjustable size/line height.

* Keyboard‑friendly navigation; focus ring visible; reduced motion option.

**Screens**

1. **Home**: continue last lesson, daily goal, quick links to modes.

2. **Language Hub**: pick C++/Rust/Python/JS/YAML; show progress and badges.

3. **Lesson Player**: split view (reference code left, typing pane right) or overlay ghost text; token highlight; optional hints.

4. **Drill Player**: timer HUD (toggleable), progress bar, error markers, mini‑map.

5. **Zen Mode**: clean editor, no HUD; minimal chrome; optional ambient sounds.

6. **Scoreboard**: composite score, tWPM, accuracy, syntax bars, heatmap, error hotspots, percentile, recommended next steps; share card.

7. **Leaderboards**: tabs (Global/Friends/Org), filters (language/mode/period), anti‑cheat notice.

8. **History/Insights**: trends, streaks, time‑of‑day performance, key heatmap.

9. **Settings**: keyboard layout, indentation, themes, privacy, data export.

10. **Admin CMS**: lesson builder with token coverage checklist and YAML validator.

## **13\. Scoring & Anti‑Cheat**

* Detect paste events, unrealistic KPS spikes, window blur focus losses, and auto‑type patterns.

* Flagged sessions appear with an icon; excluded from public leaderboards.

* Optional webcam/HID verification (opt‑in) for tournaments.

## **14\. Accessibility & Internationalization**

* WCAG 2.2 AA targets, ARIA roles on editor chrome; color‑blind safe palettes.

* Keyboard‑only operability; skip‑to‑content shortcut.

* Localization pipeline (i18n keys, ICU message format), RTL mirroring.

## **15\. Performance & Offline**

* PWA installable; lesson cache via Workbox; optimistic session buffering.

* Web Workers for parsing; WASM parsers lazy‑loaded per language.

* Target: TTI \< 2.5s on mid‑range laptop; keystroke latency \< 8ms.

## **16\. Security & Privacy**

* Anonymous mode; only device‑local storage unless user opts in.

* End‑to‑end TLS; JWT short‑lived tokens; CSRF protection.

* GDPR‑friendly data export & deletion; event minimization.

## **17\. Tech Stack**

* **Frontend**: React \+ TypeScript, Zustand/Redux, Monaco or CodeMirror 6 (read‑only target text \+ custom overlay), Tailwind for styling.

* **Parsing**: Tree‑sitter (WASM) for C++/Rust/Python/JS; YAML: `yaml`/`js-yaml` for validation.

* **Backend**: Go (Gin/Fiber), Postgres, Redis, ClickHouse (optional), OpenTelemetry.

* **CI/CD**: GitHub Actions; Docker; deploy to Cloud Run/App Engine; Terraform IaC.

## **18\. Testing Strategy**

* **Unit**: parsing alignment, metrics calculations, scoring.

* **Property‑based**: random code generators for delimiter and token tests.

* **Integration**: session → scoring → leaderboard pipeline.

* **E2E**: Playwright/Cypress scripted user journeys.

* **Accessibility**: axe‑core audits; keyboard traps.

* **Load**: ingest peak at 500 events/sec/user, 5k concurrent users.

## **19\. Analytics (Privacy‑Preserving)**

* Local insights by default; opt‑in telemetry sends aggregated, anonymized metrics.

* Feature flags & A/B tests for scoring weights and UI variants.

## **20\. Release Plan & Roadmap**

**MVP (8–10 weeks)**

* Languages: Python, JavaScript, YAML minimal.

* Modes: Lesson, Timed Drill, Zen.

* Metrics: CPM, raw accuracy, basic scoreboard.

* Offline cache, anonymous auth, local history.

**v1.0**

* Add C++ and Rust; token metrics; leaderboards; assessments; CMS.

**v1.1+**

* AST‑shape scoring; advanced anti‑cheat; org leaderboards; exports.

## **21\. Acceptance Criteria (Samples)**

* FR‑1: User can complete all lessons in each language; token coverage ≥ 95% of defined syllabus.

* FR‑2: During a 5‑min drill, HUD updates ≤ 200ms latency; final metrics match recompute within 0.5%.

* FR‑3: Scoreboard renders within 500ms post‑submit; shows all defined metrics & percentile.

* FR‑4: Zen Mode hides all metrics/HUD; post‑session summary is off by default.

## **22\. Open Questions & Risks**

* Tree‑sitter grammar performance for very large snippets; fallback to line‑diff.

* YAML whitespace sensitivity vs. user indentation preferences—final rules.

* Anti‑cheat thresholds tuning; potential false positives.

## **23\. Appendix A — Sample Scoring Pseudocode**

function computeScore(metrics) {  
  const { tWPM, rawAcc, synAcc, errWeight, backspaceRate, idleMs } \= metrics;  
  const S \= 2.0\*tWPM \+ 50\*rawAcc \+ 60\*synAcc \- 0.3\*errWeight \- 5\*backspaceRate \- 0.1\*(idleMs/1000);  
  return Math.max(0, Math.round(S));  
}

## **24\. Appendix B — Event Schema (Client → Server)**

{  
  "sessionId": "uuid",  
  "ts": 1734372940123,  
  "events": \[  
    {"t":12, "k":"a", "act":"down", "pos":123},  
    {"t":14, "k":"Backspace", "act":"down", "pos":122, "err":true}  
  \],  
  "view": {"w": 1280, "h": 800, "dpi": 2}  
}

## **25\. Appendix C — YAML Validation Rules (Draft)**

* Disallow tab characters; enforce 2 or 4 space indentation per setting.

* Validate keys for duplicates and trailing spaces.

* Optionally validate against a provided JSON Schema.

---

**End of Specification**

