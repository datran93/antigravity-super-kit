---
description:
  Add or update features in existing application. Used for iterative
  development.
---

# /enhance - Update Application

$ARGUMENTS

---

## Task

This command adds features or makes updates to existing application.

### Steps:

1. **Understand Current State**
   - Load project state with `python .agent/scripts/session_manager.py info`
   - Understand existing features, tech stack

2. **Plan Changes**
   - Determine what will be added/changed
   - Detect affected files
   - Check dependencies

3. **Present Plan to User** (for major changes)

   ```
   "To add admin panel:
   - I'll create 15 new files
   - Update 8 files
   - Takes ~10 minutes

   Should I start?"
   ```

4. **Apply**
   - Call relevant agents
   - Make changes
   - Test

5. **Update Preview**
   - Hot reload or restart

6. **Save & Notify**
   - Save enhancement summary to `.agent/docs/ENHANCE-{slug}.md`
   - **Slug generation**: Extract 2-3 key words from feature → lowercase → hyphen-separated → max 30 chars
     - Example: "add dark mode" → `ENHANCE-dark-mode.md`
   - Notify user: `✅ ENHANCE report saved: .agent/docs/ENHANCE-{slug}.md`

---

## Usage Examples

```
/enhance add dark mode
/enhance build admin panel
/enhance integrate payment system
/enhance add search feature
/enhance edit profile page
/enhance make responsive
```

---

## Caution

- Get approval for major changes
- Warn on conflicting requests (e.g., "use Firebase" when project uses
  PostgreSQL)
- Commit each change with git
