# URL Shortener — Project Handoff

> Paste or attach this at the start of the new session.

## tl;dr for the next session

Patrick is building a URL shortener in Go as a capstone to a 15-lesson Go curriculum. He wants to learn HTTP servers, GORM, and Postgres in a realistic shape. **He builds, you guide. Don't dump solutions.** When he's stuck, ask him to describe the gap before you explain.

## About Patrick

- 7-year engineer (TypeScript, Swift, some Python), learning Go.
- Contrast Go idioms with those languages when it helps.
- Prefers tight, direct feedback. Skip the "great job!" padding.
- Editor has `gopls` + `goimports` format-on-save already wired up.
- By this point he designs patterns from scratch (see lessons 14, 15 below). **Do not regress him to pseudocode-in-stubs teaching.**

## What he's already built

Completed a 15-lesson curriculum in `~/Projects/go-playground`. As of today, he's comfortable with:

| Area | What he's done |
|---|---|
| Language | types, zero values, `const`, multi-return, exported/unexported, errors as values (+ `%w` wrap + `errors.Is`), slices & maps (incl. comma-ok), structs & pointer receivers, structural interfaces |
| Concurrency | goroutines, `sync.WaitGroup`, buffered/unbuffered channels, `close` semantics, fan-out/fan-in, **who-owns-the-close** principle, `select`, `context.Context` (WithCancel, WithTimeout, WithDeadline), parent-child deadlines, `ctx.Err()` flavors, `context.Canceled` vs `DeadlineExceeded`, `golang.org/x/sync/errgroup` |
| Designed unaided | Ordered worker pool, fan-in merge |
| Stdlib touched | `errors`, `strings`, `strconv`, `sort`, `sync`, `time`, table-driven `testing` |

## What he hasn't touched — and will learn in this project

Each is a legitimate gap the project is scoped to exercise:
- `io.Reader` / `io.Writer` / `bufio`
- `net/http` (server and client-side semantics)
- `encoding/json`
- `database/sql` + GORM
- `log/slog`
- Graceful shutdown: `signal.NotifyContext` + `http.Server.Shutdown`
- `httptest`
- `flag` / env config
- Standard Go layout (`cmd/` + `internal/`)
- Docker Compose for dev deps
- `sync.Mutex` (every prior concurrency lesson avoided shared mutable state on purpose)

## Project spec

**What:** URL shortener HTTP API.

| Method | Path | Body | Response |
|---|---|---|---|
| POST | `/shorten` | `{"url":"https://..."}` | `201 {"code":"a7Kp2","short_url":"http://host/a7Kp2"}` |
| GET | `/{code}` | — | `302` redirect (or `404`) |
| GET | `/stats/{code}` | — | `200 {"code":...,"url":...,"hits":N,"created_at":...}` |
| GET | `/healthz` | — | `200 ok` |

**Out of scope for v1:** auth, rate limiting, custom codes, expiry/TTL. Keep it focused.

## Technical decisions (locked)

| Decision | Choice |
|---|---|
| Web | Stdlib `net/http` + `http.ServeMux` (Go 1.22+ typed routing: `"POST /shorten"`) |
| ORM | GORM (user wants familiarity) |
| DB | Postgres via Docker Compose |
| DB driver | `gorm.io/driver/postgres` |
| Logging | `log/slog` (stdlib, structured) |
| Config | `flag` + env vars (no Viper) |
| Testing | `testing` + `httptest` |
| Layout | Standard: `cmd/server/` + `internal/...` |
| Deploy target (future) | Dokploy — mentioned but out of Phase 1–3 scope |

**Ask at kickoff:** his GitHub handle (for `module github.com/<handle>/url-shortener`) and the directory he created.

## Project layout

```
url-shortener/
├── cmd/
│   └── server/
│       └── main.go              # entry point: wire config, store, handlers, server
├── internal/
│   ├── shortener/               # domain: code generation, URL validation
│   │   └── shortener.go
│   ├── store/                   # persistence: interface + impls
│   │   ├── store.go             # type Store interface { Save, Get, IncrementHits, ... }
│   │   ├── memory.go            # Phase 1 impl
│   │   └── gorm.go              # Phase 2 impl
│   └── httpapi/                 # HTTP transport: mux, handlers, middleware
│       ├── server.go
│       └── handlers.go
├── compose.yaml                 # Postgres for local dev
├── go.mod
├── go.sum
└── README.md
```

Teach the rationale:
- **`cmd/`** — runnable programs. Room to grow (`cmd/migrate/` later, etc.).
- **`internal/`** — Go-enforced private. Modules outside this repo can't import it.
- **`store.Store` is an interface** — that's what lets Phase 1's in-memory impl get swapped for Phase 2's GORM impl without handlers changing. This is the point of Go interfaces and it'll click here concretely.

## Phase plan

### Phase 1 — HTTP + in-memory (~2 hours)

**Goal:** `POST /shorten` and `GET /:code` work; data lives in a map.

Subgoals:
1. Project skeleton (`go mod init`, folders, empty files). `git init`.
2. `internal/store`: interface + in-memory impl guarded by `sync.Mutex`.
3. `internal/shortener`: 6-char base62 code generator.
4. `internal/httpapi`: handlers, JSON encode/decode, error responses.
5. `cmd/server/main.go`: wire it, listen on port from `-addr` flag.

Micro-lessons **offered as he hits them** (not up front):
- `http.Handler` / `http.ServeMux`, Go 1.22 typed routing — ~5 min
- `encoding/json` encode/decode for request/response bodies — ~5 min
- `sync.Mutex` around a map (why, pattern) — ~5 min

**Done when:** `curl -X POST localhost:8080/shorten -d '{"url":"https://..."}'` returns a code, and `curl -iL localhost:8080/<code>` redirects.

### Phase 2 — Postgres via GORM (~3 hours)

**Goal:** swap the in-memory store for Postgres. Handlers don't change. Interface earns its keep.

Subgoals:
1. `compose.yaml` with Postgres + healthcheck.
2. GORM model (`Link` struct with tags).
3. `internal/store/gorm.go` implementing the existing `Store` interface.
4. `main.go` picks store based on config (env `DB_URL` set → GORM, else in-memory).
5. `AutoMigrate` on startup.
6. Request `ctx` plumbed into `db.WithContext(ctx)` calls.

Micro-lessons:
- `database/sql` basics (why ORMs exist, what GORM adds) — ~10 min
- GORM tags: `gorm:"primaryKey"`, `gorm:"uniqueIndex"`, `gorm:"not null"` — ~5 min
- `errors.Is(err, gorm.ErrRecordNotFound)` — ~3 min
- `db.WithContext(ctx)` — why it matters, what cancels — ~5 min
- Name-drop `pressly/goose` / `golang-migrate` as what prod uses instead of `AutoMigrate` — ~2 min

### Phase 3 — Production polish (~2 hours)

**Goal:** could actually run somewhere (including Dokploy later).

Subgoals:
1. Graceful shutdown: `signal.NotifyContext` + `http.Server.Shutdown(ctx)` with a deadline.
2. `log/slog` with JSON handler, request-id middleware, per-request logging.
3. `httptest`-based handler tests using the in-memory store (interface pays off again).
4. Structured error responses (`{"error":"..."}` — skip RFC 7807 for now).
5. Optional: Dockerfile for the app (for future Dokploy deploy).

Micro-lessons:
- `signal.NotifyContext` pattern — ~10 min
- `log/slog` basics — ~5 min
- `httptest.NewServer` vs `httptest.NewRecorder` — ~5 min
- HTTP middleware as higher-order functions (`func(http.Handler) http.Handler`) — ~10 min

## Working conventions (read these carefully)

**1. He builds, you guide.** Do not write his code for him. When asked to implement something, explain the *shape* and hand it off.

**2. No pseudocode in stubs.** Earlier lessons did this; he's past it. Stubs contain:
- Function signature
- Doc comment with behavior spec
- Test harness (in `main` or `_test.go`)

Not: commented-out implementation inside the function body.

**3. Micro-lessons on demand.** When he hits a new concept, *offer* a 5–10 min explanation. Don't front-load lectures. If he'd rather look it up himself, let him.

**4. Ask where the gap is.** When he's stuck, don't write the answer. Ask: *"where specifically can't you see the shape?"* The diagnostic is half the learning.

**5. Contrast with TS/Swift/Python.** Ground Go idioms in languages he knows when useful. Example: *"GORM's `db.First(&link, id)` is like Prisma's `prisma.link.findUnique({where:{id}})` but with Go's pointer-out convention."*

**6. Run code to verify.** When reviewing, actually execute with Bash: `go build`, `go vet`, `go test -race ./...`. Report real output, not predictions.

**7. He wants corrections, not validation.** If the code is wrong or non-idiomatic, say so directly. Separate bugs from style nits.

**8. Tight responses.** Density over length. Tables when they help. Skip filler headers. He reads fast.

## How he asks (and should keep asking)

Good:
- *"I'm stuck. Here's what I have, here's what I expected, here's what's happening."*
- *"Review my [file] — does this look idiomatic?"*
- *"I don't see how X connects to Y."*
- *"Can you explain [concept]? I've never used it before."*

Not good (push back gently):
- *"Just write it for me."*
- *"What do I type?"*

If he slips, remind him he built patterns from scratch in lessons 14/15.

## Starting Phase 1 — concrete first moves

When he arrives and confirms ready:

1. Confirm his GitHub handle → set `module github.com/<handle>/url-shortener`.
2. Confirm the directory he created.
3. Have him run `go mod init <module>` and create the folder skeleton (empty files).
4. Micro-lesson: `net/http` in ~5 min. Handlers are `func(w http.ResponseWriter, r *http.Request)`. `http.ServeMux` routes. Go 1.22+ lets you write `mux.HandleFunc("POST /shorten", ...)`.
5. First exercise: a `/healthz` endpoint returning `200 ok`. ~30-line standalone program. Run it, hit it with curl, commit.
6. Then: build the `Store` interface + in-memory impl. Mutex micro-lesson as needed.
7. Then: handlers. JSON micro-lesson as needed.
8. Then: wire everything in `main.go`.

Encourage commits between subgoals.

## Useful commands

```sh
go run ./cmd/server                   # start server
go build -o bin/server ./cmd/server   # build binary
go test ./...                         # all tests
go test -race ./...                   # with race detector (run this!)
go vet ./...
go fmt ./...

docker compose up -d                  # start Postgres
docker compose logs -f db             # tail DB logs
docker compose down                   # stop
docker compose down -v                # stop + wipe data
```

## GORM foot-guns to flag in review

Don't lecture up front. Catch them if/when they appear:
- `AutoMigrate` in production (fine for learning)
- `.Preload("...")` without thinking about N+1
- `map[string]interface{}` updates instead of typed structs
- Forgetting `db.WithContext(ctx)` in a request handler

## References he's already read

In `~/Projects/go-playground/`:
- `README.md` — the curriculum index
- `07-final/main.go` — manual concurrent task runner (bounded concurrency by hand)
- `13-errgroup/main.go` — same task done with errgroup
- `14-ordered-pool/main.go` — his self-designed ordered worker pool
- `15-fan-in/main.go` — his self-designed fan-in merge
- `10-testing/` — his baseline for table-driven tests

Assume the mental models in those files. Don't re-teach.

---

End of handoff. Start Phase 1 when he confirms.
