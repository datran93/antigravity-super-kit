# Antigravity Kit

> AI Agent templates with Skills, Agents, and Workflows

## Quick Install

Clone the repository and link the `agk` script to your path for global access:

```bash
# 1. Clone the repository (if not already done)
git clone git@github.com:Dang-Hai-Tran/antigravity-kit.git ~/.antigravity/cache/antigravity-kit

# 2. Link the script to your bin
sudo ln -s ~/.antigravity/cache/antigravity-kit/scripts/agk /usr/local/bin/agk

# 3. Verify installation
agk
```

### ⚠️ Important Note on `.gitignore`

If you are using AI-powered editors like **Cursor** or **Windsurf**, adding the `.agent/` folder to your `.gitignore`
may prevent the IDE from indexing the workflows. This results in slash commands (like `/plan`, `/debug`) not appearing
in the chat suggestion dropdown.

**Recommended Solution:** To keep the `.agent/` folder local (not tracked by Git) while maintaining AI functionality:

1. Ensure `.agent/` is **NOT** in your project's `.gitignore`.
2. Instead, add it to your local exclude file: `.git/info/exclude`

## What's Included

| Component     | Count | Description                                                        |
| ------------- | ----- | ------------------------------------------------------------------ |
| **Agents**    | 20    | Specialist AI personas (frontend, backend, security, PM, QA, etc.) |
| **Skills**    | 37    | Domain-specific knowledge modules                                  |
| **Workflows** | 11    | Slash command procedures                                           |

## Usage

### 1. Managing .agent Folder

Run these commands in any project root:

| Command                | Description                                    |
| ---------------------- | ---------------------------------------------- |
| `agk install`          | Install .agent folder (auto-excludes from git) |
| `agk install --force`  | Force install, overwrite existing              |
| `agk update`           | Update .agent folder to latest version         |
| `agk update --dry-run` | Preview changes without applying               |
| `agk update --offline` | Update using cached repo (no network)          |
| `agk status`           | Check if updates are available                 |
| `agk remove`           | Remove .agent folder (with confirmation)       |
| `agk remove --force`   | Remove without confirmation                    |
| `agk version`          | Show version and cache commit                  |
| `agk --help`           | Show full help with examples                   |

```bash
# Install .agent folder
agk install

# Force install (overwrite existing)
agk install --force

# Update to latest version
agk update

# Preview what would change
agk update --dry-run

# Update without internet (use cache)
agk update --offline

# Check for updates
agk status

# Remove .agent folder
agk remove
```

> **Note:** The `docs/` folder inside `.agent/` is preserved during updates.

### 2. Using Agents

**No need to mention agents explicitly!** The system automatically detects and applies the right specialist(s):

```
You: "Add JWT authentication"
AI: 🤖 Applying @security-auditor + @backend-specialist...

You: "Fix the dark mode button"
AI: 🤖 Using @frontend-specialist...

You: "Login returns 500 error"
AI: 🤖 Using @debugger for systematic analysis...
```

**How it works:**

- Analyzes your request silently

- Detects domain(s) automatically (frontend, backend, security, etc.)
- Selects the best specialist(s)
- Informs you which expertise is being applied
- You get specialist-level responses without needing to know the system architecture

**Benefits:**

- ✅ Zero learning curve - just describe what you need
- ✅ Always get expert responses
- ✅ Transparent - shows which agent is being used
- ✅ Can still override by mentioning agent explicitly

### 3. Using Workflows

Invoke workflows with slash commands:

| Command          | Description                           |
| ---------------- | ------------------------------------- |
| `/brainstorm`    | Explore options before implementation |
| `/create`        | Create new features or apps           |
| `/debug`         | Systematic debugging                  |
| `/deploy`        | Deploy application                    |
| `/enhance`       | Improve existing code                 |
| `/orchestrate`   | Multi-agent coordination              |
| `/plan`          | Create task breakdown                 |
| `/preview`       | Preview changes locally               |
| `/status`        | Check project status                  |
| `/test`          | Generate and run tests                |
| `/ui-ux-pro-max` | Design with 50 styles                 |

Example:

```
/brainstorm authentication system
/create landing page with hero section
/debug why login fails
```

### 4. Using Skills

Skills are loaded automatically based on task context. The AI reads skill descriptions and applies relevant knowledge.

## License

MIT © datran
