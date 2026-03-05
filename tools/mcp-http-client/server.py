import json
import requests
import os
import traceback
from datetime import datetime
from typing import Optional, Dict, Any, List
from mcp.server.fastmcp import FastMCP

# Initialize FastMCP server
mcp = FastMCP("McpHttpClient")

# Workspace Root Configuration
WORKSPACE_ROOT = "/Users/datran/LearnDev/antigravity-kit"
REST_DIR = os.path.join(WORKSPACE_ROOT, "rest")

def ensure_rest_dir():
    """Ensure that the 'rest' directory exists."""
    if not os.path.exists(REST_DIR):
        os.makedirs(REST_DIR, exist_ok=True)

# Session state to store environment-like configurations
session_state = {
    "base_url": "",
    "headers": {},
    "env": {}  # Support for dynamic variables {{var}}
}

def replace_vars(text: str) -> str:
    """Replace {{key}} with values from session_state['env']."""
    if not isinstance(text, str): return text
    import re
    for key, val in session_state["env"].items():
        text = text.replace(f"{{{{{key}}}}}", str(val))
    return text

def save_to_history(entry: Dict[str, Any], request_headers: Dict[str, str], slug: str = "general"):
    """Helper to append request history in {slug}.rest format."""
    ensure_rest_dir()
    # Sanitize slug
    safe_slug = "".join([c if c.isalnum() or c in ("-", "_") else "_" for c in slug.lower()])
    file_path = os.path.join(REST_DIR, f"{safe_slug}.rest")

    with open(file_path, "a", encoding="utf-8") as f:
        f.write(f"### {entry['method']} Request at {entry['timestamp']}\n")
        f.write(f"# Status: {entry['status_code']}\n")

        # Method and URL
        f.write(f"{entry['method']} {entry['url']}\n")

        # Headers
        for key, value in request_headers.items():
            f.write(f"{key}: {value}\n")

        # Body (if exists)
        if entry.get('json_body'):
            f.write("\n")
            f.write(json.dumps(entry['json_body'], indent=2))
            f.write("\n")

        f.write("\n")

@mcp.tool()
def http_request(
    method: str,
    path: str,
    params: Optional[Dict[str, Any]] = None,
    json_body: Optional[Dict[str, Any]] = None,
    headers: Optional[Dict[str, str]] = None,
    save_history: bool = True,
    slug: Optional[str] = None
) -> str:
    """
    Execute an HTTP request. Automatically prepends base_url and merges default headers.
    Supports placeholders like {{variable_name}}.
    """
    global session_state

    # 1. Substitute Environment Variables
    target_path = replace_vars(path)

    # 2. URL Logic
    url = target_path
    if session_state["base_url"] and not target_path.startswith("http"):
        base = session_state["base_url"].rstrip('/')
        path_clean = target_path.lstrip('/')
        url = f"{base}/{path_clean}"

    # 3. Header Logic
    request_headers = {}
    # Merge session headers (after variable replacement)
    for k, v in session_state["headers"].items():
        request_headers[replace_vars(k)] = replace_vars(v)

    # Merge request headers
    if headers:
        for k, v in headers.items():
            request_headers[replace_vars(k)] = replace_vars(v)

    try:
        response = requests.request(
            method=method.upper(),
            url=url,
            params=params,
            json=json_body,
            headers=request_headers,
            timeout=30
        )

        if save_history:
            final_slug = slug if slug else "general"
            save_to_history({
                "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                "method": method.upper(),
                "url": url,
                "json_body": json_body,
                "status_code": response.status_code
            }, request_headers, final_slug)

        status_icon = "✅" if 200 <= response.status_code < 300 else "⚠️" if 300 <= response.status_code < 400 else "❌"
        output = [
            f"{status_icon} **{method.upper()} {url}**",
            f"**Status**: `{response.status_code} {response.reason}`",
            "\n### 📝 Response Headers",
            "```json",
            json.dumps(dict(response.headers), indent=2),
            "```",
            "\n### 📦 Response Body"
        ]

        try:
            body_json = response.json()
            output.append("```json")
            output.append(json.dumps(body_json, indent=2))
            output.append("```")
        except:
            output.append(f"```text\n{response.text[:2000]}\n```")

        return "\n".join(output)

    except requests.exceptions.RequestException as e:
        return f"❌ **Request Error**: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def set_env(key: str, value: Any) -> str:
    """Set an environment variable for use in {{key}} placeholders."""
    global session_state
    session_state["env"][key] = value
    return f"✅ Env var `{key}` set to `{value}`"

@mcp.tool()
def import_curl(curl_command: str) -> str:
    """
    Parse a raw cURL command and execute it.
    Useful for quickly testing requests from documentation or browser.
    """
    import shlex
    import argparse

    parser = argparse.ArgumentParser()
    parser.add_argument('url')
    parser.add_argument('-X', '--request', default='GET')
    parser.add_argument('-d', '--data', '--json', action='append')
    parser.add_argument('-H', '--header', action='append')

    try:
        # Simple cleanup
        cmd = curl_command.replace('curl ', '', 1).replace('\\\n', ' ')
        args, unknown = parser.parse_known_args(shlex.split(cmd))

        method = args.request
        url = args.url
        headers = {}
        if args.header:
            for h in args.header:
                if ':' in h:
                    k, v = h.split(':', 1)
                    headers[k.strip()] = v.strip()

        json_body = None
        if args.data:
            combined_data = " ".join(args.data)
            try:
                json_body = json.loads(combined_data)
            except:
                # If not JSON, we'll try to send as raw data or ignore for now
                pass

        return http_request(method=method, path=url, json_body=json_body, headers=headers, slug="curl-import")
    except Exception as e:
        return f"❌ Error parsing cURL: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def list_history() -> str:
    """View context in .rest format."""
    if not os.path.exists(REST_DIR):
        return "📭 History directory (`rest/`) does not exist yet."

    files = [f for f in os.listdir(REST_DIR) if f.endswith(".rest")]
    if not files: return "📭 No .rest files found."

    output = ["### 📜 Request History List (.rest)\n"]
    for file in sorted(files):
        path = os.path.join(REST_DIR, file)
        try:
            with open(path, "r", encoding="utf-8") as f:
                lines = f.readlines()
                latest_req = ""
                for line in reversed(lines):
                    if line.startswith("###"):
                        latest_req = line.strip()
                        break
                output.append(f"- **`{file}`**: {latest_req if latest_req else 'Empty'}")
        except Exception as e:
            output.append(f"- **`{file}`**: Read error - {str(e)}")

    return "\n".join(output)

@mcp.tool()
def clear_history() -> str:
    """Clear all request history."""
    if os.path.exists(REST_DIR):
        import shutil
        shutil.rmtree(REST_DIR)
        os.makedirs(REST_DIR, exist_ok=True)
    return "✅ Successfully cleared everything in the rest directory."

@mcp.tool()
def set_config(base_url: Optional[str] = None, auth_token: Optional[str] = None) -> str:
    """Configure base URL and auth token."""
    global session_state
    updates = []
    if base_url:
        session_state["base_url"] = base_url
        updates.append(f"Base URL set to: `{base_url}`")
    if auth_token:
        session_state["headers"]["Authorization"] = f"Bearer {auth_token}"
        updates.append("Authorization header updated.")
    return "✅ Configuration Updated:\n- " + "\n- ".join(updates) if updates else "No configuration provided."

if __name__ == "__main__":
    mcp.run(transport='stdio')
