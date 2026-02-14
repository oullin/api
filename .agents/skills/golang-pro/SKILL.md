---
name: golang-pro
description: Use for Go design, implementation, refactoring, testing, and performance work. Includes concurrency, API design, observability, and production-readiness guidance for modern Go services.
license: MIT
---

# Golang Pro

Pragmatic guidance for building reliable Go services and tools.

## Scope

This skill is maintained in `/Users/gustavo/Sites/partners-api/.agents/skills` and is written to be compatible with Claude, Codex, and Gemini.

## When to Use

- Building or refactoring Go services, CLIs, or libraries
- Designing APIs, interfaces, and package boundaries
- Fixing concurrency, latency, or memory issues
- Improving test quality and release confidence

## Compatibility Rules

- Use plain Markdown and shell examples only
- Avoid model-specific tool calls or proprietary directives
- Keep guidance portable across Claude, Codex, and Gemini

## Workflow

1. Understand constraints: runtime, throughput, error budget, deployment target
2. Design clear contracts first: interfaces, request/response models, error model
3. Implement idiomatic Go: explicit errors, context propagation, small packages
4. Validate with tests and profiling before proposing optimisation work
5. Deliver with operational basics: logging, metrics, health checks, graceful shutdown

## Engineering Standards

### Must
- Run `gofmt` and `go test ./...`
- Use `context.Context` for I/O and blocking work
- Wrap errors with `%w` and preserve root causes
- Add table-driven tests for business logic
- Document exported types and functions

### Must Not
- Ignore returned errors without a clear reason
- Start goroutines without lifecycle control
- Use `panic` for expected runtime failures
- Mix transport, domain, and persistence concerns in one package

## Delivery Format

For substantial changes, provide:

1. Interface and data model design
2. Implementation changes
3. Tests (including edge cases)
4. Notes on risk, performance, and rollback
