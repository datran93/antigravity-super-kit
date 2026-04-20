---
description:
  Continuous execution mode that seamlessly transitions between workflow phases without stopping, relying on MCP state
  as SSOT.
---

# 🚀 Auto-Pilot Workflow

---

## What is Auto-Pilot?

Auto-Pilot is a meta-workflow that chains the standard sequence of workflows (`/planner-architect` →
`/coder-implementation` → `/reviewer-audit` → `/tester-verification`) into a continuous loop.

Instead of stopping at the end of each phase and waiting for the USER to type the next command, the Agent automatically
calls the next workflow IF AND ONLY IF the Quality Gates for the current phase have passed.

## Phase 0: Activation

1. Check current state via `load_checkpoint` or `get_task_summary`.
2. Identify which role should run next based on the task tier (🟢 SMALL, 🟡 MEDIUM, 🔴 LARGE) and current progress.
3. Inform the USER: _"Auto-Pilot activated. Moving to [Next Role]."_

## Phase 1: Continuous Execution

Follow the standard workflow rules for the active role, but with these overrides:

1. **State over Files**: Use the MCP database (`context-manager`) to pass state between roles. DO NOT write or read
   markdown reports (e.g. `coder-report.md`) during the transition.
2. **Batching**: Group related components and actions together to minimize tool calls. `declare_intent` on the grouped
   files.
3. **Internal Review**: Do all self-reviews internally using `<thinking>` tags to save output tokens.
4. **Transition**: At the end of a role's workflow (e.g., Phase 6 of Coder), do NOT stop. Evaluate the Quality Gates.

## Phase 2: Quality Gates & Handoff

Before transitioning to the next role, verify the gates:

- **Coder to Reviewer**: Did code compile/run? Are verification commands passing? → Proceed to `/reviewer-audit`
- **Reviewer to Tester** (🔴 LARGE only): Is the audit APPROVED? → Proceed to `/tester-verification`
- **Tester to Completion**: Is coverage >= 70% and are all bugs fixed? → Mark `status = "completed"`, generate final
  human-readable report.

> **🛑 DRIFT ALARM**: If at any point an action fails 3 times consecutively, or a Quality Gate fails and cannot be
> resolved quickly, the Auto-Pilot MUST disengage. `record_failure` and present the issue to the USER.

## 🔴 Constraints

1. **Never skip gates**: Auto-Pilot must enforce the same quality gates as manual execution.
2. **Respect ANCHORS**: All constraints in `ANCHORS.md` still apply.
3. **Final Report Only**: Only generate a markdown report when the entire pipeline is complete or when the Auto-Pilot
   disengages due to failure.
