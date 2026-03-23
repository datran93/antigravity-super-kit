# ⚡ Performance Checklist Template

> Domain: Performance Review
> Used by: `/checklist-generator` for performance audits

## Response Time

- [ ] [CL-PERF-001] API endpoints respond within specified latency targets (p50, p99)
- [ ] [CL-PERF-002] Database queries complete in < 100ms for standard operations
- [ ] [CL-PERF-003] No N+1 query patterns in data fetching code
- [ ] [CL-PERF-004] Pagination implemented for all list endpoints

## Resource Efficiency

- [ ] [CL-PERF-005] Database connections pooled (not opened per request)
- [ ] [CL-PERF-006] Expensive computations cached with appropriate TTL
- [ ] [CL-PERF-007] File/stream resources properly closed (no leaks)
- [ ] [CL-PERF-008] Memory usage bounded — no unbounded slice/list growth

## Concurrency

- [ ] [CL-PERF-009] Concurrent access patterns identified and tested
- [ ] [CL-PERF-010] Locks scoped minimally — no database-wide locks for row operations
- [ ] [CL-PERF-011] Background jobs use worker pools with bounded concurrency
- [ ] [CL-PERF-012] Rate limiting protects against burst traffic

## Scalability

- [ ] [CL-PERF-013] Database indexes exist for all frequently-queried columns
- [ ] [CL-PERF-014] Stateless request handling (no in-memory session state)
- [ ] [CL-PERF-015] Large payload processing uses streaming (not full in-memory load)

## Observability

- [ ] [CL-PERF-016] Key endpoints have latency metrics (histogram or summary)
- [ ] [CL-PERF-017] Slow query logging enabled with threshold
- [ ] [CL-PERF-018] Error rates tracked per endpoint
- [ ] [CL-PERF-019] Resource utilization alerts configured (CPU, memory, connections)
- [ ] [CL-PERF-020] Request tracing enabled for cross-service calls
