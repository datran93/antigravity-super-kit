import os
import sys
import json
from dotenv import load_dotenv

sys.path.append("/Users/datran/LearnDev/antigravity-kit/tools/mcp-figma-reader")
import server

load_dotenv("/Users/datran/LearnDev/antigravity-kit/.env")

url = "https://www.figma.com/design/kQtFAU9yfZsqIEILCdS0Zg/-Seinetime-AI--UI-Design?node-id=403-24078&t=jTXDWNVKR2fpb51i-0"
print(f"Fetching design node 403:24078...")

try:
    # Try fetching the node details
    res = server.fetch_figma_api("nodes/kQtFAU9yfZsqIEILCdS0Zg", {"ids": "403:24078"})

    if "nodes" in res and "403:24078" in res["nodes"]:
        node = res["nodes"]["403:24078"]["document"]
        print(f"Found Node: {node.get('name')} ({node.get('type')})")

        # List immediate children components
        children = node.get("children", [])
        print(f"\nComponents identified in this frame:")
        for c in children:
            print(f"- {c.get('name')} [{c.get('type')}]")
    else:
        # Fallback to fetching file and searching
        print("Node not found directly. Fetching file structure...")
        file_data = server.fetch_figma_api("files/kQtFAU9yfZsqIEILCdS0Zg", {"depth": "3"})
        # Simple search for node in the tree
        def find_node(curr, target_id):
            if curr.get("id") == target_id:
                return curr
            for child in curr.get("children", []):
                found = find_node(child, target_id)
                if found: return found
            return None

        target = find_node(file_data["document"], "403:24078")
        if target:
            print(f"Found Target: {target.get('name')}")
            for c in target.get("children", []):
                print(f"- {c.get('name')} [{c.get('type')}]")
        else:
            print("Target node not found in depth 3.")

except Exception as e:
    print(f"ERROR: {str(e)}")
