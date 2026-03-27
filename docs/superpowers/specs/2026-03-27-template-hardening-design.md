# Template Hardening Design

Date: 2026-03-27

## Summary

This project will remain a Go + Gin starter, but it will be clarified and hardened into two layers:

1. Template foundation
2. Example modules

The template foundation is the reusable part: bootstrapping, configuration, optional dependency handling, security boundaries, health checks, observability, deployment assets, and documentation. The example modules are the current `auth`, `user`, `tasks`, and `admin` flows. They stay in place as reference business modules so the repository is still runnable immediately after clone.

The goal is not to replace the current sample application. The goal is to make the sample application sit on top of a foundation that is truly safe to reuse and consistent with the documentation.

## Goals

- Make the repository genuinely usable as an out-of-the-box starter.
- Keep the current sample modules so new users can run and inspect a working backend immediately.
- Ensure optional dependencies are truly optional.
- Remove mismatches between code, deployment assets, and documentation.
- Fix shutdown, auth, and configuration issues that block a production-grade claim.
- Add regression coverage for the behaviors being corrected.

## Non-Goals

- Rewriting the repository into a framework.
- Removing the sample modules.
- Renaming the public API surface unless required for correctness.
- Adding major new infrastructure beyond what already exists.

## Current Problems

### 1. Optional dependencies are not actually optional

`README.md` and `.env.example` say Redis is optional, but startup always constructs a Redis-backed refresh token store. That means a user following the docs can still end up with a startup failure.

### 2. Database migrations can fail on a fresh database

The SQL migrations use `gen_random_uuid()` but only enable `uuid-ossp`. A clean Postgres instance may not have the function needed by the migrations.

### 3. Shutdown behavior is unsafe

Kafka outbox shutdown is not idempotent. The current startup and shutdown flow can close the outbox twice, which is a production reliability issue.

### 4. Security behavior is inconsistent

User banning exists in the admin API but is not enforced in login, refresh, or request authorization. IP whitelist and blacklist examples use CIDR notation, but the middleware only performs exact string matches.

### 5. Deployment assets and docs are inconsistent

The repository contains example deployment files, but some of them reference files that do not exist or environment variables that are not consumed by the application. That undermines the “production-ready” claim.

### 6. Template positioning is unclear

The codebase acts like both a reusable starter and a specific task-management backend. Without clear positioning, users inherit business-specific assumptions that should instead be presented as examples.

## Design

### A. Separate the template into foundation and examples

The implementation will keep the current package structure, but the documentation and naming will clearly distinguish:

- foundation: config, database, middleware, auth plumbing, health checks, logging, metrics, publishers, startup lifecycle
- examples: handlers, services, repositories, DTOs, and models that represent the sample app domain

This keeps implementation churn low while making the repository easier to understand and reuse.

### B. Make Redis truly optional

Refresh token storage will move behind a small store interface with at least two implementations:

- Redis-backed store
- In-memory store

Startup will select Redis when `REDIS_ADDR` is set and reachable; otherwise it will use the in-memory fallback with explicit logs. This matches the current captcha behavior and makes the documented “optional Redis” path real.

Expected behavior:

- local minimal mode: app can start with Postgres only
- production mode: Redis remains recommended for shared state and token revocation durability

The documentation will explicitly explain the tradeoff: in-memory refresh tokens are suitable for single-process/local use, not for horizontally scaled production deployments.

### C. Fix migration compatibility

UUID generation will be aligned consistently across models, migration SQL, and initialization scripts. The repository should work on a fresh Postgres instance without manual extension debugging.

The preferred outcome is a single UUID strategy used everywhere, with migrations and setup scripts reflecting the same choice.

### D. Make shutdown safe and predictable

The outbox publisher and async publisher close paths will be made idempotent. Main startup and shutdown sequencing will ensure each resource is closed exactly once, even when shutdown logic is triggered from multiple places.

Expected behavior:

- no panic during shutdown
- final outbox flush still happens
- the application can safely stop in local and production modes

### E. Enforce account status in auth flows

User status checks will be enforced in the following places:

- login
- refresh token
- authenticated request handling, where needed to prevent already-banned users from continuing to act

The admin ban/unban endpoints will then reflect actual access control behavior instead of metadata only.

### F. Support CIDR-based IP controls

IP whitelist and blacklist middleware will accept:

- single IPs
- CIDR ranges

This aligns runtime behavior with the example environment files and deployment docs.

### G. Align deployment assets and docs

The repository’s deployment story will be made internally consistent:

- `docker-compose.prod.yml` will reference assets that actually exist in the repository
- Nginx integration will remain optional and explicitly documented as example infrastructure
- Render docs and config will describe the environment variables the application truly uses
- quickstart and deployment docs will be updated to reflect the actual dependency model

### H. Clarify the template promise

README and related docs will explicitly state:

- this repository is a production-oriented starter
- the sample business modules are examples
- the reusable value lives in the foundation layer
- replacing the sample modules is a normal intended workflow

## Testing Strategy

The changes will be validated with tests first where behavior is changing.

New or expanded coverage should include:

- refresh token store selection and fallback behavior
- IP whitelist and blacklist CIDR matching
- banned user rejection during login and refresh
- auth middleware rejection for non-active users when applicable
- outbox close behavior is safe when called repeatedly

Existing repository tests will remain and should still pass.

## Documentation Changes

The following docs will be updated:

- `README.md`
- `docs/QUICKSTART.md`
- `docs/DEPLOYMENT.md`
- environment examples
- deployment asset references where needed

Key wording changes:

- “Redis optional” will be qualified correctly
- “production-ready” claims will be backed by actual behavior
- sample modules will be identified as examples

## Success Criteria

This work is complete when all of the following are true:

- The app starts successfully in a Postgres-only local setup.
- Redis and Kafka are optional in both code and docs.
- Fresh database setup does not fail because of UUID extension mismatch.
- Ban and unban status materially affect authentication behavior.
- CIDR examples in environment files work in runtime behavior.
- Shutdown does not panic when Kafka/outbox is enabled.
- Production compose and optional Nginx setup reference real files.
- README and deployment docs match actual runtime behavior.
- Verification commands pass after the changes.

## Risks and Mitigations

- Risk: changing auth behavior could break existing sample flows.
  Mitigation: add focused tests around login, refresh, and middleware behavior.

- Risk: changing token store behavior could weaken production assumptions.
  Mitigation: keep Redis as the recommended production path and document in-memory limitations clearly.

- Risk: touching startup lifecycle could introduce regressions.
  Mitigation: keep changes localized, add close-path tests, and verify with full build/test runs.

## Recommended Execution Order

1. Introduce tests for the broken behaviors.
2. Refactor refresh token storage to support optional backends.
3. Fix account-status enforcement in auth flows.
4. Fix CIDR IP controls.
5. Fix migration UUID alignment.
6. Fix outbox shutdown idempotency.
7. Align deployment assets and documentation.
8. Run full verification and update final positioning language.
