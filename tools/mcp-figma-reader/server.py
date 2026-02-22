import os
import json
import re
from typing import Optional, List, Dict, Any
from urllib.request import Request, urlopen
from urllib.error import HTTPError
from mcp.server.fastmcp import FastMCP

# Initialize FastMCP Server
mcp = FastMCP("FigmaReader")

def get_token() -> str:
    """Retrieve Figma Access Token from environment."""
    token = os.environ.get("FIGMA_ACCESS_TOKEN")
    if not token:
        # Check if there is a .env file locally (simple check)
        try:
            with open(".env", "r") as f:
                for line in f:
                    if line.startswith("FIGMA_ACCESS_TOKEN="):
                        return line.split("=", 1)[1].strip().strip('"').strip("'")
        except:
            pass
    return token

def parse_figma_url(url: str) -> Dict[str, str]:
    """Extract file_key and node_id from Figma URL."""
    # Design file: https://www.figma.com/design/FILE_KEY/title?node-id=NODE_ID
    # Old format: https://www.figma.com/file/FILE_KEY/title?node-id=NODE_ID

    file_key_match = re.search(r"figma\.com/(?:file|design)/([a-zA-Z0-9]+)", url)
    node_id_match = re.search(r"node-id=([a-zA-Z0-9%-]+)", url)

    res = {}
    if file_key_match:
        res["file_key"] = file_key_match.group(1)
    if node_id_match:
        # Convert dash format (1-541) to colon format (1:541)
        node_id = node_id_match.group(1).replace("-", ":")
        res["node_id"] = node_id

    return res

def fetch_figma_api(endpoint: str, query_params: Dict[str, str] = None) -> Any:
    """Generic Figma API fetcher using urllib."""
    token = get_token()
    if not token:
        raise ValueError("Missing FIGMA_ACCESS_TOKEN. Please set it in your environment variables or .env file.")

    url = f"https://api.figma.com/v1/{endpoint}"
    if query_params:
        import urllib.parse
        url += "?" + urllib.parse.urlencode(query_params)

    req = Request(url)
    req.add_header("X-Figma-Token", token)

    try:
        with urlopen(req) as response:
            return json.loads(response.read().decode("utf-8"))
    except HTTPError as e:
        error_body = e.read().decode("utf-8")
        try:
            error_json = json.loads(error_body)
            error_msg = error_json.get("err", error_body)
        except:
            error_msg = error_body
        raise Exception(f"Figma API Error ({e.code}): {error_msg}")

@mcp.tool()
def read_figma_design(url: str, depth: int = 1) -> str:
    """
    Reads raw design data from a Figma URL.
    Returns the file structure and metadata.

    Args:
        url: The full Figma design URL.
        depth: How deep to traverse the file tree (default 1 for high-level structure).
    """
    try:
        parsed = parse_figma_url(url)
        if "file_key" not in parsed:
            return "❌ Invalid Figma URL. Could not extract file key."

        file_key = parsed["file_key"]
        params = {"depth": str(depth)}

        # If the URL has a specific node ID, we might want to fetch that branch specifically
        # but Figma's /v1/files/:key endpoint returns the whole file unless we use /v1/nodes/:key
        if "node_id" in parsed:
            result = fetch_figma_api(f"nodes/{file_key}", {"ids": parsed["node_id"]})
        else:
            result = fetch_figma_api(f"files/{file_key}", params)

        return json.dumps(result, indent=2)
    except Exception as e:
        return f"❌ Error: {str(e)}"

@mcp.tool()
def export_figma_images(url: str, format: str = "png", scale: float = 1.0) -> str:
    """
    Renders specific nodes from a Figma URL as images.
    Returns a map of node IDs to temporary image URLs.

    Args:
        url: The full Figma design URL (must contain a node-id to render specifically).
        format: Image format (png, jpg, svg, pdf).
        scale: Scale factor (0.01 to 4).
    """
    try:
        parsed = parse_figma_url(url)
        if "file_key" not in parsed:
            return "❌ Invalid Figma URL. Could not extract file key."

        if "node_id" not in parsed:
            return "❌ No node-id found in URL. Please select a specific element in Figma and copy the link to that element."

        file_key = parsed["file_key"]
        params = {
            "ids": parsed["node_id"],
            "format": format.lower(),
            "scale": str(scale)
        }

        result = fetch_figma_api(f"images/{file_key}", params)
        return json.dumps(result, indent=2)
    except Exception as e:
        return f"❌ Error: {str(e)}"

@mcp.tool()
def get_design_details(url: str, ids: Optional[str] = None) -> str:
    """
    Gets detailed JSON data for specific nodes in a Figma file.

    Args:
        url: The Figma file URL.
        ids: Comma-separated list of node IDs (optional if URL already has node-id).
    """
    try:
        parsed = parse_figma_url(url)
        if "file_key" not in parsed:
            return "❌ Invalid Figma URL."

        file_key = parsed["file_key"]
        target_ids = ids if ids else parsed.get("node_id")

        if not target_ids:
            return "❌ No Node IDs provided. Use `read_figma_design` first to find IDs."

        result = fetch_figma_api(f"nodes/{file_key}", {"ids": target_ids})
        return json.dumps(result, indent=2)
    except Exception as e:
        return f"❌ Error: {str(e)}"

if __name__ == "__main__":
    mcp.run(transport='stdio')
