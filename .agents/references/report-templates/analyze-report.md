# Analyze Report Template

```
## 📊 Cross-Artifact Analysis Report

**Task ID**: {task-id}
**Date**: {date}
**Artifacts Analyzed**: features/YYYY-MM-DD-{slug}/spec.md, design.md, MCP progress.md

---

### 🗂️ Inventory Summary

| Artifact | Items Found |
|----------|-------------|
| Requirements (from spec) | N user stories, M acceptance criteria |
| Design Elements (from design) | N modules, M data models, K contracts |
| Task Actions (from plan) | N tasks across M phases |

---

### 🔍 Coverage Mapping

| Requirement (AC-ID) | Design Element | Task(s) | Status |
|---------------------|----------------|---------|--------|
| AC-1: ... | Module X | [T001] | ✅ Covered |
| AC-2: ... | — | — | ❌ Gap |
| — | Module Y | [T003] | ⚠️ Orphan (no requirement) |

---

### 🚨 Findings

| # | Severity | Category | Description | Artifact | Recommendation |
|---|----------|----------|-------------|----------|----------------|
| 1 | CRITICAL | Coverage Gap | AC-2 has no design element or task | spec | Add design + task for AC-2 |

**Severity Definitions**:
- **CRITICAL**: Violates ANCHORS, missing core artifact, zero-coverage blocking requirement
- **HIGH**: Conflicting requirement, ambiguous security/performance attribute
- **MEDIUM**: Terminology drift, missing NFR task coverage
- **LOW**: Style/wording improvement

**Categories**: Duplication | Ambiguity | Underspecification | ANCHORS Alignment | Coverage Gap | Inconsistency

---

### 📈 Metrics

| Metric | Value |
|--------|-------|
| Total Requirements | N |
| Covered Requirements | M (X%) |
| Uncovered Requirements | K |
| Orphan Tasks (no requirement) | J |
| Orphan Design Elements | I |
| CRITICAL findings | N |
| HIGH findings | N |

---

### 💡 Recommended Actions

1. Address CRITICAL findings before proceeding to implementation
2. Review HIGH findings with USER for prioritization
3. MEDIUM/LOW findings may be deferred with justification

---

### Verdict
- [ ] ✅ CONSISTENT — No blocking issues, safe to proceed to /coder-implementation
- [ ] ⚠️ GAPS FOUND — Address CRITICAL/HIGH findings before implementation
- [ ] ❌ MISALIGNED — Significant inconsistencies require spec/design revision
```
