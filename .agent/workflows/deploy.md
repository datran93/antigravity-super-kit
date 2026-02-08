---
description: Deployment command for production releases. Pre-flight checks and deployment execution for production.
---

# /deploy - Production Deployment Workflow

Guide agents to execute zero-downtime, professional production deployments with safety checks, monitoring, and quick
recovery.

---

## When to Use

- `/deploy production` - **Standard production release**
- `/deploy rollback` - Emergency rollback
- `/deploy check` - Dry-run or pre-deployment checks
- `/deploy maintenance` - Toggle maintenance mode

---

## 🔴 Critical Safety Rules

1. **Never deploy on Fridays** (unless critical hotfix).
2. **Backups required** before database migrations.
3. **Notify stakeholders** before major changes.
4. **Identify rollback strategy** BEFORE deployment.
5. **Freeze code** during deployment window.

---

## Phase 1: Pre-Flight (Go/No-Go Decision) 🚦

### Step 1.1: Environment Verification

```markdown
### Environment Check

| Check         | Status                | Notes                      |
| ------------- | --------------------- | -------------------------- |
| **Branch**    | [main/master/release] | Must be stable branch      |
| **CI Status** | ✅ Passed             | All tests green            |
| **Staging**   | ✅ Verified           | Feature tested on staging  |
| **Database**  | ✅ Backed up          | [timestamp of last backup] |
| **Team**      | ✅ Online             | Deployment team ready      |
```

### Step 1.2: Impact Assessement

```markdown
### Impact Analysis

**Migration Required:** Yes / No **Downtime Expected:** Yes / No **Breaking Changes:** Yes / No **Risk Level:** 🟢 Low |
🟡 Medium | 🔴 High
```

### Step 1.3: Communication Plan

Before proceeding, announce:

```markdown
📢 **Deploying to Production** **Version:** [version] **Downtime:** [None / expected duration] **Why:** [Release notes
summary]
```

---

## Phase 2: Deployment Execution 🚀

### Step 2.1: Database Preparation (If needed)

If migrations are required:

1. **Backup Database:** ensure point-in-time recovery is available.
2. **Review Migrations:** check for locking operations on large tables.
3. **Run Migrations:** (often done during deploy step, but critical to monitor).

### Step 2.2: Maintenance Mode (Optional)

If downtime is required or safe deployment strategy dictates:

```bash
# Enable maintenance page
[command to enable maintenance mode]
```

### Step 2.3: Execute Deployment Strategy

Select strategy based on platform capabilities:

| Strategy           | Description                         | Best For                        |
| :----------------- | :---------------------------------- | :------------------------------ |
| **Rolling Update** | Replace instances one by one        | Zero downtime, most apps        |
| **Blue/Green**     | Deploy parallel env, switch traffic | Critical apps, instant rollback |
| **Canary**         | Release to % of users first         | High risk features              |
| **Recreate**       | Stop all -> Start all               | Dev/Staging, downtime okay      |

**Command:** `[deployment command]`

---

## Phase 3: Verification & Health Check ✅

### Step 3.1: System Health

Verify all components are operational:

- [ ] **Web App:** Responds with 200 OK
- [ ] **API:** Endpoints reachable
- [ ] **Database:** Connections successful
- [ ] **Cache:** ([Redis/Memcached]) Connected
- [ ] **Background Jobs:** Processing queues

### Step 3.2: Critical Path Tests

Manually or automatically verify critical user flows:

1. **Login/Auth:** Can users sign in?
2. **Core Feature:** Can users perform main action?
3. **Payment/Checkout:** Can users pay? (if applicable)

### Step 3.3: Monitoring

Watch metrics for 10-15 minutes:

- **Error Rate:** Should be < 1% (or baseline)
- **Latency:** Should be within normal range
- **CPU/Memory:** Stable usage

---

## Phase 4: Post-Deployment 🏁

### Step 4.1: Cleanup

- [ ] Disable maintenance mode (if enabled)
- [ ] Remove old artifacts/images (if applicable)
- [ ] Close deployment ticket

### Step 4.2: Announce Success

```markdown
✅ **Deployment Complete** **Version:** [version] **Status:** Stable **Key Features:** [list]
```

---

## Phase 5: Emergency Rollback 🚨

**TRIGGER:** If Error Rate > Threshold OR Critical Function Broken.

### Step 5.1: Execute Rollback

**Strategy:** Revert to previous stable version/commit.

```bash
# Rollback command
[platform rollback command]
```

### Step 5.2: Database Reversion (If needed)

**CAUTION:** Only revert migrations if safe and data loss is acceptable or managed.

### Step 5.3: Post-Mortem

After stabilization, investigate root cause.

---

## Quick Reference

### Deployment Checklist

- [ ] Code frozen & CI passed
- [ ] Staging verified
- [ ] Database backed up
- [ ] Team notified
- [ ] **DEPLOY**
- [ ] Health checks passed
- [ ] Monitoring stable
- [ ] Success announced

### Recovery Commands

- **Rollback:** `[rollback command]`
- **Restart:** `[restart command]`
- **Logs:** `[logs command]`

---

## Anti-Patterns (AVOID)

| ❌ Anti-Pattern              | ✅ Instead                               |
| :--------------------------- | :--------------------------------------- |
| Deploying uncommitted code   | Always deploy from git tag/commit        |
| Manual server updates        | Use CI/CD or automated scripts           |
| Ignoring database migrations | Plan & test migrations on staging        |
| "It works on my machine"     | Rely on Staging environment verification |
| Deploying & leaving          | Monitor metrics for at least 15 mins     |
