# Copilot Instructions for Aihub

## Project Overview

Aihub is an AI-native web platform for managing multi-agent AI workflows, team orchestration, VPS deployment, and real-time collaboration.

- **Backend**: Go 1.24 with Chi router, GORM ORM, PostgreSQL 16
- **Frontend**: React 19, TypeScript 5.9, Vite 7, Tailwind CSS v4
- **Real-time**: gorilla/websocket for chat, terminal, and notifications

## Changelog Management

All changes must be logged in `.github/changelogs/`. Follow these rules:

1. **Create a new changelog entry** for every meaningful change (bug fix, feature, refactor, security fix).
2. **File naming**: `YYYY-MM-DD_<short-description>.md` (e.g., `2026-03-14_fix-error-handling.md`).
3. **Format** each changelog entry as follows:

```markdown
# Changelog: <Short Title>

**Date**: YYYY-MM-DD HH:MM UTC
**Agent/Author**: {agentuuid} or human username
**Type**: fix | feature | refactor | security | docs | chore

## Changes
- Description of what changed

## Files Modified
- path/to/file.go

## Reason
Why this change was made.
```

4. **Review changelogs** before starting any session to understand recent changes and avoid duplicating or conflicting work.

## Memory System

Agent memory is stored in `.github/memory/`. This system enables AI agents to learn from past sessions:

1. **Load memory** at the start of every session by reading all files in `.github/memory/`.
2. **Log every session** by creating or updating a memory file with:
   - Timestamp (ISO 8601)
   - Agent/AI identifier (`{agentuuid}` or `{aiuuid}`)
   - What was done
   - What was learned
   - Mistakes avoided (based on previous memory entries)
3. **File naming**: `YYYY-MM-DD_<agent-uuid>.md`
4. **Never delete** memory files — they form the learning history.

## Coding Standards

### Go (Backend)
- Always handle errors explicitly — never use `_` to discard errors unless the value is truly unused.
- Use `fmt.Errorf` with `%w` for error wrapping.
- Use `log.Printf` for logging errors in goroutines (where errors cannot be returned).
- Follow the existing `service.go` + `handler.go` pattern for new packages.
- Use UUID primary keys throughout.

### TypeScript/React (Frontend)
- Use functional components with hooks.
- Use Zustand for state management.
- Use Axios via the `client.ts` wrapper for API calls.
- Follow Tailwind CSS v4 utility classes for styling.

### Security
- Never return `valid: true` for unvalidated operations — be honest about validation status.
- Always validate WebSocket origins in production (configure via CORS middleware).
- Log deployment and webhook errors — never silently discard them.
- Use AES-256-GCM for token encryption.
- Validate SSH host keys in production deployments.

### General
- Check `.github/memory/` for past mistakes before making changes.
- Check `.github/changelogs/` for recent changes to understand context.
- Run `go build ./...` after backend changes.
- Run `cd web && npm run build` after frontend changes.
- Always check for existing error handling patterns before adding new code.

## Build & Ignore Rules

The following directories are **excluded from production builds** (see `.dockerignore` and `.gitignore`):
- `.github/changelogs/` — development-only changelogs
- `.github/memory/` — AI agent memory (development-only)
- `.github/spec/` — specification and planning docs

These directories exist only for GitHub-based AI agent workflows and self-improvement. They must never be included in Docker images or compiled binaries.
