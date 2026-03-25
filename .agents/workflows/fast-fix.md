---
description:
  Fast-track workflow for 🟢 SMALL tasks (bug fixes, typos, config tweaks, doc updates, dependency bumps).
  Single-pass — reads, fixes, self-reviews, reports. No spec, no design, no feature directory. Does NOT commit.
---

# ⚡ Fast-Fix Workflow (🟢 SMALL)

> Universal Protocols from GEMINI.md apply: Role Anchoring, Drift Detection, No Self-Escalation.
> Ghost Context is OPTIONAL for small tasks.

---

## When to Use

This workflow is for tasks where **ALL** of the following are true:

- < 50 LOC estimated change
- No new files (only modify existing)
- No DB schema or API contract changes
- No auth / payment / security-critical logic

**Examples**: bug fix, typo, config tweak, dependency bump, style change, doc update, workflow tweak.

> If any condition is false → use `/build` (🟡 MEDIUM) or `/planner-architect` (🔴 LARGE) instead.

---

## Phase 0: Understand 📖

1. Read the USER's request carefully.
2. Identify target file(s) — read them before modifying.
3. Quick pattern scan — note error handling, naming, format conventions.

> ❌ NEVER modify a file without reading it first.

---

## Phase 1: Fix 🔧

- Make the change. Keep it minimal and focused.
- Follow existing patterns in the file.
- Clean Code: clear naming, no hardcoded values.

---

## Phase 2: Self-Review 🔍

Re-read ALL changes. Check:

- **Logic**: No copy-paste bugs, all error paths handled, nil/null checks.
- **Pattern**: Matches surrounding code style.
- **Scope**: Only changed what was requested — nothing else.

> Fix issues BEFORE proceeding.

---

## Phase 3: Verify ✅

- If a relevant test suite exists → run it to confirm no regressions.
- On **fail**: Fix and re-run. After 3 consecutive failures → `record_failure` → STOP → ask USER.
- On **pass**: Proceed to report.

---

## Phase 4: Report 📋

Present a brief summary:

```
⚡ Fast-Fix Complete

**Changed**: file(s) modified
**What**: brief description of the fix
**Verified**: test result or manual verification
```

> 🛑 **STOP HERE.** The USER decides the next step (commit, review, or continue).

---

## 🔴 Constraints

1. **Read before write**: NEVER modify a file without reading it first.
2. **Minimal scope**: Change ONLY what was requested. No opportunistic refactoring.
3. **No artifacts**: Do NOT create spec.md, design.md, tasks.md, or feature directories.
4. **No intent locking**: Skip `declare_intent` / `check_intent_lock` for SMALL tasks.
5. **No commit**: NEVER commit — the USER or Planner decides when to commit.
6. **Escape hatch**: If the fix turns out to be bigger than expected → STOP → recommend `/build` or `/planner-architect`.
