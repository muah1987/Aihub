# Aihub Improvement Plan

## Completed

### Phase 1: Code Quality Fixes (2026-03-14)
- [x] Fix provider validation stub returning false `valid: true` for unsupported providers
- [x] Fix swallowed error in `provider/handler.go` GetByID call
- [x] Fix swallowed `json.Unmarshal` error in `integration/service.go` retry logic
- [x] Fix swallowed `io.ReadAll` error in `integration/service.go` HTTP response
- [x] Fix swallowed `json.Marshal` error in `integration/service.go` payload formatting
- [x] Fix swallowed `CountUnread` error in `notification/handler.go` list endpoint
- [x] Fix swallowed `CountUnread` return in `notification/handler.go` WebSocket connect
- [x] Fix swallowed `Deploy` error in `webhook/handler.go` with proper logging
- [x] Add `log` import to webhook handler for error logging

### Phase 2: Developer Infrastructure (2026-03-14)
- [x] Create `.github/copilot-instructions.md` with coding standards and changelog/memory rules
- [x] Create `.github/changelogs/` folder for tracking changes
- [x] Create `.github/memory/` folder for agent memory system
- [x] Create `.github/spec/` folder with this improvement plan
- [x] Update `.gitignore` to keep `.github/` tracked but exclude from builds
- [x] Create `.dockerignore` to exclude development-only files from Docker builds

## Planned

### Phase 3: Testing Infrastructure
- [ ] Add Go unit tests for critical service packages (auth, provider, integration)
- [ ] Add frontend test infrastructure (Vitest)
- [ ] Add CI/CD workflow with GitHub Actions

### Phase 4: Security Hardening
- [ ] Implement proper WebSocket origin checking using CORS configuration
- [ ] Add SSH host key pinning for production deployments
- [ ] Add rate limiting to webhook endpoints

### Phase 5: Observability
- [ ] Add structured logging (replace `log.Printf` with structured logger)
- [ ] Add request tracing / correlation IDs
- [ ] Add health check endpoints
