---
description:
  AST Dependency Graph sub-routine. Given a symbol name or file path, returns a 360° view: definition, callers,
  callees, and blast radius. Uses LSP when available, falls back to MCP codebase-explorer. Use before modifying
  shared code.
---

# 🕸️ Codebase Graph

> Lightweight sub-routine — not a full workflow. No Role Anchoring required. Can be invoked inline from any workflow.
>
> **Invoke as**: `/codebase-graph <symbol>` or `/codebase-graph <file_path>`

---

## When to Use

Call this sub-routine when you need to understand the **blast radius** of a change BEFORE writing any code:

- You are about to modify a function/class used in more than one file.
- You are implementing a refactor and need to find all callers.
- You need to trace data flow through a system.
- The Reviewer or Tester flags a symbol as potentially broken across callers.

---

## Execution

### Step 1: Detect LSP Availability 🔍

Try LSP first. It is faster and more precise than semantic search.

```
LSP operations to try (in order of precision):
1. goToDefinition    → locate the canonical definition
2. findReferences    → all usages / callers across the project
3. incomingCalls     → call hierarchy (who calls this function)
4. outgoingCalls     → what this function calls
5. hover             → type signature + docstring
```

**LSP is available** when the project has a language server configured (e.g., `gopls` for Go, `tsserver` for TypeScript,
`pyright` for Python, `rust-analyzer` for Rust). Use the `LSP` tool with the appropriate operation.

**How to detect**: Attempt `LSP.hover` at the symbol position. If it returns type info → LSP is active. If it errors
with "no server" → fall back to MCP.

### Step 2: MCP Fallback (if LSP unavailable) 🛡️

When LSP is not available, use `@mcp:codebase-explorer` tools:

```
1. search_symbol(query=<symbol>)    → find definition file + line
2. find_usages(symbol_name=<symbol>) → all references grouped by file
3. context(name=<symbol>)           → 360° view: definition + usages + semantic chunks
```

### Step 3: Compile & Output the Graph 📊

Produce a concise **Symbol Graph Report** in this format:

```
## 🕸️ Symbol: <SymbolName>
**File**: path/to/definition.ext (line N)
**Type**: function | class | interface | variable
**Signature**: <type signature or function signature>

### Callers (incomingCalls / find_usages)
- path/to/caller_a.ext:42 — brief context of how it calls the symbol
- path/to/caller_b.ext:88 — brief context

### Callees (outgoingCalls)
- path/to/dep_x.ext — <SymbolX> called at line N
- path/to/dep_y.ext — <SymbolY> called at line N

### Blast Radius Assessment
⚠️ HIGH   — used in N≥5 files / touches auth or DB
🟡 MEDIUM — used in 2–4 files
🟢 LOW    — used in 1 file (or private/unexported)

### Recommendation
[What the caller should be careful about when modifying this symbol]
```

---

## Integration Notes

- **Coder** calls this in Phase 2 (Pattern & Dependency Discovery) before editing shared symbols.
- **Reviewer** calls this in Phase 1 (Mechanical Checks) via `search_code` to verify callers are unbroken.
- **Planner** may call this in Phase 1 (Discovery) to assess blast radius of a proposed change.

---

## 🔴 Constraints

1. **Read-only**: This sub-routine NEVER modifies files.
2. **LSP preferred**: Always try LSP first. Only fall back to MCP if LSP is unavailable.
3. **Output only the graph**: Keep the report concise. No implementation advice, no code.
4. **Blast radius MUST be stated**: Never omit the assessment. It is the primary output.
