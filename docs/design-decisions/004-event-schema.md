# ADR-004: Versioned Event Schema

## Status
Accepted

## Context
procscope produces machine-readable output (JSONL, evidence bundles) that may be consumed by other tools, scripts, or downstream analysis pipelines. Schema changes could break consumers.

## Decision
Explicitly version the event schema with a `schema_version` field in every event. Follow semantic versioning for the schema.

## Rationale
- **Forward compatibility:** Consumers can check the version and handle unknown fields gracefully
- **Stable contracts:** Teams building integrations can rely on documented field names and types
- **Change tracking:** Schema changes are versioned and documented

## Schema Rules
1. Every event includes `schema_version` (currently "1.0.0")
2. **Patch version bump:** Documentation fixes, new optional fields with defaults
3. **Minor version bump:** New event types, new optional fields
4. **Major version bump:** Breaking changes to existing field names, types, or semantics
5. Removed fields are never reused with different semantics

## Consequences
- `MarshalJSON` always injects `schema_version`
- Downstream tools should be forgiving of unknown fields
- Schema changes require CHANGELOG entry
- Version is hardcoded in `events.SchemaVersion` constant
