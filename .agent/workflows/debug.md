---
description: Debugging command. Activates DEBUG mode for systematic problem investigation.
---

# /debug - Systematic Problem Investigation

Guide agents to debug issues methodically: understand → hypothesize → investigate → fix → prevent.

---

## When to Use

- `/debug [issue description]` - Start debug workflow
- Keywords: "bug", "error", "not working", "broken", "fails", "crash", "issue"

---

## 🔴 Critical Rules

1. **Reproduce first** - Confirm the issue before fixing
2. **One change at a time** - Don't change multiple things
3. **Verify fix works** - Test after each change
4. **Understand root cause** - Don't just patch symptoms

---

## Phase 1: Information Gathering 📋

### Step 1.1: Capture the Problem

Gather all available information:

```markdown
### Problem Report

**Symptom:** [What is happening] **Expected:** [What should happen] **Actual:** [What actually happens] **Frequency:**
[Always / Sometimes / Once] **Environment:** [Dev / Staging / Prod]
```

### Step 1.2: Collect Evidence

```markdown
### Evidence Collected

**Error Message:**
```

[exact error message]

```

**Stack Trace:**
```

[stack trace if available]

```

**Relevant Logs:**
```

[log entries around the time of issue]

```

**Recent Changes:**
- [Change 1 - date/commit]
- [Change 2 - date/commit]
```

### Step 1.3: Reproduction Steps

```markdown
### Reproduction Steps

1. [Step 1]
2. [Step 2]
3. [Step 3]
4. **Bug occurs here**

**Reproducible:** ✅ Yes | ❌ No | ⚠️ Intermittent
```

### Step 1.4: Clarifying Questions (If Needed)

If information is missing:

```markdown
### 🔍 Need More Information

1. **When** did this start happening?
2. **What** changed recently?
3. **Can you reproduce** the issue consistently?
4. **What's the exact** error message?
5. **Which environment** is this in?
```

---

## Phase 2: Hypothesis Formation 🧠

### Step 2.1: Generate Hypotheses

List possible causes, ordered by likelihood:

```markdown
### Hypotheses

| #   | Hypothesis           | Likelihood | Why I Think This           |
| --- | -------------------- | ---------- | -------------------------- |
| 1   | [Most likely cause]  | High       | [Evidence supporting this] |
| 2   | [Second possibility] | Medium     | [Evidence supporting this] |
| 3   | [Less likely cause]  | Low        | [Why it's still possible]  |
```

### Step 2.2: Hypothesis Prioritization

Test in this order:

1. **Most likely** based on evidence
2. **Easiest to verify** (quick wins first)
3. **Most impactful** if confirmed

### Step 2.3: Common Bug Categories

Use this to guide hypothesis formation:

| Category        | Common Causes                                   |
| --------------- | ----------------------------------------------- |
| **Data**        | Null/undefined, wrong type, missing field       |
| **State**       | Race condition, stale state, initialization     |
| **Integration** | API contract changed, timeout, auth failed      |
| **Logic**       | Off-by-one, wrong operator, missing condition   |
| **Environment** | Wrong config, missing env var, version mismatch |
| **Concurrency** | Race condition, deadlock, missing lock          |

---

## Phase 3: Systematic Investigation 🔍

### Step 3.1: Test Each Hypothesis

For each hypothesis:

```markdown
### Testing Hypothesis 1: [Description]

**Test:** [What I'm checking] **Method:** [How I'm checking] **Result:** ✅ Confirmed | ❌ Ruled Out | ⚠️ Inconclusive
**Evidence:** [What I found]
```

### Step 3.2: Investigation Techniques

| Technique         | When to Use                              |
| ----------------- | ---------------------------------------- |
| **Print/Log**     | Trace execution flow and variable values |
| **Breakpoint**    | Pause at specific point to inspect state |
| **Binary Search** | Narrow down: works in A, fails in B      |
| **Minimal Repro** | Strip away until only bug remains        |
| **Diff Analysis** | Compare working vs broken version        |
| **Isolation**     | Test component independently             |

### Step 3.3: Track Investigation

```markdown
### Investigation Log

| Time  | Action                  | Result                     |
| ----- | ----------------------- | -------------------------- |
| 00:00 | Checked error logs      | Found [finding]            |
| 00:05 | Added logging to [file] | Discovered [finding]       |
| 00:10 | Tested hypothesis 1     | Ruled out                  |
| 00:15 | Tested hypothesis 2     | **CONFIRMED** - root cause |
```

---

## Phase 4: Root Cause Analysis 🎯

### Step 4.1: Identify Root Cause

```markdown
### 🎯 Root Cause

**What:** [Technical description of the bug]

**Why:** [Why this caused the symptom]

**Where:**

- File: `path/to/file.ts`
- Line: [line number]
- Function: `functionName()`

**When Introduced:** [If known - commit/date/change]
```

### Step 4.2: Understand the Chain

```markdown
### Cause → Effect Chain

1. [Root cause] ↓
2. [Intermediate effect] ↓
3. [Visible symptom]
```

---

## Phase 5: Fix Implementation 🔧

### Step 5.1: Design the Fix

Before coding, plan:

```markdown
### Fix Plan

**Approach:** [How to fix] **Files to Modify:**

- `path/to/file1.ts`
- `path/to/file2.ts`

**Risk Level:** 🟢 Low | 🟡 Medium | 🔴 High **Breaking Changes:** Yes / No
```

### Step 5.2: Implement Fix

````markdown
### Fix Applied

**File:** `path/to/file.ts`

**Before:**

```typescript
// Broken code
const result = data.value; // data can be null
```
````

**After:**

```typescript
// Fixed code
const result = data?.value ?? defaultValue;
```

**Explanation:** [Why this fixes the issue]

````

### Step 5.3: Verify Fix

```markdown
### Verification

- [ ] Bug no longer reproduces
- [ ] Original functionality still works
- [ ] No new errors introduced
- [ ] Tests pass
````

---

## Phase 6: Prevention & Documentation 🛡️

### Step 6.1: Add Prevention Measures

````markdown
### Prevention Measures

**Test Added:**

```typescript
test("should handle null data gracefully", () => {
  // Test code
});
```
````

**Validation Added:**

```typescript
// Input validation
if (!data) throw new ValidationError("Data required");
```

**Documentation Updated:**

- [ ] Code comments added
- [ ] README updated (if needed)
- [ ] Known issues documented

````

### Step 6.2: Lessons Learned

```markdown
### Lessons Learned

**What went wrong:** [Root cause summary]
**How to prevent:** [Specific practices]
**What to watch for:** [Warning signs]
````

### Step 6.3: Save & Notify

1. Save to `agent-docs/DEBUG-{slug}.md`
2. **Slug generation**: Extract 2-3 key words → lowercase → hyphen-separated → max 30 chars
   - "login not working" → `DEBUG-login-issue.md`
   - "API returns 500" → `DEBUG-api-500.md`
   - "cart total wrong" → `DEBUG-cart-total.md`
3. Notify: `✅ Bug fixed! Report saved: agent-docs/DEBUG-{slug}.md`

---

## Output Template

````markdown
## 🔍 Debug Report: [Issue Title]

### 1. Problem

**Symptom:** [What's happening] **Expected:** [What should happen] **Reproducible:** Yes / No / Intermittent

### 2. Evidence

**Error:** `[error message]` **Location:** `file:line`

### 3. Hypotheses Tested

1. ❌ [Hypothesis 1] - Ruled out because [reason]
2. ✅ [Hypothesis 2] - **CONFIRMED**
3. ⏭️ [Hypothesis 3] - Not tested (root cause found)

### 4. Root Cause

🎯 [Clear explanation of why this happened]

### 5. Fix

**File:** `path/to/file`

```diff
- broken code
+ fixed code
```
````

### 6. Prevention

🛡️ [Test/validation added to prevent recurrence]

### 7. Verification

- [x] Bug fixed
- [x] Tests pass
- [x] No regressions

```

---

## Quick Reference

### Workflow Flow

```

Information → Hypothesis → Investigation → Root Cause → Fix → Prevention ↓ ↓ ↓ ↓ ↓ ↓ Gather Generate Test each Identify
Implement Add tests evidence ordered one by one cause + verify + document

```

### Debug Decision Tree

```

Is error message clear? ├─ Yes → Search codebase for related code └─ No → Add logging to trace execution

Can you reproduce? ├─ Yes → Binary search to narrow down └─ No → Add monitoring, wait for recurrence

Is it data-related? ├─ Yes → Check input validation, null checks └─ No → Check logic, state management

Is it environment-specific? ├─ Yes → Compare configs, check versions └─ No → Check code logic

````

---

## Anti-Patterns (AVOID)

| ❌ Anti-Pattern                 | ✅ Instead                            |
| ------------------------------ | ------------------------------------ |
| Guess and check randomly       | Form hypotheses, test systematically |
| Change multiple things at once | One change at a time                 |
| Fix symptom, not cause         | Find and fix root cause              |
| Skip reproduction              | Always confirm bug exists first      |
| No verification after fix      | Test that fix actually works         |
| No prevention                  | Add test to prevent regression       |
| Just fix and move on           | Document for future reference        |

---

## Examples

```bash
/debug login button not responding
/debug API returns 500 on user update
/debug form validation not working
/debug data not persisting after refresh
/debug images not loading in production
/debug checkout total calculates wrong
````
