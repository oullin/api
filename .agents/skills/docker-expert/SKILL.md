---
name: docker-expert
description: Use for Dockerfile design, container hardening, image optimisation, compose orchestration, and container runtime troubleshooting.
license: MIT
---

# Docker Expert

Practical guidance for secure, reproducible, and efficient container workflows.

## Scope

This skill is maintained in `/Users/gustavo/Sites/partners-api/.agents/skills` and is written to be compatible with Claude, Codex, and Gemini.

## When to Use

- Creating or refactoring Dockerfiles
- Reducing image size and build duration
- Hardening containers for production
- Fixing compose networking, volume, or startup issues

## Compatibility Rules

- Use plain Markdown and shell examples only
- Avoid model-specific commands or workflow assumptions
- Keep instructions portable across Claude, Codex, and Gemini

## Workflow

1. Inspect current Dockerfile, compose files, and runtime assumptions
2. Separate build and runtime concerns with multi-stage builds
3. Reduce attack surface: minimal base image, least privilege, pinned dependencies
4. Improve reproducibility: deterministic build inputs and explicit versions
5. Validate build, startup, health checks, and shutdown behaviour

## Engineering Standards

### Must
- Use `.dockerignore` to keep build context small
- Prefer multi-stage builds for compiled applications
- Run as non-root where possible
- Add health checks when services expose liveness endpoints
- Keep runtime images minimal and explicit

### Must Not
- Bake secrets into images
- Use broad `COPY . .` too early in the Dockerfile
- Depend on mutable tags without justification
- Ignore container limits and restart strategy

## Delivery Format

For substantial changes, provide:

1. Revised Dockerfile or compose configuration
2. Security and optimisation rationale
3. Validation commands and expected results
4. Operational risks and rollback notes
