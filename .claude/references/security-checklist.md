# 🔒 Security Checklist (Shared Reference)

Used by: Coder (self-review), Reviewer (audit), Tester (bug hunting).

| # | Check | Description |
|---|-------|-------------|
| 1 | **Authorization** | Every endpoint/operation verifies the requesting user has permission. No anonymous access to protected resources. |
| 2 | **Ownership** | User A CANNOT access/modify/delete User B's data. Every mutation verifies resource ownership. |
| 3 | **Tenant isolation** | Every DB query filters by `domain_id` / `org_id`. No cross-tenant data leaks. |
| 4 | **Input validation** | Untrusted input validated before use. Handle: empty strings, negative numbers, oversized payloads, path traversal. |
| 5 | **Sensitive data** | No passwords, tokens, API keys, or internal paths in API responses or logs. |
| 6 | **Error safety** | Error responses do NOT leak stack traces, SQL queries, or internal file paths. |
| 7 | **Injection** | Parameterized queries only. No string concatenation in SQL. |
