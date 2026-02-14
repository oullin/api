---
name: postgresql-expert
description: Use for PostgreSQL schema design, query tuning, indexing, migrations, transactional safety, and operational reliability.
license: MIT
---

# PostgreSQL Expert

Focused guidance for designing and operating PostgreSQL-backed systems safely at scale.

## Scope

This skill is maintained in `/Users/gustavo/Sites/partners-api/.agents/skills` and is written to be compatible with Claude, Codex, and Gemini.

## When to Use

- Designing relational schemas and constraints
- Writing and tuning SQL queries
- Building safe migration plans
- Investigating locks, deadlocks, and performance regressions

## Compatibility Rules

- Use plain Markdown and SQL/shell examples only
- Avoid model-specific tool calls or proprietary syntax
- Keep guidance portable across Claude, Codex, and Gemini

## Workflow

1. Capture workload shape: read/write ratio, query patterns, cardinality
2. Design schema with constraints first: primary keys, foreign keys, uniqueness, checks
3. Add indexes from real query paths, not assumptions
4. Plan reversible migrations with clear backward-compatibility steps
5. Validate with `EXPLAIN (ANALYSE, BUFFERS)` and representative data

## Engineering Standards

### Must
- Use explicit transaction boundaries
- Add foreign keys and required uniqueness constraints
- Prefer `NUMERIC` for monetary values
- Review execution plans for critical queries
- Include rollback-safe migration steps

### Must Not
- Use `SELECT *` in performance-sensitive paths
- Create redundant indexes without workload evidence
- Perform destructive schema changes without a transition plan
- Ignore lock impact during large migrations

## Delivery Format

For substantial changes, provide:

1. Proposed schema or SQL changes
2. Migration plan with rollback path
3. Query plan observations and index rationale
4. Operational risks and mitigation steps
