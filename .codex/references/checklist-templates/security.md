# 🔒 Security Checklist Template

> Domain: Security Review Used by: `/checklist-generator` for security audits Extends:
> `**/references/security-checklist.md` (shared reference)

## Authentication & Authorization

- [ ] [CL-SEC-001] Every endpoint verifies the requesting user has permission
- [ ] [CL-SEC-002] No anonymous access to protected resources
- [ ] [CL-SEC-003] Token validation: expiry, signature, issuer checked
- [ ] [CL-SEC-004] Password storage: bcrypt/argon2 with sufficient cost factor

## Data Protection

- [ ] [CL-SEC-005] User A CANNOT access/modify/delete User B's data (ownership check)
- [ ] [CL-SEC-006] Every DB query filters by tenant (`domain_id` / `org_id`)
- [ ] [CL-SEC-007] Sensitive data (passwords, tokens, keys) never in API responses or logs
- [ ] [CL-SEC-008] Encryption at rest for PII and financial data
- [ ] [CL-SEC-009] TLS enforced for all external communication

## Input Validation

- [ ] [CL-SEC-010] All untrusted input validated before use
- [ ] [CL-SEC-011] Handle: empty strings, negative numbers, oversized payloads, path traversal
- [ ] [CL-SEC-012] Parameterized queries only — no SQL string concatenation
- [ ] [CL-SEC-013] File upload validation: type, size, content scanning

## Error Handling

- [ ] [CL-SEC-014] Error responses do NOT leak stack traces, SQL queries, or internal paths
- [ ] [CL-SEC-015] Rate limiting on authentication endpoints
- [ ] [CL-SEC-016] Account lockout after repeated failed attempts

## Supply Chain

- [ ] [CL-SEC-017] Dependencies scanned for known vulnerabilities
- [ ] [CL-SEC-018] No hardcoded secrets in source code
- [ ] [CL-SEC-019] Environment variables used for all configuration secrets
- [ ] [CL-SEC-020] CORS configuration restricts origins to known domains
