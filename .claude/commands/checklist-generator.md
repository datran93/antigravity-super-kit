---
description:
  Generate domain-specific checklists from templates and project context.
  Supports requirements, security, ux, performance, and custom domains.
---

# ✅ Checklist Generator

> All Universal Protocols from CLAUDE.md apply (Role Anchoring, Ghost Context, Drift Detection, No Self-Escalation).
>
> **Role**: Utility command — can be invoked at any phase of the pipeline.

---

## Phase 1: Domain Selection & Context Loading 📖

1. **Identify domain** from USER request. Supported domains:
   - `requirements` → `.agents/references/checklist-templates/requirements.md`
   - `security` → `.agents/references/checklist-templates/security.md`
   - `ux` → `.agents/references/checklist-templates/ux.md`
   - `performance` → `.agents/references/checklist-templates/performance.md`
   - `custom` → `.agents/references/checklist-templates/custom-template.md`

   If domain is ambiguous → ask USER to choose from the list above.

2. **Load template** for the selected domain.

3. **Load project context** to customize checklist items:
   - `features/{NNN}-{slug}/spec.md` — for requirements-based items (if exists)
   - `features/{NNN}-{slug}/design.md` — for technical items (if exists)
   - `.agents/rules/ANCHORS.md` — for constraint-based items (always)
   - `load_checkpoint` — for task plan context (if exists)

---

## Phase 2: Generate Checklist 📝

Using the template as a base, generate a **project-specific** checklist:

1. **Include all template items** as baseline
2. **Customize items** based on loaded context:
   - Add project-specific items derived from spec ACs or design decisions
   - Remove items that don't apply (e.g., "tenant isolation" if single-tenant)
   - Add ANCHORS-derived items (e.g., "All code in English" from language anchor)
3. **Add references** to relevant files or sections where applicable:
   - `→ See: features/{NNN}-{slug}/spec.md §3 Key Entities`
   - `→ See: .agents/rules/ANCHORS.md §Quality Standards`

### Item Format

```markdown
- [ ] [CL-{DOMAIN}-NNN] Item description → See: {reference}
```

### Sections

Group items by categories from the template. Each category should have 3-10 items.

---

## Phase 3: Write & Deliver 📦

1. Write checklist to `features/{NNN}-{slug}/checklists/{domain}-checklist.md`
2. Create the `features/{NNN}-{slug}/checklists/` directory if it doesn't exist
3. Present summary to USER:

```
✅ Generated {domain} checklist: features/{NNN}-{slug}/checklists/{domain}-checklist.md
   Items: N (X categories)
   Context sources: spec, design, ANCHORS
```

> 🛑 **STOP HERE.** Checklist delivered. USER decides how to use it.

---

## 🔴 Constraints

1. **Template-based**: Always start from an existing template. Never generate from scratch.
2. **Context-aware**: Customize based on loaded project context when available.
3. **Checklist items are independently verifiable**: Each item is yes/no, never "partially done".
4. **No implementation**: NEVER write code, run tests, or modify source files.
5. **Idempotent**: Re-running overwrites the previous checklist for that domain.
6. **Reference existing security checklist**: Security domain MUST include items from `.agents/references/security-checklist.md`.
