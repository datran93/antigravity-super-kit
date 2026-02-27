import json
import decimal
import datetime
import uuid
import re
from mcp.server.fastmcp import FastMCP
from sqlalchemy import create_engine, inspect, text
from sqlalchemy.engine.reflection import Inspector

# Initialize MCP Server
mcp = FastMCP("McpDatabaseInspector")

# Cache engines to avoid creating new connections on every tool call
engines_cache = {}

def get_engine(connection_string: str):
    if connection_string not in engines_cache:
        # Create engine with some safe defaults
        engines_cache[connection_string] = create_engine(
            connection_string,
            pool_pre_ping=True,  # Check connection health before using
            pool_recycle=3600    # Prevent stale connections
        )
    return engines_cache[connection_string]

def get_inspector(connection_string: str) -> Inspector:
    """Helper method to return inspector from cached engine."""
    return inspect(get_engine(connection_string))

def json_serial(obj):
    """JSON serializer for objects not serializable by default json code"""
    if isinstance(obj, (datetime.datetime, datetime.date)):
        return obj.isoformat()
    if isinstance(obj, decimal.Decimal):
        return float(obj) # Handle money/numeric as float or str depending on needs
    if isinstance(obj, uuid.UUID):
        return str(obj)
    return str(obj)

@mcp.tool()
def list_tables(connection_string: str) -> str:
    """
    List all tables and views in the database.
    For Redis, it returns a summary of keys by exploring the database.

    Args:
        connection_string: SQLAlchemy connection string.
                           Example Postgres: postgresql://user:password@localhost:5432/dbname
                           Example MySQL: mysql+pymysql://user:password@localhost:3306/dbname
                           Example Redis: redis://localhost:6379/0
    """
    try:
        if connection_string.startswith("redis://") or connection_string.startswith("rediss://"):
            import redis
            r = redis.from_url(connection_string)
            dbsize = r.dbsize()
            # Scan a few keys to give a sample
            _, keys = r.scan(0, count=100)

            result = [f"🗄 REDIS DATABASE (Size: {dbsize} keys)\n"]
            result.append("### Sample Keys (up to 100):")
            for k in keys:
                result.append(f"- {k.decode('utf-8')}")

            if not keys:
                return "❌ No keys found in this Redis database."

            return "\n".join(result)

        inspector = get_inspector(connection_string)
        tables = inspector.get_table_names()
        views = inspector.get_view_names()

        result = [f"🗄 DATABASE ENTITIES\n"]
        result.append("### Tables:")
        for t in tables:
            result.append(f"- {t}")

        result.append("\n### Views:")
        for v in views:
            result.append(f"- {v}")

        if not tables and not views:
            return "❌ No tables or views found in this database."

        return "\n".join(result)
    except Exception as e:
        return f"❌ Error connecting or reading database: {str(e)}"

@mcp.tool()
def inspect_schema(connection_string: str, table_name: str) -> str:
    """
    Get detailed schema information for a specific table. Includes columns, types, primary keys, and foreign keys.
    For Redis, table_name acts as the key name. Returns key type and properties.

    Args:
        connection_string: SQLAlchemy connection string.
        table_name: The exact name of the table to inspect.
    """
    try:
        if connection_string.startswith("redis://") or connection_string.startswith("rediss://"):
            import redis
            r = redis.from_url(connection_string)
            if not r.exists(table_name):
                return f"❌ Key '{table_name}' does not exist in Redis."

            key_type_bytes = r.type(table_name)
            key_type = key_type_bytes.decode('utf-8') if isinstance(key_type_bytes, bytes) else key_type_bytes
            ttl = r.ttl(table_name)

            result = [f"📋 SCHEMA FOR KEY: `{table_name}`\n"]
            result.append(f"### Metadata:")
            result.append(f"- **Type**: {key_type}")
            result.append(f"- **TTL**: {ttl} seconds")

            result.append("\n### Details:")
            if key_type == "string":
                # Check length
                val_len = r.strlen(table_name)
                result.append(f"- String Length: {val_len}")
            elif key_type == "hash":
                hlen = r.hlen(table_name)
                result.append(f"- Hash Entries: {hlen}")
            elif key_type == "list":
                llen = r.llen(table_name)
                result.append(f"- List Length: {llen}")
            elif key_type == "set":
                scard = r.scard(table_name)
                result.append(f"- Set Members: {scard}")
            elif key_type == "zset":
                zcard = r.zcard(table_name)
                result.append(f"- Sorted Set Members: {zcard}")

            return "\n".join(result)

        inspector = get_inspector(connection_string)

        if table_name not in inspector.get_table_names() and table_name not in inspector.get_view_names():
            return f"❌ Table '{table_name}' does not exist in the database."

        columns = inspector.get_columns(table_name)
        pk_constraint = inspector.get_pk_constraint(table_name)
        foreign_keys = inspector.get_foreign_keys(table_name)

        result = [f"📋 SCHEMA FOR TABLE: `{table_name}`\n"]

        # Setup Primary Keys display
        pks = pk_constraint.get('constrained_columns', [])

        result.append("### Columns:")
        for col in columns:
            col_name = col['name']
            col_type = str(col['type'])
            nullable = "NULL" if col.get('nullable') else "NOT NULL"

            # Format row
            prefix = "🔑 " if col_name in pks else "   "
            result.append(f"{prefix}`{col_name}` : **{col_type}** ({nullable})")

        if foreign_keys:
            result.append("\n### Foreign Keys:")
            for fk in foreign_keys:
                constrained_cols = ", ".join(fk['constrained_columns'])
                referred_table = fk['referred_table']
                referred_cols = ", ".join(fk['referred_columns'])
                result.append(f"- `{constrained_cols}` -> `{referred_table}`.`{referred_cols}`")

        return "\n".join(result)
    except Exception as e:
        return f"❌ Error inspecting table '{table_name}': {str(e)}"

@mcp.tool()
def explain_query(connection_string: str, query: str) -> str:
    """
    Analyze query execution plan (EXPLAIN ANALYZE) for Postgres/MySQL/SQLite.
    Helps identify performance bottlenecks and index usage.
    """
    if connection_string.startswith(("redis://", "rediss://")):
        return "⚠️ EXPLAIN is not supported for Redis in this server."

    # Forbidden keywords for EXPLAIN (don't allow explain delete etc without confirm)
    forbidden = r'\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|GRANT|REVOKE|CREATE)\b'
    if re.search(forbidden, query, re.IGNORECASE):
        return "❌ SECURITY BLOCK: explain_query is only for SELECT statements."

    try:
        engine = get_engine(connection_string)
        db_type = engine.dialect.name

        # Dialect specific explain syntax
        explain_prefix = "EXPLAIN "
        if db_type == "postgresql":
            explain_prefix = "EXPLAIN (ANALYZE, VERBOSE, BUFFERS) "
        elif db_type == "mysql":
            explain_prefix = "EXPLAIN ANALYZE "

        full_query = explain_prefix + query

        with engine.connect() as conn:
            result = conn.execute(text(full_query))
            rows = result.fetchall()

            plan = "\n".join([str(row[0]) for row in rows])
            return f"🔍 **QUERY PLAN ({db_type.upper()})**\n\n```text\n{plan}\n```"

    except Exception as e:
        return f"❌ Error explaining query: {str(e)}"

@mcp.tool()
def get_table_sample(connection_string: str, table_name: str) -> str:
    """
    Retrieve schema (DDL) and 5 sample rows for matching a table structure.
    Returns result as a clean Markdown report.
    """
    if connection_string.startswith(("redis://", "rediss://")):
        return "⚠️ get_table_sample is not supported for Redis. Use list_tables instead."

    try:
        engine = get_engine(connection_string)
        db_type = engine.dialect.name

        # 1. Get Schema Info
        inspector = inspect(engine)
        columns = inspector.get_columns(table_name)
        pk = inspector.get_pk_constraint(table_name)

        schema_md = f"### 📊 Table Schema: `{table_name}`\n\n"
        schema_md += "| Column | Type | Nullable | Default | PK |\n"
        schema_md += "| :--- | :--- | :--- | :--- | :--- |\n"

        pk_cols = pk.get('constrained_columns', [])
        for col in columns:
            is_pk = "✅" if col['name'] in pk_cols else ""
            schema_md += f"| `{col['name']}` | `{col['type']}` | {col['nullable']} | `{col['default']}` | {is_pk} |\n"

        # 2. Get Sample Data
        sample_query = f"SELECT * FROM {table_name} LIMIT 5"
        with engine.connect() as conn:
            result = conn.execute(text(sample_query))
            rows = result.mappings().all()

            if not rows:
                schema_md += "\n*No sample data found (table is empty).*"
            else:
                schema_md += "\n### 📝 Sample Data (First 5 Rows)\n\n"
                headers = list(rows[0].keys())
                schema_md += "| " + " | ".join(headers) + " |\n"
                schema_md += "| " + " | ".join(["---"] * len(headers)) + " |\n"

                for row in rows:
                    vals = [json.dumps(row[h], default=json_serial) for h in headers]
                    schema_md += "| " + " | ".join(vals) + " |\n"

        return schema_md

    except Exception as e:
        return f"❌ Error getting sample for '{table_name}': {str(e)}"

@mcp.tool()
def run_read_query(connection_string: str, query: str, limit: int = 500, offset: int = 0) -> str:
    """
    Execute a read-only SQL query to preview raw data. Results are returned in JSON format.

    Args:
        connection_string: SQLAlchemy connection string.
        query: SQL string (usually SELECT).
        limit: Max rows to return (default 500, max 2000).
        offset: Number of rows to skip.
    """
    try:
        # Enforce max limit for safety
        limit = min(limit, 2000)

        # Handle Redis
        if connection_string.startswith(("redis://", "rediss://")):
            # ... (Existing Redis logic stays largely same but we can use list slicing)
            import redis
            r = redis.from_url(connection_string)
            query_parts = query.strip().split()
            if not query_parts: return "❌ Empty query"
            command = query_parts[0].upper()

            safe_commands = {
                'GET', 'MGET', 'HGET', 'HGETALL', 'HMGET', 'HKEYS', 'HVALS', 'HLEN',
                'LRANGE', 'LLEN', 'LINDEX', 'SMEMBERS', 'SCARD', 'SISMEMBER',
                'ZRANGE', 'ZCARD', 'ZSCORE', 'ZREVRANGE', 'TYPE', 'TTL', 'EXISTS',
                'SCAN', 'INFO', 'DBSIZE', 'PING'
            }

            if command not in safe_commands:
                return f"❌ SECURITY BLOCK: Command '{command}' is not allowed in Read-only mode."

            res = r.execute_command(*query_parts)

            def decode_redis(obj):
                if isinstance(obj, bytes): return obj.decode('utf-8', errors='replace')
                if isinstance(obj, (list, tuple, set)): return [decode_redis(i) for i in obj]
                if isinstance(obj, dict): return {decode_redis(k): decode_redis(v) for k, v in obj.items()}
                return obj

            decoded_res = decode_redis(res)
            # Apply pseudo-pagination for lists
            if isinstance(decoded_res, list):
                total = len(decoded_res)
                sliced = decoded_res[offset : offset + limit]
                return f"✅ REDIS RESULT ({offset}-{offset+len(sliced)} of {total})\n```json\n{json.dumps({'result': sliced}, indent=2)}\n```"

            return f"✅ REDIS RESULT\n```json\n{json.dumps({'result': decoded_res}, indent=2)}\n```"

        # Handle SQL
        forbidden = r'\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|GRANT|REVOKE|CREATE)\b'
        if re.search(forbidden, query, re.IGNORECASE):
            if not query.strip().upper().startswith("SELECT"):
                return "❌ SECURITY BLOCK: Only SELECT/Read-only queries are allowed in this tool."

        # Auto-inject LIMIT and OFFSET if not present and if it's a simple SELECT
        # This is a bit naive but helpful
        clean_query = query.strip()
        if "LIMIT" not in clean_query.upper() and clean_query.upper().startswith("SELECT"):
            clean_query = f"{clean_query} LIMIT {limit} OFFSET {offset}"

        engine = get_engine(connection_string)
        with engine.connect() as conn:
            result = conn.execute(text(clean_query))

            if result.returns_rows:
                rows = result.mappings().all()
                if not rows: return "✅ Query executed. 0 rows returned."

                data = [dict(row) for row in rows]
                json_result = json.dumps(data, default=json_serial, indent=2)

                suffix = f"\n*(Showing {len(rows)} rows starting at offset {offset})*"
                return f"✅ QUERY RESULTS\n```json\n{json_result}\n```" + suffix
            else:
                return f"✅ Query executed successfully. Rows affected: {result.rowcount}"

    except Exception as e:
        return f"❌ Error: {str(e)}"

@mcp.tool()
def run_write_query(connection_string: str, query: str, confirm: bool = False) -> str:
    """
    Execute a write SQL query (INSERT/UPDATE/DELETE/ALTER/DROP/CREATE).
    This tool allows modifying the database.
    AGENT: You MUST ask the user for explicit confirmation in the chat before calling this tool with confirm=True.

    Args:
        connection_string: SQLAlchemy connection string.
        query: The raw SQL query or Redis command to run.
        confirm: Must be True to execute. If False, the tool will return a request for confirmation.
    """
    if not confirm:
        return f"⚠️  CONFIRMATION REQUIRED: You are about to execute a WRITE/DML operation:\n\n`{query}`\n\nPlease confirm if you want to proceed. Set 'confirm=True' only after user approval."

    try:
        # Handle Redis
        if connection_string.startswith("redis://") or connection_string.startswith("rediss://"):
            import redis
            r = redis.from_url(connection_string)
            query_parts = query.strip().split()
            if not query_parts:
                return "❌ Empty query"

            # For Redis, we allow all commands in this tool
            res = r.execute_command(*query_parts)

            # Decode recursive function for bytes
            def decode_redis(obj):
                if isinstance(obj, bytes):
                    return obj.decode('utf-8', errors='replace')
                elif isinstance(obj, list) or isinstance(obj, tuple) or isinstance(obj, set):
                    return [decode_redis(i) for i in obj]
                elif isinstance(obj, dict):
                    return {decode_redis(k): decode_redis(v) for k, v in obj.items()}
                return obj

            decoded_res = decode_redis(res)
            return f"✅ REDIS WRITE SUCCESS\nCMD = `{query}`\nResult: {json.dumps(decoded_res, indent=2) if isinstance(decoded_res, (dict, list)) else decoded_res}"

        # Handle SQL
        engine = get_engine(connection_string)

        try:
            # Process multi-statement queries
            # We strip comments and split by semicolon, filter empty lines
            statements = [s.strip() for s in query.split(';') if s.strip()]

            if not statements:
                return "❌ No valid SQL statements found."

            total_rows_affected = 0

            # Using engine.connect() with an explicit transaction
            with engine.connect() as conn:
                with conn.begin():
                    for i, stmt in enumerate(statements):
                        # Use execution_options(autocommit=True) for statements that
                        # cannot run in a transaction (like VACUUM or CREATE DATABASE)
                        # but warning: conn.begin() already started a transaction.
                        # For now, we assume standard DML.
                        result = conn.execute(text(stmt))

                        # Accumulate rowcount if available
                        if result.rowcount > 0:
                            total_rows_affected += result.rowcount

            return f"✅ WRITE SUCCESS\nExecuted {len(statements)} statement(s).\nTotal rows affected: {total_rows_affected}"

        except Exception as db_err:
            # Re-raise or return specific error
            error_msg = str(db_err)
            if "not exist" in error_msg.lower():
                return f"❌ Table or column does not exist: {error_msg}"
            if "duplicate key" in error_msg.lower() or "integrity" in error_msg.lower():
                return f"❌ Integrity Error (Constraint violation): {error_msg}"
            return f"❌ SQLAlchemy Error: {error_msg}"

    except Exception as e:
        return f"❌ Error executing write query: {str(e)}"

if __name__ == "__main__":
    mcp.run(transport='stdio')
