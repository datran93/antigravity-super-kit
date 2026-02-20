---
description: Deployment command for production releases. Pre-flight checks and deployment execution for production.
---

# /deploy - Production Deployment Workflow

Guide agents to execute zero-downtime, professional production deployments with safety checks, monitoring, and quick
recovery.

## When to Use

- `/deploy production` - Standard production release
- `/deploy rollback` - Emergency rollback
- `/deploy check` - Dry-run or pre-deployment checks

## 🔴 Critical Safety Rules

1. **Never deploy on Fridays** (unless critical hotfix).
2. **Backups required** before database migrations.
3. **Notify stakeholders** before major changes.
4. **Identify rollback strategy** BEFORE deployment.
5. **Freeze code** during deployment window.

---

## Phase 1: Classification & Skill Mapping 🔀

Look up deployment scripts and infrastructure patterns in `.agent/CATALOG.md` depending on the target stack (e.g., AWS,
Vercel, Docker).

---

## Phase 2: Pre-Flight (Go/No-Go Decision) 🚦 (Socratic Gate)

Perform environment checks:

- Must deploy from stable branch.
- CI/Tests must be green.
- Staging verified.
- DB backed up (if migrating).

Analyze Impact (Downtime expected? Breaking changes? Risk Level?). Ask user for final confirmation before executing.

---

## Phase 3: Deployment Execution 🚀

1. **Database:** Run non-locking migrations safely if needed.
2. **Maintenance Mode:** Enable if required.
3. **Execute Strategy:** Follow best practices (Rolling, Blue/Green, Canary, Recreate).

---

## Phase 4: Verification & Health Check ✅

Verify system health post-deploy:

- Check endpoints (HTTP 200).
- Check database connectivity.
- Verify critical paths (Login, Data Writing). Monitor logs and error rates for anomalies.

---

## Phase 5: Post-Deployment or Emergency Rollback 🚨

**Post-Deployment:** Cleanup cache, announce success. **Emergency Rollback:** If error rates spike, instantly revert to
previous commit/image.

### Quick Checklist

- [ ] CI passed
- [ ] DB Backed up
- [ ] Rolled out
- [ ] Health checks passed
