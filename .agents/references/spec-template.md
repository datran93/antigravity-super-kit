# 📝 Specification Template

> This is the canonical template for `features/YYYY-MM-DD-{slug}/spec.md`. The Spec Writer MUST follow this structure.

---

## 1. Overview

> _Brief, one-paragraph summary of the feature. What problem does it solve? For whom?_

---

## 2. User Stories (Prioritized)

> List user stories in descending priority (P1 = must-have for MVP, P2 = important, P3 = nice-to-have). Each story MUST
> include: priority justification, independent testability, and GIVEN/WHEN/THEN acceptance criteria.

### [P1] Story Title

**Priority Justification**: _Why this is highest priority — what breaks or is impossible without it._

**Independent Test**: _Can be fully tested by [specific action, e.g. "POST /api/register with valid payload → 201"]._

**Acceptance Criteria**:

```
AC-1: GIVEN <precondition>, WHEN <action>, THEN <measurable outcome>
AC-2: GIVEN <precondition>, WHEN <action>, THEN <measurable outcome>
```

### [P2] Story Title

**Priority Justification**: _Why this is secondary — depends on P1, or enhances but doesn't enable core flow._

**Independent Test**: _Can be fully tested by [specific action]._

**Acceptance Criteria**:

```
AC-3: GIVEN <precondition>, WHEN <action>, THEN <measurable outcome>
```

### [P3] Story Title _(optional)_

**Priority Justification**: _Nice to have — improves experience but not blocking._

**Independent Test**: _Can be fully tested by [specific action]._

**Acceptance Criteria**:

```
AC-4: GIVEN <precondition>, WHEN <action>, THEN <measurable outcome>
```

---

## 3. Key Entities

> Data entities that are first-class citizens in this feature. Define them at the spec level.

| Entity | Core Attributes       | Relationships              | Uniqueness/Identity | State Transitions             |
| ------ | --------------------- | -------------------------- | ------------------- | ----------------------------- |
| _Name_ | _attr1, attr2, attr3_ | _belongs to X, has many Y_ | _Unique by [field]_ | _created → active → archived_ |

---

## 4. Core Ontology

> Domain boundaries, property definitions, state machines. What IS this system, really?

---

## 5. Explicit Non-Goals

> What we are deliberately NOT building. Be specific.

- _NOT implementing X because..._
- _Out of scope: Y_

---

## 6. Constraining Assumptions

> Explicit boundaries, trade-offs, and assumptions that constrain the design space.

- _Assumption: Maximum N concurrent users_
- _Trade-off: Chose X over Y because..._

---

## 7. Clarification Markers

> Maximum 3 items that CANNOT proceed without USER input. Only for scope/security/UX ambiguities where no reasonable
> default exists.

- `[NEEDS CLARIFICATION]` _Description of what is ambiguous and why it matters_

---

## 8. Clarifications

> _Populated by `/clarify-specification` sessions. Do not manually edit._

### Session YYYY-MM-DD

- Q: _question_ → A: _answer_
