import os
import sqlite3
import json
import traceback
from contextlib import closing
from datetime import datetime
from mcp.server.fastmcp import FastMCP
from typing import List, Optional

mcp = FastMCP("McpContextManager")

def get_db_connection(workspace_path: str):
    if not workspace_path:
        raise ValueError("workspace_path is required")

    # Force absolute paths
    workspace_path = os.path.abspath(workspace_path)

    db_path = os.path.join(workspace_path, "context.db")

    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    cursor = conn.cursor()
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS checkpoints (
            task_id TEXT PRIMARY KEY,
            description TEXT,
            status TEXT,
            completed_steps TEXT,
            next_steps TEXT,
            active_files TEXT,
            notes TEXT,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    ''')
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS intents (
            task_id TEXT PRIMARY KEY,
            tactic TEXT,
            locked_files TEXT,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    ''')
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS drift_tracker (
            task_id TEXT PRIMARY KEY,
            failure_count INTEGER DEFAULT 0,
            last_failed_at TIMESTAMP
        )
    ''')
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS anchors (
            key TEXT PRIMARY KEY,
            value TEXT,
            rule TEXT,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    ''')
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS file_annotations (
            file_path TEXT PRIMARY KEY,
            gotchas TEXT,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    ''')
    cursor.execute('''
        CREATE VIRTUAL TABLE IF NOT EXISTS knowledge_fts USING fts5(
            tactic_name,
            ki_path UNINDEXED,
            summary,
            decisions
        )
    ''')
    conn.commit()
    return conn

def write_markdown_progress(workspace_path, task_id, description, status, completed_steps, next_steps, active_files, notes):
    try:
        md_path = os.path.join(workspace_path, "progress.md")

        # Get history of other completed tasks
        with closing(get_db_connection(workspace_path)) as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT task_id, description FROM checkpoints WHERE status = 'completed' AND task_id != ? ORDER BY updated_at DESC", (task_id,))
            historical_tasks = cursor.fetchall()

        total_steps = len(completed_steps) + len(next_steps)
        progress_pct = (len(completed_steps) / total_steps * 100) if total_steps > 0 else 0

        # Progress bar
        bar_len = 20
        filled = int(bar_len * progress_pct // 100)
        bar = "█" * filled + "░" * (bar_len - filled)

        content = f"# 🚀 Project Progress: {task_id}\n\n"
        content += f"**Status:** `{status.upper()}` | **Progress:** `[{bar}] {progress_pct:.1f}%` ({len(completed_steps)}/{total_steps})\n\n"
        content += f"> {description}\n\n"



        if active_files:
            content += "### 📁 Active Files\n"
            for f in active_files: content += f"- `{f}`\n"
            content += "\n"

        content += "### ✅ Completed\n"
        if not completed_steps: content += "_None yet_\n"
        else:
            for s in completed_steps: content += f"- [x] {s}\n"
        content += "\n"

        content += "### ⏳ Next Steps\n"
        if not next_steps: content += "_All tasks done!_ 🎉\n"
        else:
            for s in next_steps: content += f"- [ ] {s}\n"
        content += "\n"

        if notes:
            content += "### 📝 Log & Notes\n"
            content += f"```text\n{notes}\n```\n\n"

        if historical_tasks:
            content += "---\n### 🏆 Historically Completed Tasks\n"
            for t in historical_tasks:
                content += f"- **{t['task_id']}**: {t['description']}\n"
            content += "\n"

        content += f"---\n*Last sync: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}*"

        with open(md_path, 'w', encoding='utf-8') as f:
            f.write(content)
    except Exception as e:
        print(f"Error writing progress.md: {e}")

@mcp.tool()
def save_checkpoint(
    workspace_path: str,
    task_id: str,
    description: str,
    status: str,
    completed_steps: List[str],
    next_steps: List[str],
    active_files: List[str],
    notes: str
) -> str:
    """
    Save or update a task checkpoint/context.
    """
    try:
        with closing(get_db_connection(workspace_path)) as conn:
            cursor = conn.cursor()
            now = datetime.now().isoformat()

            cursor.execute('''
                INSERT INTO checkpoints (task_id, description, status, completed_steps, next_steps, active_files, notes, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                ON CONFLICT(task_id) DO UPDATE SET
                    description=excluded.description,
                    status=excluded.status,
                    completed_steps=excluded.completed_steps,
                    next_steps=excluded.next_steps,
                    active_files=excluded.active_files,
                    notes=excluded.notes,
                    updated_at=excluded.updated_at
            ''', (task_id, description, status, json.dumps(completed_steps), json.dumps(next_steps), json.dumps(active_files), notes, now))

            conn.commit()

        write_markdown_progress(workspace_path, task_id, description, status, completed_steps, next_steps, active_files, notes)

        msg = f"✅ Checkpoint '{task_id}' saved."
        if not next_steps and completed_steps:
            msg += "\n\n🎉 ALL TASKS COMPLETED! Great job."
        return msg
    except Exception as e:
        return f"❌ Error saving checkpoint: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def initialize_task_plan(workspace_path: str, task_id: str, description: str, steps: List[str]) -> str:
    """Start a new task with a list of steps."""
    init_notes = f"[{datetime.now().strftime('%H:%M:%S')}] Task started."
    return save_checkpoint(
        workspace_path=workspace_path,
        task_id=task_id,
        description=description,
        status="in_progress",
        completed_steps=[],
        next_steps=steps,
        active_files=[],
        notes=init_notes
    )

@mcp.tool()
def complete_task_step(workspace_path: str, task_id: str, step_name: str, active_files: Optional[List[str]] = None, notes: Optional[str] = None) -> str:
    """Mark step as done, track active files, update graph and bar."""
    try:
        with closing(get_db_connection(workspace_path)) as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT * FROM checkpoints WHERE task_id = ?", (task_id,))
            row = cursor.fetchone()
            if not row:
                return f"❌ Task '{task_id}' not found."

            comp = json.loads(row['completed_steps'])
            nxt = json.loads(row['next_steps'])
            curr_active_files = json.loads(row['active_files']) if row['active_files'] else []

            if step_name in nxt:
                nxt.remove(step_name)
                comp.append(step_name)
            else:
                return f"⚠️ Step '{step_name}' not in queue."

            stat = "completed" if not nxt else row['status']

            # Track Time & Log
            end_time_str = datetime.now().strftime('%H:%M:%S')
            log = row['notes'] + f"\n[{end_time_str}] ✅ Done: {step_name}"

            if active_files:
                log += f"\n  - Files: {', '.join(active_files)}"
                for f in active_files:
                    if f not in curr_active_files:
                        curr_active_files.append(f)

            if notes:
                log += f"\n  - Notes: {notes}"

        return save_checkpoint(
            workspace_path=workspace_path,
            task_id=task_id,
            description=row['description'],
            status=stat,
            completed_steps=comp,
            next_steps=nxt,
            active_files=curr_active_files,
            notes=log
        )
    except Exception as e:
        return f"❌ Error completing step: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def add_task_step(workspace_path: str, task_id: str, new_step: str) -> str:
    """Add a new task step to the next_steps list of an existing task."""
    try:
        with closing(get_db_connection(workspace_path)) as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT * FROM checkpoints WHERE task_id = ?", (task_id,))
            row = cursor.fetchone()
            if not row:
                return f"❌ Task '{task_id}' not found."

            nxt = json.loads(row['next_steps'])
            comp = json.loads(row['completed_steps'])

            if new_step in nxt or new_step in comp:
                return f"⚠️ Step '{new_step}' already exists in task '{task_id}'."

            nxt.append(new_step)

            log = row['notes'] + f"\n[{datetime.now().strftime('%H:%M:%S')}] Added new step: {new_step}"

        return save_checkpoint(
            workspace_path=workspace_path,
            task_id=task_id,
            description=row['description'],
            status=row['status'],
            completed_steps=comp,
            next_steps=nxt,
            active_files=json.loads(row['active_files']),
            notes=log
        )
    except Exception as e:
        return f"❌ Error adding step: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def load_checkpoint(workspace_path: str, task_id: str) -> str:
    """
    Load a previously saved task checkpoint.
    """
    try:
        with closing(get_db_connection(workspace_path)) as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT * FROM checkpoints WHERE task_id = ?", (task_id,))
            row = cursor.fetchone()

        if not row:
            return f"❌ Checkpoint '{task_id}' not found."

        res = [f"🔄 {row['task_id']} [{row['status'].upper()}]", f"Last Update: {row['updated_at']}\n"]

        comp = json.loads(row['completed_steps'])
        nxt = json.loads(row['next_steps'])
        total = len(comp) + len(nxt)
        pct = (len(comp)/total*100) if total > 0 else 0
        res.append(f"Progress: {pct:.1f}% ({len(comp)}/{total} steps)\n")

        res.append("## ✅ Completed")
        for s in comp: res.append(f"- [x] {s}")

        res.append("\n## ⏳ Next")
        for s in nxt: res.append(f"- [ ] {s}")

        res.append(f"\n## 📝 Notes\n{row['notes']}")
        return "\n".join(res)
    except Exception as e:
        return f"❌ Error loading: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def list_active_tasks(workspace_path: str) -> str:
    """
    List all active tasks.
    """
    try:
        with closing(get_db_connection(workspace_path)) as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT task_id, status, updated_at FROM checkpoints ORDER BY updated_at DESC")
            rows = cursor.fetchall()

        if not rows: return "No tasks found."

        res = ["📋 WORKSPACE TASKS:"]
        for r in rows:
            res.append(f"- **{r['task_id']}** ({r['status']}) - {r['updated_at']}")
        return "\n".join(res)
    except Exception as e:
        return f"❌ Error: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def declare_intent(workspace_path: str, task_id: str, tactic: str, locked_files: List[str]) -> str:
    """Declare working intention and restrict changes to specific locked_files."""
    try:
        with closing(get_db_connection(workspace_path)) as conn:
            cursor = conn.cursor()
            cursor.execute('''
                INSERT INTO intents (task_id, tactic, locked_files)
                VALUES (?, ?, ?)
                ON CONFLICT(task_id) DO UPDATE SET
                    tactic=excluded.tactic,
                    locked_files=excluded.locked_files,
                    updated_at=CURRENT_TIMESTAMP
            ''', (task_id, tactic, json.dumps(locked_files)))
            conn.commit()
        return f"🔒 Intent declared. Lock applied to files: {', '.join(locked_files)}"
    except Exception as e:
        return f"❌ Error declaring intent: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def check_intent_lock(workspace_path: str, task_id: str, file_path: str) -> str:
    """Check if a file is unlocked for the current intent."""
    try:
        with closing(get_db_connection(workspace_path)) as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT locked_files FROM intents WHERE task_id = ?", (task_id,))
            row = cursor.fetchone()
            if not row: return "⚠️ No intent declared for this task. Please declare_intent first."
            files = json.loads(row['locked_files'])

            is_locked = any(f in file_path or file_path in f for f in files)
            if is_locked:
                cursor.execute("SELECT gotchas FROM file_annotations WHERE file_path = ?", (file_path,))
                ann_row = cursor.fetchone()
                ghost_ctx = f"\n👻 GHOST CONTEXT: {ann_row['gotchas']}" if ann_row else ""
                return f"✅ File '{file_path}' is unlocked for current intent.{ghost_ctx}"
            return f"❌ ALARM: Scope Creep! File '{file_path}' is NOT in the active_files lock. Switch to Planner to update intent via declare_intent."
    except Exception as e:
        return f"❌ Error checking lock: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def annotate_file(workspace_path: str, file_path: str, gotchas: str) -> str:
    """Add 'Ghost Context' (lessons/gotchas) to a specific file, ensuring the Agent remembers them next time it touches this file."""
    try:
        with closing(get_db_connection(workspace_path)) as conn:
            cursor = conn.cursor()
            cursor.execute('''
                INSERT INTO file_annotations (file_path, gotchas)
                VALUES (?, ?)
                ON CONFLICT(file_path) DO UPDATE SET
                    gotchas=excluded.gotchas,
                    updated_at=CURRENT_TIMESTAMP
            ''', (file_path, gotchas))
            conn.commit()
        return f"👻 Ghost Context added to '{file_path}'."
    except Exception as e:
        return f"❌ Error annotating file: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def recall_knowledge(workspace_path: str, query: str, top_k: int = 3) -> str:
    """Recall relevant Knowledge Items (Local RAG) based on a search query using SQLite FTS5."""
    try:
        with closing(get_db_connection(workspace_path)) as conn:
            cursor = conn.cursor()

            clean_query = ''.join(e for e in query if e.isalnum() or e.isspace())
            tokens = [f"{t}*" for t in clean_query.split() if t.strip()]
            fts_query = " OR ".join(tokens)

            if not fts_query:
                return "🔍 Please provide a valid search query."

            cursor.execute('''
                SELECT tactic_name, ki_path, summary, decisions
                FROM knowledge_fts
                WHERE knowledge_fts MATCH ?
                ORDER BY rank
                LIMIT ?
            ''', (fts_query, top_k))

            rows = cursor.fetchall()
            if not rows:
                return f"🔍 No relevant Knowledge Items found for query: '{query}'"

            res = [f"🧠 Recalled Knowledge for '{query}':\n"]
            for r in rows:
                res.append(f"### KI: {r['tactic_name']}")
                res.append(f"**Path**: `{r['ki_path']}`")
                res.append(f"**Summary**: {r['summary']}")
                res.append(f"**Decisions**: {r['decisions']}\n---")

            return "\n".join(res)
    except Exception as e:
        return f"❌ Error recalling knowledge: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def compact_memory(workspace_path: str, task_id: str, tactic_name: str, summary: str, decisions: str) -> str:
    """Generate a Knowledge Item (KI) and prune context at the end of a tactic. Also indexes KI into Semantic SQLite."""
    try:
        knowledge_dir = os.path.join(workspace_path, "knowledge")
        os.makedirs(knowledge_dir, exist_ok=True)

        safe_name = tactic_name.replace(" ", "_").lower()
        ki_path = os.path.join(knowledge_dir, f"{safe_name}.md")

        with closing(get_db_connection(workspace_path)) as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT locked_files FROM intents WHERE task_id = ?", (task_id,))
            row = cursor.fetchone()
            files = json.loads(row['locked_files']) if row else []

            content = f"# KI: {tactic_name}\n\n## Summary\n{summary}\n\n## Affected Files\n"
            for f in files: content += f"- `{f}`\n"
            content += f"\n## Architecture & Decisions\n{decisions}\n"

            with open(ki_path, 'w', encoding='utf-8') as f:
                f.write(content)

            # Index into FTS5 for Local RAG!
            cursor.execute('''
                INSERT INTO knowledge_fts (tactic_name, ki_path, summary, decisions)
                VALUES (?, ?, ?, ?)
            ''', (tactic_name, ki_path, summary, decisions))

            cursor.execute("SELECT * FROM checkpoints WHERE task_id = ?", (task_id,))
            chk_row = cursor.fetchone()
            if chk_row:
                new_notes = chk_row['notes'] + f"\n[COMPACTION] Tactic '{tactic_name}' completed. KI saved to {ki_path}"
                cursor.execute("UPDATE checkpoints SET active_files='[]', notes=? WHERE task_id=?", (new_notes, task_id))
                cursor.execute("UPDATE intents SET locked_files='[]' WHERE task_id=?", (task_id,))
                cursor.execute("UPDATE drift_tracker SET failure_count=0 WHERE task_id=?", (task_id,))

            conn.commit()

        return f"🗜️ Context Compaction successful. Knowledge Item indexed into local RAG and saved to {ki_path}. Memory flushed."
    except Exception as e:
        return f"❌ Error compacting memory: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def record_failure(workspace_path: str, task_id: str) -> str:
    """Record a failure (e.g. test failing, compile error) to detect context drift."""
    try:
        with closing(get_db_connection(workspace_path)) as conn:
            cursor = conn.cursor()
            cursor.execute('''
                INSERT INTO drift_tracker (task_id, failure_count, last_failed_at)
                VALUES (?, 1, CURRENT_TIMESTAMP)
                ON CONFLICT(task_id) DO UPDATE SET
                    failure_count=failure_count + 1,
                    last_failed_at=CURRENT_TIMESTAMP
            ''', (task_id,))
            cursor.execute("SELECT failure_count FROM drift_tracker WHERE task_id=?", (task_id,))
            count = cursor.fetchone()['failure_count']
            conn.commit()

            if count >= 3:
                return f"🚨 DRIFT DETECTED (Failures: {count}): You have failed 3 or more times. Trigger 'think_back' panic protocol immediately!"
            return f"⚠️ Failure recorded. Count: {count}"
    except Exception as e:
        return f"❌ Error tracking drift: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def clear_drift(workspace_path: str, task_id: str) -> str:
    """Reset the failure counter after a success or planner intervention."""
    try:
        with closing(get_db_connection(workspace_path)) as conn:
            cursor = conn.cursor()
            cursor.execute("UPDATE drift_tracker SET failure_count=0 WHERE task_id=?", (task_id,))
            conn.commit()
        return "🧹 Drift counter reset to 0."
    except Exception as e:
        return f"❌ Error resetting drift: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def manage_anchors(workspace_path: str, action: str, key: str = "", value: str = "", rule: str = "") -> str:
    """Manage architectural anchors. Action can be 'set', 'get', or 'list'."""
    try:
        with closing(get_db_connection(workspace_path)) as conn:
            cursor = conn.cursor()
            if action == "set":
                if not key or not value: return "❌ Key and value required for 'set'."
                cursor.execute('''
                    INSERT INTO anchors (key, value, rule)
                    VALUES (?, ?, ?)
                    ON CONFLICT(key) DO UPDATE SET
                        value=excluded.value,
                        rule=excluded.rule,
                        updated_at=CURRENT_TIMESTAMP
                ''', (key, value, rule))
                conn.commit()
                return f"⚓ Anchor '{key}' secured successfully."
            elif action == "get":
                cursor.execute("SELECT * FROM anchors WHERE key = ?", (key,))
                row = cursor.fetchone()
                if not row: return f"⚠️ Anchor '{key}' not found."
                return f"⚓ ANCHOR [{key}]: {row['value']} (Rule: {row['rule']})"
            elif action == "list":
                cursor.execute("SELECT * FROM anchors ORDER BY key")
                rows = cursor.fetchall()
                if not rows: return "No anchors defined."
                res = ["⚓ PROJECT ANCHORS:"]
                for r in rows:
                    res.append(f"- **{r['key']}**: {r['value']}")
                    if r['rule']: res.append(f"  Rule: {r['rule']}")
                return "\n".join(res)
            else:
                return "❌ Unknown action. Use 'set', 'get', or 'list'."
    except Exception as e:
        return f"❌ Error managing anchors: {str(e)}\n{traceback.format_exc()}"

if __name__ == "__main__":
    mcp.run(transport='stdio')
