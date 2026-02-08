---
description: Preview server start, stop, and status check. Local development server management.
---

# /preview - Local Development Workflow

Guide agents to manage local development environments: start, stop, monitor, and troubleshoot.

---

## When to Use

- `/preview start` - **Start local server**
- `/preview stop` - Stop server
- `/preview restart` - Restart server
- `/preview url` - Get running URL
- `/preview logs` - Check recent output
- `/preview status` - Full status report

---

## Phase 1: Configuration & Detection ⚙️

### Step 1.1: Detect Start Command

Identify the correct command to start the project:

```markdown
### Project Detection

| Type        | Command                              | Priority |
| :---------- | :----------------------------------- | :------- |
| **Node.js** | `npm run dev` / `yarn dev`           | 1        |
| **Python**  | `python app.py` / `uvicorn main:app` | 1        |
| **Go**      | `go run .`                           | 1        |
| **Docker**  | `docker compose up`                  | 2        |
| **Custom**  | `make run`                           | 3        |
```

### Step 1.2: Check Environment

Ensure required ports and variables are available:

- [ ] **Port:** Is default port (e.g., 3000, 8000) free?
- [ ] **Env Vars:** Is `.env` file present?
- [ ] **Dependencies:** Are `node_modules` / `venv` installed?

---

## Phase 2: Execution 🚀

### Step 2.1: Start Process

Execute the start command in a background process (or new terminal tab).

```bash
# Example: Start in background
nohup [start_command] > .agent/logs/preview.log 2>&1 &
echo $! > .agent/logs/preview.pid
```

### Step 2.2: Stream Logs

Monitor output for "Ready" signal:

```markdown
### Log Monitoring

- **Wait for:** "Ready on http://..." or "Server started"
- **Timeout:** 30 seconds
- **Success:** URL detectable
- **Failure:** Process exits or prints error
```

---

## Phase 3: Verification ✅

### Step 3.1: Health Check

Verify the server is actually responding:

```bash
# Check if port is listening
lsof -i :[port]

# Check response (if HTTP)
curl -s -o /dev/null -w "%{http_code}" http://localhost:[port]
```

### Step 3.2: URL Discovery

Report the specific URL for user access:

```markdown
### 🌐 Preview Ready

- **Local:** `http://localhost:[port]`
- **Network:** `http://[ip]:[port]` (if available)
```

---

## Phase 4: Troubleshooting 🔧

### Issue: Port in Use

If detecting `EADDRINUSE` or `Address already in use`:

1. **Identify PID:** `lsof -ti :[port]`
2. **Kill Process:** `kill -9 [PID]`
3. **Retry Start**

### Issue: Start Crash

If process exits immediately:

1. **Read Logs:** `cat .agent/logs/preview.log`
2. **Fix Missing Env:** check `.env`
3. **Fix Missing Dep:** run install command (`npm install`, `pip install`)

---

## Phase 5: Reporting 📝

### Step 5.1: Status Report

Generata a status report for the preview environment.

```markdown
# 🌐 Preview Status Report

## 🟢 System Status

- **State:** Running | Stopped | Error
- **URL:** [http://localhost:port]
- **PID:** [Process ID]
- **Port:** [Port Number]

## 📋 Configuration

- **Command:** `[Start Command]`
- **Project Type:** [Node/Python/Go/Docker]

## 🩺 Logic Check

- **Port Listening:** ✅ Yes | ❌ No
- **HTTP Response:** [Status Code]
```

### Step 5.2: Save & Notify (Optional)

If requested to generate a report:

1. Save report to `agent-docs/PREVIEW-{slug}.md`
2. **Slug generation**: "status-[timestamp]" → `PREVIEW-status-20240101.md`
3. Notify: `✅ Preview status saved: agent-docs/PREVIEW-{slug}.md`

---

## Quick Reference

### Commands map

| Action     | Command Pattern                        |
| :--------- | :------------------------------------- |
| **Start**  | `[package_manager] run dev`            |
| **Stop**   | `kill $(cat .agent/logs/preview.pid)`  |
| **Logs**   | `tail -f .agent/logs/preview.log`      |
| **Status** | `ps -p $(cat .agent/logs/preview.pid)` |

### Status Indicators

| Indicator       | Meaning                                            |
| :-------------- | :------------------------------------------------- |
| 🟢 **Running**  | Process active, port listening.                    |
| 🔴 **Stopped**  | No process found.                                  |
| ⚠️ **Error**    | Process active but port unreachable (or crashing). |
| 🔄 **Starting** | Process active, waiting for ready signal.          |
