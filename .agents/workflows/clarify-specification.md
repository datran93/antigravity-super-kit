---
description:
  Structured workflow for post-specification ambiguity resolution. Performs a 10-category structured scan, asks ≤5
  targeted clarification questions with recommendations, and integrates answers back into the spec.
---

# 🔎 Clarify Specification

> All Universal Protocols from GEMINI.md apply (Role Anchoring, Ghost Context, Drift Detection, No Self-Escalation).
>
> **Role**: This is a **Spec Writer sub-role**. It refines specs — it NEVER generates architecture, task plans, or code.

---

## Phase 0: Load Spec 📖

1. Identify the active spec: `features/{NNN}-{slug}/spec.md`.
2. `load_checkpoint` — load task context if exists.
3. Read the full spec file. If no spec exists, abort and instruct the USER to run `/specifications-writer` first.

> ❌ NEVER run without an existing spec file.

---

## Phase 1: Structured Ambiguity Scan 🔬

Perform a systematic scan using the 10-category taxonomy from `**/references/clarify-taxonomy.md`.

For each category, mark status:

| Status      | Meaning                                               |
| ----------- | ----------------------------------------------------- |
| **Clear**   | Requirements are specific, measurable, and actionable |
| **Partial** | Some information exists but gaps remain               |
| **Missing** | No meaningful specification for this category         |

Build an internal coverage map (do NOT output the raw map unless no questions will be asked).

**Skip categories where**:

- Clarification would not materially change implementation or validation strategy
- Information is better deferred to planning phase (note internally)

---

## Phase 2: Generate Question Queue 🎯

From the coverage map, generate a **prioritized queue** of candidate questions.

**Hard Constraints**:

- **Maximum 5 questions total** across the entire session
- Each question must be answerable with EITHER:
  - A **multiple-choice selection** (2–5 distinct, mutually exclusive options), OR
  - A **short answer** (≤ 5 words)
- Only include questions whose answers materially impact: architecture, data modeling, task decomposition, test design,
  UX behavior, operational readiness, or compliance validation
- **Category coverage balance**: Cover highest-impact unresolved categories first; avoid two low-impact questions when a
  high-impact area is unresolved
- Exclude questions already answered in the spec or clarifications section
- Favor clarifications that **reduce downstream rework risk**
- If more than 5 categories remain unresolved, select top 5 by **Impact × Uncertainty** heuristic

If **no valid questions** exist → immediately report "No critical ambiguities found" with coverage statistics → STOP.

---

## Phase 3: Sequential Questioning Loop 🗣️

Present **EXACTLY ONE question** at a time. Never reveal future queued questions.

### For Multiple-Choice Questions:

1. **Analyze all options** and determine the **most suitable option** based on:
   - Best practices for the project type
   - Common patterns in similar implementations
   - Risk reduction (security, performance, maintainability)
   - Alignment with explicit project goals or constraints

2. Present your **recommended option prominently**:

   ```
   **Recommended:** Option [X] — <reasoning in 1-2 sentences>
   ```

3. Render all options as a Markdown table:

   | Option | Description                                 |
   | ------ | ------------------------------------------- |
   | A      | Option A description                        |
   | B      | Option B description                        |
   | C      | Option C description                        |
   | Short  | Provide a different short answer (≤5 words) |

4. Add: _"Reply with the option letter (e.g., 'A'), accept the recommendation by saying 'yes', or provide your own short
   answer."_

### For Short-Answer Questions:

1. Provide your **suggested answer** based on best practices:

   ```
   **Suggested:** <your proposed answer> — <brief reasoning>
   ```

2. Add: _"Format: Short answer (≤5 words). Accept the suggestion by saying 'yes', or provide your own answer."_

### After Each Answer:

1. If user replies "yes" / "recommended" / "suggested" → use your previously stated answer
2. Otherwise, validate the answer maps to one option or fits ≤5 word constraint
3. If ambiguous → ask for quick disambiguation (counts as same question, do NOT advance)
4. Record answer in working memory (do NOT write to disk yet)
5. Move to the next queued question

### Stop Conditions:

- All critical ambiguities resolved early (remaining items become unnecessary)
- User signals completion ("done", "good", "no more")
- 5 questions reached

---

## Phase 4: Integrate Answers into Spec 📝

After the questioning loop completes:

1. **Ensure structure exists** in `features/{NNN}-{slug}/spec.md`:
   - Add `## Clarifications` section if not present (place after the highest-level overview section)
   - Create `### Session YYYY-MM-DD` subheading for today's date

2. **Record each Q&A**:

   ```markdown
   - Q: <question> → A: <final answer>
   ```

3. **Apply to relevant sections**: For each answer, update the most appropriate spec section:
   - New AC entries → add to relevant User Story
   - Entity clarifications → update Key Entities table
   - NFR decisions → add to Constraining Assumptions or relevant category
   - Security decisions → add to relevant section

4. **Remove resolved `[NEEDS CLARIFICATION]` markers** if the answer addresses them

5. `save_checkpoint` with updated notes

---

## Phase 5: Hand-off 🔄

Present summary of changes made to the spec:

```
✅ Clarification complete. X questions asked, Y spec sections updated.
→ Recommended next step: /planner-architect
```

> 🛑 **STOP HERE.** NEVER generate architecture, task plans, or code.

---

## 🔴 Constraints

1. **Spec Writer sub-role ONLY**: No architecture, no code, no task plans.
2. **Maximum 5 questions**: Hard limit. No exceptions.
3. **One question at a time**: NEVER batch questions or reveal future ones.
4. **Always recommend**: Every question MUST have a recommended/suggested answer.
5. **Read-only until Phase 4**: Do NOT modify the spec during questioning.
6. **Respect existing answers**: NEVER re-ask questions already answered in the spec.
7. **Material impact only**: Skip trivial or plan-level questions.
