---
description: Structured brainstorming for projects and features. Explores multiple options before implementation.
---

# /brainstorm - Structured Idea Exploration

Guide agents to brainstorm user ideas systematically: understand → explore → evaluate → decide.

---

## When to Use

- `/brainstorm [topic]` - Start structured brainstorm
- Keywords: "explore", "options", "ideas", "approach", "how should I", "what if"

---

## Phase 1: Understanding Lock 🔒

**🛑 MANDATORY: Do not generate options before fully understanding the problem.**

### Step 1.1: Extract Context

From user request, identify:

| Dimension        | Question                                  |
| ---------------- | ----------------------------------------- |
| **Problem**      | What specific problem needs to be solved? |
| **Goal**         | What is the desired outcome?              |
| **Constraints**  | Any limitations (tech, time, resources)?  |
| **Context**      | Greenfield / Existing system / Migration? |
| **Stakeholders** | Who will be affected by this decision?    |

### Step 1.2: Clarifying Questions (If Needed)

If insufficient information, ASK before brainstorming:

```markdown
### 🔍 Clarifying Questions

1. **Problem**: [Specific question about the problem]
2. **Scope**: [Question about boundaries]
3. **Constraints**: [Question about limitations]
4. **Priority**: [Question about what matters most]
```

### Step 1.3: Confirm Understanding

After gathering sufficient info, summarize to confirm:

```markdown
### ✅ Understanding Lock

**Problem:** [1-2 sentences] **Goal:** [desired outcome] **Constraints:** [key limitations] **Success Criteria:** [how
we know it works]

> Is this correct? If yes, I will proceed with brainstorming.
```

---

## Phase 2: Divergent Thinking 🌐

**Goal: Generate as many options as possible without judgment.**

### Step 2.1: Option Generation Techniques

Use these techniques to generate ideas:

| Technique            | How to Apply                                |
| -------------------- | ------------------------------------------- |
| **First Principles** | Break down to core, rebuild from scratch    |
| **Analogies**        | How is this solved in other domains?        |
| **Inversion**        | What if we do the opposite? What to avoid?  |
| **Extremes**         | Unlimited resources? Only 1 day to deliver? |
| **Combinations**     | Combine 2 approaches into a hybrid?         |

### Step 2.2: Generate Minimum 3 Options

Each option must have:

- **Clear name** - memorable, describes the approach
- **Short description** - 2-3 sentences explaining core idea
- **How it works** - high-level implementation steps

```markdown
### Option A: [Descriptive Name]

[2-3 sentence description of the approach]

**How it works:**

1. [Step 1]
2. [Step 2]
3. [Step 3]
```

### Step 2.3: Include Unconventional Option

**Required: At least 1 "unconventional" option:**

- An option that seems crazy at first but might be viable
- A counter-intuitive approach
- An approach borrowed from another domain

---

## Phase 3: Convergent Analysis 📊

**Goal: Evaluate each option objectively.**

### Step 3.1: Evaluation Criteria

Determine criteria important for this decision:

| Criteria Category | Examples                                          |
| ----------------- | ------------------------------------------------- |
| **Effort**        | Development time, complexity, learning curve      |
| **Impact**        | User value, business value, problem solved        |
| **Risk**          | Technical risk, adoption risk, maintenance burden |
| **Scalability**   | Can this grow with needs?                         |
| **Reversibility** | How hard to change later?                         |

### Step 3.2: Pros/Cons Analysis

For each option:

```markdown
### Option A: [Name]

✅ **Pros:**

- [Benefit with explanation WHY it matters]
- [Benefit with explanation WHY it matters]

❌ **Cons:**

- [Drawback with explanation of IMPACT]
- [Drawback with explanation of IMPACT]

📊 **Effort:** Low | Medium | High ⚡ **Impact:** Low | Medium | High ⚠️ **Risk:** Low | Medium | High
```

### Step 3.3: Comparison Matrix

Create comparison matrix:

```markdown
| Criteria      | Option A | Option B | Option C |
| ------------- | -------- | -------- | -------- |
| Effort        | 🟢 Low   | 🟡 Med   | 🔴 High  |
| Impact        | 🟡 Med   | 🟢 High  | 🟢 High  |
| Risk          | 🟢 Low   | 🟡 Med   | 🔴 High  |
| Scalability   | 🟡 Med   | 🟢 High  | 🟢 High  |
| Reversibility | 🟢 Easy  | 🟡 Med   | 🔴 Hard  |
```

---

## Phase 4: Recommendation 💡

### Step 4.1: Synthesize Recommendation

```markdown
## 💡 Recommendation

**Recommended:** Option [X] - [Name]

**Reasoning:**

- [Why this option best fits the constraints]
- [Why the trade-offs are acceptable]
- [What makes this better than alternatives]

**When to reconsider:**

- [Condition that would make Option Y better]
- [Condition that would make Option Z better]
```

### Step 4.2: Request User Decision

```markdown
---

**Would you like to:**

1. ✅ Proceed with Option [X]?
2. 🔄 Explore Option [Y] deeper?
3. 💬 Discuss trade-offs further?
4. 🆕 Generate more options?
```

---

## Phase 5: Decision & Documentation 📝

### Step 5.1: Record Decision

After user selects option:

```markdown
## 📝 Decision Record

**Date:** [YYYY-MM-DD] **Decision:** Option [X] - [Name] **Rationale:** [Why this was chosen]

### Context

[Problem being solved]

### Options Considered

1. [Option A] - [Brief description]
2. [Option B] - [Brief description]
3. [Option C] - [Brief description]

### Chosen Approach

[Detailed description of chosen option]

### Trade-offs Accepted

- [Trade-off 1]
- [Trade-off 2]

### Next Steps

1. [Action item 1]
2. [Action item 2]
```

### Step 5.2: Save & Notify

1. Save to `.agent/docs/BRAINSTORM-{slug}.md`
2. **Slug generation**: Extract 2-3 key words → lowercase → hyphen-separated → max 30 chars
   - "authentication options" → `BRAINSTORM-auth-options.md`
   - "caching strategy" → `BRAINSTORM-caching.md`
3. Notify: `✅ Brainstorm saved: .agent/docs/BRAINSTORM-{slug}.md`

---

## Output Template

```markdown
## 🧠 Brainstorm: [Topic]

### Understanding Lock ✅

**Problem:** [problem statement] **Goal:** [desired outcome] **Constraints:** [limitations]

---

### Option A: [Name]

[Description]

**How it works:**

1. [Step]
2. [Step]

✅ **Pros:** [benefits] ❌ **Cons:** [drawbacks] 📊 **Effort:** Low | Medium | High

---

### Option B: [Name]

[Description]

**How it works:**

1. [Step]
2. [Step]

✅ **Pros:** [benefits] ❌ **Cons:** [drawbacks] 📊 **Effort:** Low | Medium | High

---

### Option C: [Name] _(Unconventional)_

[Description]

**How it works:**

1. [Step]
2. [Step]

✅ **Pros:** [benefits] ❌ **Cons:** [drawbacks] 📊 **Effort:** Low | Medium | High

---

### Comparison Matrix

| Criteria | Option A | Option B | Option C |
| -------- | -------- | -------- | -------- |
| Effort   | 🟢       | 🟡       | 🔴       |
| Impact   | 🟡       | 🟢       | 🟢       |
| Risk     | 🟢       | 🟡       | 🔴       |

---

## 💡 Recommendation

**Option [X]** because [reasoning].

---

**Next:** Proceed? Explore deeper? More options?
```

---

## Quick Reference

### Workflow Flow

```
Understanding Lock → Divergent Thinking → Convergent Analysis → Recommendation → Decision
        ↓                   ↓                    ↓                   ↓              ↓
    Clarify             Generate           Pros/Cons           Synthesize      Document
    + Confirm           3+ options         + Matrix            + Rationale     + Save
```

### Idea Generation Techniques

| When Stuck               | Try This                                  |
| ------------------------ | ----------------------------------------- |
| No ideas                 | First Principles - break down to basics   |
| All options similar      | Inversion - what's the opposite?          |
| Options too conservative | Extremes - what if unlimited resources?   |
| Missing perspective      | Analogies - how does domain X solve this? |
| Want something new       | Combinations - merge two approaches       |

---

## Anti-Patterns (AVOID)

| ❌ Anti-Pattern               | ✅ Instead                                 |
| ----------------------------- | ------------------------------------------ |
| Jump to solutions immediately | Complete Understanding Lock first          |
| Only 1-2 obvious options      | Generate minimum 3, include unconventional |
| Pros without context          | Explain WHY each benefit matters           |
| Cons without mitigation       | Note how to mitigate or accept trade-off   |
| No recommendation             | Always synthesize with clear reasoning     |
| Forgot to document            | Save decision record for future reference  |
