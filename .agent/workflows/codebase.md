---
description: Analyze specific codebase folder and document it.
---

# /codebase - Codebase Documentation Workflow

Analyze a specific folder and document its structure, purpose, and key components. Results are saved to the `agent-docs/` folder at the root of the current project directory as `CODEBASE-{folder-name}.md`.

---

## When to Use

- `/codebase <path>` - Analyze a specific folder (e.g., `/codebase ./services/auth`, `/codebase frontend`) and save documentation to `./agent-docs/CODEBASE-{folder-name}.md`.
- Keywords: "analyze folder", "document module", "explain this part of the codebase"

---

## Phase 1: Input Analysis & Context 📂

### Step 1.1: Parse Input

Extract the target folder path from the user request.

- **Target Path**: `[path/to/folder]`
- **Folder Name**: `[last-segment-of-path]` (e.g., `services/auth` -> `auth`)
- **Output File**: `agent-docs/CODEBASE-{folder-name}.md`

### Step 1.2: Check Existing Documentation

Check if `./agent-docs/CODEBASE-{folder-name}.md` already exists.

- **Status**: `EXISTS` or `MISSING`

---

## Phase 2: Codebase Analysis 🔍

**Action**: Explore the **Target Path**.

### Step 2.1: Project Identity

Identify using files *within the target folder*:

| Item          | Look for...                               |
| ------------- | ----------------------------------------- |
| **Type**      | Service, Library, CLI, UI Component, etc. |
| **Language**  | File extensions (.ts, .go, .py, etc.)     |
| **Framework** | Importing specific libraries              |

### Step 2.2: Structural Map

Map key subdirectories and files *relative to the target folder*.

```
[folder-name]/
├── [subdir]/      # Purpose
├── [file].ext     # Purpose
└── ...
```

### Step 2.3: Key Components

Identify important files:

- **Entry Points**: (e.g., `index.ts`, `main.go`)
- **Core Logic**: (e.g., `service.ts`, `handler.go`)
- **Configuration**: (e.g., `config.json`, `.env.example`)
- **Tests**: (e.g., `*.test.ts`, `*_test.go`)

### Step 2.4: Dependencies (Local & External)

- **External**: Libraries imported from package managers.
- **Internal**: Imports from other parts of the monorepo (e.g., `../../shared`).

---

## Phase 3: Documentation Generation 📝

### Step 3.1: Define Content

**Template for `agent-docs/CODEBASE-{folder-name}.md`**:

```markdown
# Codebase Analysis: [Folder Name]

> **Path**: `[Target Path]`
> **Last Updated**: [YYYY-MM-DD]

## Overview

[Brief description of what this folder contains and its role in the larger project.]

## Tech Stack

- **Language**: [Language]
- **Type**: [Service/Library/etc.]
- **Key Frameworks**: [List]

## Directory Structure

\`\`\`
[Tree structure of the folder]
\`\`\`

## Key Components

### Entry Points
- \`[File]\`: [Description]

### Core Logic
- \`[File]\`: [Description]

## Dependencies

- **Internal**: [List imports from other project folders]
- **External**: [Key external libraries]

## Architecture Notes

[Notes on data flow, patterns used, or specific implementation details.]
```

---

## Phase 4: Create or Update 🔄

### Step 4.1: CREATE Mode (If file missing)

1. Create `./agent-docs/CODEBASE-{folder-name}.md`.
2. Fill with the generated content.

### Step 4.2: UPDATE Mode (If file exists)

1. **Read** existing `./agent-docs/CODEBASE-{folder-name}.md`.
2. **Compare** with current analysis.
3. **Update**:
   - Update structure tree if changed.
   - Add new files/components.
   - Update descriptions if logic changed.
   - **Keep** existing manual notes or descriptions that are still valid.
   - Update "Last Updated" date.

---

## Phase 5: Save & Notify 💾

### Step 5.1: Save File

Save the content to `./agent-docs/CODEBASE-{folder-name}.md`.

### Step 5.2: Notify User

```markdown
✅ **Codebase Documented**: `./agent-docs/CODEBASE-{folder-name}.md`

**Status**: [Created / Updated]
**Target**: `[Target Path]`
```

---

## Quick Reference

### Workflow Flow

```
Parse Path → Check Exists → Analyze Folder → Generate Content → Save → Notify
     ↓             ↓              ↓                 ↓             ↓
Folder Name    Create/Update    Structure       Template      CODEBASE-{name}.md
```

### Anti-Patterns (AVOID)

| ❌ Don't                                 | ✅ Do                                     |
| --------------------------------------- | ---------------------------------------- |
| Analyze entire repo if path is specific | Focus ONLY on the target folder          |
| Overwrite manual notes                  | Preserve existing insights during update |
| List every single file                  | Focus on KEY files                       |
