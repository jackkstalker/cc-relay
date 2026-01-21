# cc-relay

## What This Is

A multi-provider proxy for Claude Code written in Go. It sits between Claude Code and multiple LLM providers
(Anthropic, Z.AI, Ollama), allowing you to access all their models from a single interface and switch between
them seamlessly. Personal tool for maximizing model availability and choice.

## Core Value

Access all models from all three providers (Anthropic, Z.AI, Ollama) in Claude Code and switch between them seamlessly.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Route requests to Anthropic provider
- [ ] Route requests to Z.AI provider
- [ ] Route requests to Ollama provider
- [ ] Implement basic routing strategy (round-robin or similar)
- [ ] Expose /v1/models endpoint listing all available models from all providers
- [ ] Create Claude Code skill that lists available models
- [ ] Preserve Anthropic API compatibility (SSE streaming, tool_use_id, etc.)
- [ ] Support model switching mid-session

### Out of Scope

- TUI dashboard with Bubble Tea — defer to v0.1.0+ (cool but not essential for core function)
- WebUI management interface — defer to v0.1.0+ (visual management not needed initially)
- Cloud providers (Bedrock, Azure, Vertex AI) — not needed for personal use case
- Advanced routing strategies (cost-based, latency-based, model-based) — defer to v0.1.0+ after basic routing proven
- Multi-key pooling and rate limiting — defer to v0.1.0+ (single key per provider sufficient initially)
- Hot reload configuration — defer to v0.1.0+ (restart acceptable for config changes)
- Prometheus metrics and observability — defer to v0.1.0+ (not critical for personal use)
- Circuit breakers and health tracking — defer to v0.1.0+ (manual failover acceptable initially)
- gRPC management API — defer to v0.1.0+ (not needed without TUI/WebUI)

## Context

**Personal Use Case:**

- Want to use cheaper Z.AI models for simple tasks
- Use Anthropic models for complex reasoning
- Use local Ollama for privacy/offline capability
- Need to see and switch between all available models easily

**Existing Work:**

- Comprehensive SPEC.md with 6-phase roadmap exists
- README.md and documentation written
- Architecture designed but not implemented
- v0.0.1 targets subset of Phase 1 from spec

**Technical Environment:**

- Go project using standard library
- Must be compatible with Claude Code (exact Anthropic API match)
- Development workflow uses go-task, air for live reload, lefthook for git hooks

## Constraints

- **API Compatibility**: Must exactly match Anthropic Messages API format — Claude Code expects specific headers,
  SSE event sequence, tool_use_id preservation
- **Tech Stack**: Go with standard library preferred — keep dependencies minimal for maintainability
- **Claude Code Integration**: Must work seamlessly as drop-in proxy (user sets ANTHROPIC_BASE_URL and it just works)

## Key Decisions

| Decision                     | Rationale                                                                 | Outcome   |
| ---------------------------- | ------------------------------------------------------------------------- | --------- |
| Config-based model awareness | Claude Code skill reads proxy config to discover available models         | Pending   |
| Simple routing for v0.0.1    | Ship basic round-robin first, defer smart routing until proven            | Pending   |
| Skill just lists models      | Defer auto-selection to v0.1.0+, keep v0.0.1 focused on visibility        | Pending   |
| Defer TUI/WebUI              | Focus on core proxy functionality, add visual management later            | Pending   |

## Last Updated

2026-01-20 after initialization
