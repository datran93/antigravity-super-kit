import argparse
import shlex
import re
import random

def strip_ansi_codes(text):
    ansi_escape = re.compile(r'(?:\x1B[@-_]|[\x80-\x9F])[0-?]*[ -/]*[@-~]')
    return ansi_escape.sub('', text)

def get_db_connection(db_path):
    conn = sqlite3.connect(db_path)
    conn.execute('''
        CREATE TABLE IF NOT EXISTS messages (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            topic TEXT NOT NULL,
            sender_role TEXT NOT NULL,
            receiver_role TEXT NOT NULL,
            content TEXT NOT NULL,
            is_read BOOLEAN DEFAULT 0,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    ''')
    conn.row_factory = sqlite3.Row
    return conn

def read_unread_messages(db_path, role):
    try:
        conn = get_db_connection(db_path)
        cursor = conn.cursor()
        cursor.execute("SELECT id, sender_role, content FROM messages WHERE receiver_role = ? AND is_read = 0 ORDER BY created_at ASC", (role,))
        rows = cursor.fetchall()

        if rows:
            msg_ids = [r['id'] for r in rows]
            placeholders = ', '.join(['?'] * len(msg_ids))
            cursor.execute(f"UPDATE messages SET is_read = 1 WHERE id IN ({placeholders})", msg_ids)
            conn.commit()

        conn.close()
        return rows
    except Exception as e:
        print(f"DB Error: {e}")
        return []

def log_to_bus(db_path, sender, topic, content):
    try:
        clean_content = strip_ansi_codes(content) if content else ""
        conn = get_db_connection(db_path)
        cursor = conn.cursor()
        cursor.execute("INSERT INTO messages (topic, sender_role, receiver_role, content) VALUES (?, ?, ?, ?)",
                      (topic, sender, "all", clean_content))
        conn.commit()
        conn.close()
    except Exception as e:
        print(f"DB Error log_to_bus: {e}")

def run_opencode(prompt, workspace, db_path, role):
    try:
        # Use kilocode instead of opencode as requested by the user
        cmd = f"kilocode run {shlex.quote(prompt)}"
        my_env = os.environ.copy()

        log_to_bus(db_path, role, f"subagent_{role}", "Executing task block with kilocode...")
        process = subprocess.Popen(
            cmd,
            shell=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            bufsize=1,
            cwd=workspace,
            env=my_env
        )

        for line in iter(process.stdout.readline, ''):
            if line:
                # Log stripped line to bus
                clean_line = strip_ansi_codes(line.strip())
                if clean_line:
                    log_to_bus(db_path, role, f"subagent_{role}", clean_line)

        process.stdout.close()

        process.wait()
        log_to_bus(db_path, role, f"subagent_{role}", f"Task block execution finished. State saved.")
    except Exception as e:
        log_to_bus(db_path, "system", f"subagent_{role}", f"Failed to run kilocode: {e}")

def get_recent_history(db_path, role, limit=10):
    try:
        conn = get_db_connection(db_path)
        cursor = conn.cursor()
        cursor.execute("SELECT sender_role, receiver_role, content FROM messages WHERE sender_role = ? OR receiver_role = ? OR receiver_role = 'all' ORDER BY created_at DESC LIMIT ?", (role, role, limit))
        rows = cursor.fetchall()
        conn.close()
        if not rows: return ""
        history = "=== RECENT HISTORY (MEMORY) ===\n"
        for r in reversed(rows):
            history += f"From {r['sender_role']} to {r['receiver_role']}: {r['content'][:1500]}\n---\n"
        history += "================================\n\n"
        return history
    except Exception as e:
        return ""

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--workspace', required=True)
    parser.add_argument('--role', required=True)
    parser.add_argument('--instruction', required=True)
    parser.add_argument('--task', required=False, default="")
    parser.add_argument('--resume', action='store_true', help="Resume agent with previous memory context")
    args = parser.parse_args()

    log_to_bus(db_path, "system", f"subagent_{args.role}", f"Daemon Worker for ROLE [{args.role}] started (Resume={args.resume}). Listening for messages...")

    # If planner and NOT resuming, execute initial task right away
    if args.role == 'planner' and args.task and not args.resume:
        log_to_bus(db_path, args.role, f"subagent_{args.role}", "I am the Planner. Starting FRESH initial task...")
        prompt = f"SYSTEM INSTRUCTION: {args.instruction}\n\nTASK: {args.task}\n\nDo not ask the user for confirmation, just do it and output summary."
        run_opencode(prompt, args.workspace, db_path, args.role)

    # Core event loop: wait for messages and process them
    while True:
        msgs = read_unread_messages(db_path, args.role)
        if msgs:
            for msg in msgs:
                log_to_bus(db_path, args.role, f"subagent_{args.role}", f"Received message from '{msg['sender_role']}'. Reading it and taking action...")

                # Combine instruction, memory history, and message content
                prompt = f"SYSTEM INSTRUCTION: {args.instruction}\n\n"

                if args.task:
                    prompt += f"YOUR ORIGINAL MISSION/TASK WAS: {args.task}\n\n"

                # Inject recent historical context
                recent_history = get_recent_history(db_path, args.role, limit=10)
                if recent_history:
                    prompt += recent_history

                prompt += f"NEW MESSAGE RECEIVED FROM '{msg['sender_role']}':\n{msg['content']}\n\n"
                prompt += "Please read the message, reflect on your recent history, take appropriate actions via tools, and output a summary. DO NOT ask the user for confirmation."

                run_opencode(prompt, args.workspace, db_path, args.role)

        # Sleep tight to avoid eating CPU, with jitter to avoid rate limits
        time.sleep(random.randint(5, 15))

if __name__ == "__main__":
    main()
