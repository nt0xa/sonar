# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Behavior

### Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

### Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

### Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

### Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.


## Project Overview

Sonar is a security researcher's tool for finding and exploiting vulnerabilities that require out-of-band interactions. Written in Go, it provides both a server and CLI client for capturing and managing interactions across multiple protocols (DNS, HTTP, SMTP, FTP).

Documentation: https://nt0xa.github.io/sonar/

## Development Environment Details

There is fully configured development environment available in `dev/docker-compose.yml`. To started use `make up` command.

**Docker Compose Services** (`dev/docker-compose.yml`):

- `postgres`: Database (port 5432)
- `pebble`: Let's Encrypt mock for testing ACME
- `dev`: Main development container with Go tooling, runs `server` binary with hot reload via `air`
- `docs`: Docusaurus docs site (port 3000)

**Development Tools** (installed via `make devtools`):

- `migrate` - Database migrations
- `air` - Hot reload for Go
- `go-enum` - Enum generation
- `mockery` - Mock generation
- `goreleaser` - Release building

## Build and Development Commands

:warning: Most of the commands in `Makefile` are executed inside development container. Make sure to run `make up` before doing anything else.

### Docker Development Environment

```bash
make up                # Start all services (postgres, pebble, dev, docs)
make down              # Stop all services
make restart           # Restart dev container
make recreate          # Recreate dev container
make enter             # Open bash shell in dev container
make watch             # Auto-rebuild and restart on code changes
```

All commands are docker-aware: they run inside the dev container if not already inside Docker.

### Build

```bash
make build              # Build both server and client
make build/server       # Build server only (output: build/server)
make build/client       # Build CLI client only (output: build/sonar)
make clean/build        # Remove build artifacts
```

### Testing
```bash
make test              # Run all tests with coverage (sequential: -p 1)
make coverage          # Open coverage report in browser

# Tests require PostgreSQL connection via SONAR_DB_DSN environment variable
# Example: export SONAR_DB_DSN='postgres://db:db@localhost:5432/db_test?sslmode=disable'
```

### Code Quality
```bash
make fmt               # Format code with go fmt
make lint              # Run golangci-lint
```

### Code Generation
```bash
make generate                # Generate all (mocks)
make generate/mocks          # Generate test mocks with mockery (service.ServerService)
make generate/db             # Generate database access code with sqlc
```

The API, CLI, and API client are hand-written (no codegen). Enum types in
`internal/service/models.go` are generated via `go-enum` (`go generate ./...`).

### Database Migrations
```bash
make migrations/create NAME=description  # Create new migration files
make migrations/up N=1                   # Apply N migrations
make migrations/down N=1                 # Revert N migrations
make migrations/goto V=version           # Go to specific version
make migrations/force V=version          # Force version (recovery)
make migrations/drop                     # Drop everything (dangerous)
```

### Release
```bash
make release/snapshot  # Create snapshot release with goreleaser
```

## Architecture

### High-Level Design

Sonar uses an **event-driven architecture** where protocol handlers (DNS, HTTP, SMTP, FTP) capture interactions and emit events through a buffered channel. A worker pool (default: 10 goroutines) processes events: storing to database, triggering notifications, and executing post-processing.

```
Protocol Handlers → Event Channel → Worker Pool → [DB Storage, Notifiers, Processing]
     ↓                                                           ↓
  DNS:53                                                   PostgreSQL
  HTTP:80                                                  Telegram/Slack/Lark
  HTTPS:443
  SMTP:25
  FTP:21
```

### Key Components

**`/code/cmd/`** - Entry points
- `server/main.go` - Server binary
- `client/main.go` - CLI client binary (builds the `internal/cmd` tree over a `remotesvc` client)

**`/code/internal/database/`** - Data persistence layer
- `models/` - Domain models (User, Payload, Event, DNSRecord, HTTPRoute)
- `migrations/` - Sequence-numbered SQL migrations (e.g., 000001_initial_schema.up.sql)
- `db.go` - Database initialization, migration runner, connection pooling
- `fixtures/` - YAML test fixtures loaded via testfixtures library

**`/code/internal/service/`** - Business logic interface
- Defines `Service` (business operations) and `ServerService` (`Service` plus identity resolvers: `AuthContextBy*`, `LarkProvisionUser`)
- Per-operation `Input`/`Output` types (`Output` types are distinct named types, used for template dispatch)
- Uses standard Go `error` (`service.Error`) and validation via `pkg/valid`
- Subpackages:
  - `dbsvc/` - DB-backed implementation of `ServerService` (PostgreSQL)
  - `remotesvc/` - client implementation of `Service` over the HTTP API (stdlib only)
  - `auditsvc/` - decorator wrapping `ServerService` to write audit records
  - `mock/` - generated `ServerService` mock for tests
- Lets client (remote) and server (DB) share the same abstraction

**`/code/internal/cmd/`** - CLI command tree
- Builds the cobra command tree over `service.Service` (hand-written, closure-per-command)
- `server/` - server entrypoint package (config, module wiring, protocol handler setup)

**`/code/internal/modules/`** - Pluggable features
- `api/` - REST API server (custom `webx` router, bearer auth, JSON) over `service.ServerService`; `apimodels/` holds the request DTOs shared with `remotesvc`
- `telegram/`, `slack/`, `lark/` - Notification integrations; each builds an `internal/cmd` tree over `service.ServerService` and depends only on the service (not the DB)
- Each module implements `modules.Notifier` interface
- Modules receive events and decide whether to notify based on criteria

**`/code/pkg/`** - Protocol implementations and utilities
- `dnsx/` - DNS server implementation
- `httpx/` - HTTP/HTTPS server with middleware
- `smtpx/` - SMTP server
- `ftpx/` - FTP server
- `webx/` - HTTP router used by the API module
- `valid/` - validation helpers used by the service layer
- `certmgr/` - TLS certificate management (Let's Encrypt integration)
- `geoipx/` - GeoIP lookups for events
- `telemetry/` - OpenTelemetry setup (traces, metrics, logs)

### Critical Patterns

**Dependency Injection**: Components receive dependencies in constructors. Use interfaces (`service.Service`, `service.ServerService`, `modules.Notifier`) to enable testing with mocks.

**Configuration Management**:
- Multi-source loading: defaults → TOML file → environment variables
- Environment variable pattern: `SONAR_*` (e.g., `SONAR_DB_DSN`, `SONAR_DOMAIN`)
- CLI client config: `~/.config/sonar/config.toml` (XDG standard)

**Error Handling**:
- The service layer returns standard Go `error`; `service.Error` carries a kind (bad request, unauthorized, not found, conflict, validation, internal) that the API maps to HTTP status codes and `remotesvc` reconstructs from responses
- Validation errors track field-level issues via `pkg/valid` (`service.Validation`)

**Event Processing**:
- Buffered channel decouples protocol handlers from processing
- Worker pool prevents resource exhaustion
- Events stored in database with protocol, request/response bytes, metadata, geo info

**Database Observer Pattern**:
- Database layer supports observers that listen for entity changes
- Notifiers register as observers to react to new events/payloads

### Adding/Modifying an Operation

The API, CLI, and client are hand-written and must be kept in sync manually:
1. Add the method (+ `Input`/`Output` types) to the `Service`/`ServerService` interface in `internal/service/`
2. Implement it in `internal/service/dbsvc/` and `internal/service/remotesvc/`
3. Run `make generate/mocks` to refresh the `ServerService` mock
4. Add an HTTP handler + route in `internal/modules/api/` (and a request DTO in `apimodels/` if needed)
5. Add a CLI command in `internal/cmd/`
6. Add a render template for the new `Output` type in `internal/templates/` (type switch in `RenderResult`)

### Testing Strategy

**Unit Tests**: Co-located `*_test.go` files using testify assertions

**Integration Tests**:
- API tests in `internal/modules/api/api_test.go` run handlers against the `ServerService` mock via `httptest`
- `remotesvc` tests exercise the client against a mock-backed API server (`httptest`)
- Database-backed tests (e.g. `internal/service/auditsvc`) use testfixtures to load YAML fixtures
- Require PostgreSQL connection (set `SONAR_DB_DSN`)

**Test Fixtures**:
- Located in `internal/database/fixtures/`
- YAML format defines seed data (users, payloads, events, etc.)
- Reloaded per-test for isolation

**Test Tokens** (hardcoded for tests):
- AdminToken: `00112233445566778899aabbccddeeff`
- User1Token: Check fixture files
- User2Token: Check fixture files

**Running Single Tests**:
```bash
# Run specific test
go test ./internal/cmd -run TestCmd -v

# Run specific package
go test ./internal/modules/api -v

# Run with coverage
go test ./... -coverprofile=coverage.out
```

### Database Schema

**Core Models**:
- `users` - API token, admin flag, notification IDs (Telegram, Slack, Lark)
- `payloads` - Name, subdomain, notify protocols, store events flag
- `events` - UUID, protocol, request/response data, metadata (IP, geo, headers), timestamps
- `dns_records` - Name, type (A/AAAA/CNAME/etc.), TTL, values, strategy
- `http_routes` - Method, path, response (code, headers, body), dynamic flag

**Migration Format**:
- Sequence-numbered pairs: `000001_description.up.sql`, `000001_description.down.sql`
- Database name must end with `_test` for fixture loading (e.g., `db_test`, `postgres_test`)

### Protocol Handlers

Each protocol handler (`dnsx`, `httpx`, `smtpx`, `ftpx`) follows this pattern:
1. Listen on standard port
2. Parse incoming requests
3. Match against configured rules (DNS records, HTTP routes)
4. Emit event to channel
5. Return response

**DNS**: Supports A, AAAA, CNAME, MX, TXT, NS records with strategies (all, round-robin, rebind)

**HTTP**: Supports static routes and dynamic routes (wildcards), configurable responses

**SMTP**: Captures email interactions

**FTP**: Captures file transfer interactions

### TLS Certificate Management

- **Custom**: Load from files
- **Let's Encrypt**: Automatic ACME protocol integration
  - Uses `certmgr` package
  - Stores accounts in database via `certstorage` package
  - Dev environment uses Pebble (Let's Encrypt mock server)

## Important Architectural Decisions

1. **PostgreSQL as single source of truth** - All state persisted to database
2. **Worker pool for event processing** - Prevents resource exhaustion under load
3. **Interface-based design** - `service.Service`/`ServerService` abstracts DB-backed (`dbsvc`) vs remote (`remotesvc`) implementations
4. **Hand-written API/CLI/client** - kept in sync manually against the service interface (no codegen)
5. **Multi-notifier support** - Events can trigger multiple channels (Telegram + Slack + Lark)
6. **Let's Encrypt integration** - Production-ready TLS with auto-renewal
7. **OpenTelemetry observability** - Structured logging, distributed traces, metrics
8. **Buffered event channel** - Decouples protocol handling from event processing

## Common Workflows

### Adding a New Operation
1. Add the method (+ `Input`/`Output` types) to `Service`/`ServerService` in `internal/service/`
2. Implement it in `internal/service/dbsvc/` (and `internal/service/remotesvc/` for the client)
3. Run `make generate/mocks` to refresh the `ServerService` mock
4. Add the HTTP handler + route in `internal/modules/api/` (request DTO in `apimodels/` if needed)
5. Add the CLI command in `internal/cmd/` and a render template in `internal/templates/`
6. Add API tests in `internal/modules/api/api_test.go`

### Adding a New Protocol Handler
1. Create package in `pkg/` (e.g., `pkg/ldapx/`)
2. Implement handler with event emission
3. Register handler in server initialization
4. Add configuration options
5. Document protocol behavior

### Adding a New Notifier
1. Create package in `internal/modules/` (e.g., `internal/modules/discord/`)
2. Implement `modules.Notifier` interface
3. Add configuration options
4. Register in module registry
5. Add to `SONAR_MODULES_ENABLED` environment variable

### Debugging Tests
- Tests run sequentially (`-p 1`) to avoid database conflicts
- Check `SONAR_DB_DSN` is set correctly
- Ensure database name ends with `_test` for fixtures
- Use `go test -v` for verbose output
- Check `coverage.out` for coverage gaps

## Documentation

Documentation site source: `/code/docs/`
Built with: Docusaurus (React-based static site generator)

When developer environment is running, available on http://localhost:3000/sonar.
