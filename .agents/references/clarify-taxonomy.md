# 🔎 Clarification Taxonomy

> 10-category structured ambiguity scan for `/clarify-specification`. For each category, the agent marks status:
> **Clear** | **Partial** | **Missing**. Only categories with Partial/Missing status generate candidate questions.

---

## 1. Functional Scope & Behavior

- Core user goals & success criteria
- Explicit out-of-scope declarations
- User roles / personas differentiation
- Feature boundaries (what's included vs excluded)

## 2. Domain & Data Model

- Entities, attributes, relationships
- Identity & uniqueness rules
- Lifecycle / state transitions
- Data volume / scale assumptions
- Validation rules & constraints

## 3. Interaction & UX Flow

- Critical user journeys / sequences
- Error / empty / loading states
- Accessibility or localization notes
- Input/output formats from user perspective

## 4. Non-Functional Quality Attributes

- **Performance**: Latency targets (p50, p99), throughput limits
- **Scalability**: Horizontal/vertical, concurrency limits
- **Reliability & Availability**: Uptime expectations, recovery time objectives
- **Observability**: Logging signals, metrics, tracing requirements

## 5. Security & Privacy

- Authentication / authorization model
- Data protection (encryption at rest, in transit)
- Threat assumptions (what attacks to defend against)
- Compliance / regulatory constraints (GDPR, HIPAA, PCI-DSS)
- Secrets management approach

## 6. Integration & External Dependencies

- External services / APIs and their failure modes
- Data import / export formats
- Protocol / versioning assumptions
- Third-party rate limits or SLAs

## 7. Edge Cases & Failure Handling

- Negative scenarios (invalid input, unauthorized access)
- Rate limiting / throttling behavior
- Conflict resolution (concurrent edits, race conditions)
- Graceful degradation strategy

## 8. Constraints & Tradeoffs

- Technical constraints (language, storage, hosting, ANCHORS.md)
- Explicit tradeoffs or rejected alternatives
- Budget / timeline constraints affecting scope
- Known technical debt being accepted

## 9. Terminology & Consistency

- Canonical glossary terms (single name for each concept)
- Avoided synonyms / deprecated terms
- Domain language alignment

## 10. Completion Signals

- Acceptance criteria testability (can each AC be mechanically verified?)
- Measurable Definition of Done indicators
- Test strategy clarity (unit, integration, E2E)
- Deployment / rollout criteria

---

## Usage Rules

1. **Maximum 5 questions** across the entire clarification session
2. **One question at a time** — never reveal future questions
3. **Prioritize by Impact × Uncertainty** — highest-risk categories first
4. **Skip clear categories** — do not ask about areas already well-specified
5. **Only material questions** — must affect architecture, data modeling, task decomposition, test design, UX behavior,
   operational readiness, or compliance validation
6. **Category coverage balance** — avoid two low-impact questions when a high-impact area is unresolved
7. **Favor rework-reduction** — prioritize clarifications that prevent downstream rework or misaligned acceptance tests
