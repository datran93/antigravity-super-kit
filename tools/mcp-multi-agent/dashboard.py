import os
import sqlite3
from fastapi import FastAPI, Request
from fastapi.templating import Jinja2Templates
from fastapi.staticfiles import StaticFiles
from fastapi.responses import HTMLResponse
import uvicorn

app = FastAPI(title="Multi-Agent Interactor Dashboard")
BASE_DIR = os.path.dirname(os.path.abspath(__file__))
templates = Jinja2Templates(directory=os.path.join(BASE_DIR, "templates"))

# DB Path configuration - reads from env or defaults to getting from the executing directory
import sys
if len(sys.argv) > 1 and sys.argv[1].endswith(".db"):
    DB_PATH = sys.argv[1]
else:
    DB_PATH = os.environ.get("MULTI_AGENT_DB_PATH", os.path.join(os.getcwd(), ".agent_logs", "multi_agent_bus.db"))

def get_db_connection():
    if not os.path.exists(DB_PATH):
        # Create an empty db file if it doesn't exist yet but the dashboard was started
        os.makedirs(os.path.dirname(DB_PATH), exist_ok=True)
        conn = sqlite3.connect(DB_PATH)
        cursor = conn.cursor()
        cursor.execute('''
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
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS agent_status (
                role TEXT PRIMARY KEY,
                status TEXT NOT NULL,
                last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                current_task TEXT
            )
        ''')
        conn.commit()
        conn.close()

    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    return conn

@app.get("/", response_class=HTMLResponse)
async def index(request: Request):
    return templates.TemplateResponse("index.html", {"request": request, "db_path": DB_PATH})

@app.get("/api/messages")
def get_messages(topic: str = None, unread_only: bool = False):
    try:
        conn = get_db_connection()
        cursor = conn.cursor()
        query = "SELECT * FROM messages WHERE 1=1"
        params = []

        if topic:
            query += " AND topic = ?"
            params.append(topic)

        if unread_only:
            query += " AND is_read = 0"

        query += " ORDER BY created_at DESC LIMIT 100"

        cursor.execute(query, params)
        rows = cursor.fetchall()
        conn.close()

        # Format as list of dicts
        messages = [dict(r) for r in rows]
        # Reverse to show chronological order
        return {"messages": messages[::-1]}
    except Exception as e:
        return {"error": str(e), "messages": []}

@app.get("/api/statuses")
def get_statuses():
    try:
        conn = get_db_connection()
        cursor = conn.cursor()
        cursor.execute("SELECT * FROM agent_status ORDER BY role ASC")
        rows = cursor.fetchall()
        conn.close()
        return {"statuses": [dict(r) for r in rows]}
    except Exception as e:
        return {"error": str(e), "statuses": []}

@app.post("/api/clear")
def clear_all_messages():
    try:
        conn = get_db_connection()
        cursor = conn.cursor()
        cursor.execute("DELETE FROM messages")
        conn.commit()
        conn.close()
        return {"status": "success"}
    except Exception as e:
        return {"error": str(e)}

if __name__ == "__main__":
    print(f"Starting dashboard using DB at: {DB_PATH}")
    uvicorn.run("dashboard:app", host="0.0.0.0", port=6060, reload=True)
