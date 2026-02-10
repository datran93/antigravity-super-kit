---
trigger: always_on
description: Instructions for Using Graphiti's MCP Tools for Agent Memory
---

# GRAPHITI.md - Antigravity Kit

## 🧠 The Memory Cycle Protocol

Always follow this 4nd-phase cycle to ensure consistent knowledge management.

### P1: Discovery (Search)

**Systematic search before starting any work.**

- `search_nodes`: Look for relevant `Preference`, `Procedure`, or `Decision`.
- `search_facts`: Find atomic relationships between entities.

### P2: Acquisition (Read)

**Read specific details of discovered knowledge.**

- Use `get_episode` or `get_entity_edge` for full context from UUIDs.

### P3: Integration (Context)

**Apply knowledge to current reasoning.**

- Align implementation with found `Procedures`.
- Respect identified `Preferences`.
- Adhere to functional `Requirements`.

### P4: Retention (Save)

**Capture new intelligence immediately.**

- `add_memory`: Save narrative context (episodes). Split long inputs.
- **Categorize everything**:
  - `Preference`: User style, likes, dislikes.
  - `Procedure`: How-to guides, workflows.
  - `Requirement`: Constraints, specs.
  - `Decision`: Tech choices, ADRs.

---

## ⚡ Memory Triggers

### When to Search (P1)

Activate discovery at these critical moments:

- **Session Start**: Before performing any complex task or analysis.
- **Ambiguity**: When encountering a new concept, acronym, or project-specific term.
- **Decision Prep**: Prior to proposing architectural or technical decisions.

### When to Retain (P4)

Save intelligence as soon as it surfaces:

- **Preference Statements**: "I like...", "I prefer...", "Avoid...".
- **Finalized Decisions**: When a tech choice or ADR is confirmed by the user.
- **Standardized Procedures**: After successfully executing a complex multi-step workflow.

---

## 🛠️ Best Practices

- **Surgical Fact Management**: Use `search_facts` for quick atomic lookup; use narrative episodes for deep context.
- **Hygiene**: Use `delete_episode` if a procedure or preference becomes obsolete.
- **Proactive Recording**: If a user states a rule once, it belongs in memory immediately.
- **UUID Centering**: Use `center_node_uuid` to explore related facts around a specific topic.

**Remember**: Your intelligence is directly proportional to how well you use your memory.
