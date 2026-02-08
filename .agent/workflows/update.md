---
description: Add or update features in existing application. Used for iterative development.
---

# /update - Update Existing Feature Workflow

Guide agents to update, enhance, or modify existing features systematically with minimal risk.

---

## When to Use

- `/update [description]` - Update existing feature
- Keywords: "update", "change", "modify", "enhance", "improve", "add to", "edit"

---

## 🔴 Critical Rules

1. **Understand before modifying** - Read existing code first
2. **Impact analysis required** - Know what will be affected
3. **Preserve working functionality** - Don't break what works
4. **Test after changes** - Verify nothing broke

---

## Phase 1: Context Discovery 🔍

### Step 1.1: Understand Current State

Before any changes:

```markdown
### Current State Analysis

**Feature:** [what exists now] **Location:** [files/modules involved] **Dependencies:** [what depends on this] **Last
Modified:** [when/why if known]
```

### Step 1.2: Load Project Context

```bash
# Check project structure
cat CODEBASE.md  # or equivalent

# Identify relevant files
grep -r "feature_name" src/
```

### Step 1.3: Clarifying Questions (If Needed)

If request is ambiguous:

```markdown
### 🔍 Before Updating

1. **Scope**: What exactly should change?
2. **Behavior**: How should it work after update?
3. **Backwards Compatibility**: Should existing behavior be preserved?
4. **Related Features**: Any other areas that should change together?
```

---

## Phase 2: Impact Analysis ⚡

### Step 2.1: Identify Affected Areas

```markdown
### Impact Analysis

| Area             | Impact Level | Description                   |
| ---------------- | ------------ | ----------------------------- |
| **Direct**       | High         | Files being modified directly |
| **Dependencies** | Medium       | Files that import/use these   |
| **Downstream**   | Low          | Features that may be affected |
```

### Step 2.2: File Change Summary

```markdown
### Files to Change

| File                    | Action | Changes             |
| ----------------------- | ------ | ------------------- |
| `src/feature/main.ts`   | MODIFY | Add new logic       |
| `src/feature/types.ts`  | MODIFY | Update interface    |
| `src/utils/helper.ts`   | MODIFY | Add helper function |
| `src/feature/new.ts`    | CREATE | New component       |
| `tests/feature.test.ts` | MODIFY | Add new test cases  |
```

### Step 2.3: Risk Assessment

```markdown
### Risk Level

| Factor           | Assessment   | Notes              |
| ---------------- | ------------ | ------------------ |
| Scope            | Low/Med/High | [how much changes] |
| Breaking Changes | Yes/No       | [what might break] |
| Reversibility    | Easy/Hard    | [how to rollback]  |
| Test Coverage    | Good/Poor    | [existing tests]   |

**Overall Risk:** 🟢 Low | 🟡 Medium | 🔴 High
```

---

## Phase 3: Change Planning 📋

### Step 3.1: Break Down Changes

```markdown
### Change Plan

1. [ ] **Preparation**
   - [ ] Read and understand existing code
   - [ ] Identify all affected files
   - [ ] Create backup/branch if needed

2. [ ] **Core Changes**
   - [ ] [Specific change 1]
   - [ ] [Specific change 2]
   - [ ] [Specific change 3]

3. [ ] **Integration**
   - [ ] Update imports/exports
   - [ ] Update dependencies
   - [ ] Update configurations

4. [ ] **Testing**
   - [ ] Update existing tests
   - [ ] Add new tests
   - [ ] Run all tests

5. [ ] **Cleanup**
   - [ ] Remove deprecated code
   - [ ] Update documentation
```

### Step 3.2: Present Plan (For Major Changes)

For significant updates, confirm with user:

```markdown
### 📋 Update Plan Summary

**Updating:** [feature name] **Scope:** [brief description]

**Changes:**

- Modify: X files
- Create: Y files
- Delete: Z files

**Estimated Time:** ~X minutes **Risk Level:** 🟢 Low | 🟡 Medium | 🔴 High

> Should I proceed with this update?
```

---

## Phase 4: Implementation 🔨

### Step 4.1: Execute Changes

For each change:

1. **Read** existing code first
2. **Modify** with minimal disruption
3. **Verify** change works
4. **Move** to next change

```markdown
### 🔨 Progress

- [x] Step 1: [description] ✅
- [x] Step 2: [description] ✅
- [ ] Step 3: [description] 🔄
- [ ] Step 4: [description] ⏳
```

### Step 4.2: Preserve Backwards Compatibility

When updating interfaces or APIs:

```typescript
// Option 1: Keep old signature, add new
function doThing(oldParam: string): void;
function doThing(oldParam: string, newParam?: Options): void;

// Option 2: Deprecate old, introduce new
/** @deprecated Use doThingV2 instead */
function doThing(oldParam: string): void;
function doThingV2(params: NewParams): void;
```

### Step 4.3: Handle Edge Cases

- [ ] Null/undefined handling
- [ ] Empty state handling
- [ ] Error cases
- [ ] Loading states (if UI)

---

## Phase 5: Verification ✅

### Step 5.1: Smoke Test

```bash
# Run the application
npm run dev  # or equivalent

# Check for errors
npm run lint
npm run typecheck
```

### Step 5.2: Feature Verification

```markdown
### Verification Checklist

**New Behavior:**

- [ ] [Expected behavior 1] works
- [ ] [Expected behavior 2] works

**Existing Behavior:**

- [ ] [Existing feature 1] still works
- [ ] [Existing feature 2] still works

**Edge Cases:**

- [ ] Empty state handled
- [ ] Error state handled
- [ ] Loading state handled (if applicable)
```

### Step 5.3: Run Tests

```bash
# Run affected tests
npm run test -- --related

# Or run all tests
npm run test
```

---

## Phase 6: Documentation & Delivery 📝

### Step 6.1: Update Summary

````markdown
## ✅ Update Complete: [Feature Name]

### What Changed

[Brief description of changes]

### Files Modified

| File           | Changes        |
| -------------- | -------------- |
| `path/to/file` | [what changed] |

### New Behavior

- [New behavior 1]
- [New behavior 2]

### Breaking Changes

- [None | List breaking changes]

### Testing

- [x] Unit tests updated
- [x] All tests passing
- [x] Manual testing done

### Rollback Instructions

If needed, revert with:

```bash
git revert [commit-hash]
```
````

```

### Step 6.2: Save & Notify

1. Save to `agent-docs/UPDATE-{slug}.md`
2. **Slug generation**: Extract 2-3 key words → lowercase → hyphen-separated → max 30 chars
   - "add dark mode" → `UPDATE-dark-mode.md`
   - "improve search performance" → `UPDATE-search-perf.md`
   - "fix user authentication" → `UPDATE-user-auth.md`
3. Notify: `✅ Update complete! Report saved: agent-docs/UPDATE-{slug}.md`

---

## Quick Reference

### Workflow Flow

```

Context Discovery → Impact Analysis → Change Planning → Implementation → Verification → Delivery ↓ ↓ ↓ ↓ ↓ ↓ Read
existing Identify Break down Execute Test all Document code first affected files changes carefully scenarios + Save

````

### Change Actions

| Action   | Icon | Description                         |
| -------- | ---- | ----------------------------------- |
| MODIFY   | ✏️    | Change existing file                |
| CREATE   | ➕    | Add new file                        |
| DELETE   | 🗑️    | Remove file                         |
| RENAME   | 📝    | Rename/move file                    |
| REFACTOR | ♻️    | Restructure without behavior change |

### Risk Indicators

| Indicator                   | Risk Level |
| --------------------------- | ---------- |
| Single file change          | 🟢 Low      |
| Multiple files, same module | 🟡 Medium   |
| Cross-module changes        | 🔴 High     |
| Database schema change      | 🔴 High     |
| API contract change         | 🔴 High     |
| Core business logic         | 🔴 High     |

---

## Anti-Patterns (AVOID)

| ❌ Anti-Pattern               | ✅ Instead                        |
| ---------------------------- | -------------------------------- |
| Modify without reading first | Understand existing code first   |
| Change everything at once    | Small, incremental changes       |
| Skip impact analysis         | Always identify affected areas   |
| No testing after changes     | Verify nothing broke             |
| Silent breaking changes      | Document and communicate changes |
| Hardcode fixes               | Proper, maintainable solutions   |
| Skip documentation           | Document what changed and why    |

---

## Examples

```bash
/update add dark mode toggle to settings
/update improve search with fuzzy matching
/update change user avatar upload to support multiple files
/update refactor authentication to use JWT
/update add pagination to product listing
/update fix cart total calculation
````
