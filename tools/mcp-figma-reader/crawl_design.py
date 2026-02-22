import os
import sys
import json
from dotenv import load_dotenv

sys.path.append("/Users/datran/LearnDev/antigravity-kit/tools/mcp-figma-reader")
import server

load_dotenv("/Users/datran/LearnDev/antigravity-kit/.env")

target_id = "403:24078"
print(f"Searching for components in node {target_id} via file traversal...")

try:
    # Fetch file structure with sufficient depth
    # Depth 2 covers Page -> Top Level Frames.
    # Depth 3 covers Frames -> Components inside them.
    file_data = server.fetch_figma_api("files/kQtFAU9yfZsqIEILCdS0Zg", {"depth": "4"})

    def find_node(curr, tid):
        if curr.get("id") == tid:
            return curr
        for child in curr.get("children", []):
            res = find_node(child, tid)
            if res: return res
        return None

    target = find_node(file_data["document"], target_id)

    if target:
        print(f"\n✅ Found Frame: {target.get('name')}")
        print("Components/Elements found inside:")

        def list_components(curr, depth=0):
            indent = "  " * depth
            # Filter children to show meaningful parts
            for child in curr.get("children", []):
                name = child.get("name")
                ctype = child.get("type")
                print(f"{indent}- {name} [{ctype}]")
                if depth < 1: # Only go one level deeper for summary
                    list_components(child, depth + 1)

        list_components(target)
    else:
        print("❌ Could not find node 403:24078 in the file structure (Depth 4).")
        # List all top level frames in the first page to see context
        ui_page = file_data["document"]["children"][0]
        print(f"\nTop level frames in '{ui_page.get('name')}':")
        for frame in ui_page.get("children", []):
            print(f"- {frame.get('name')} ({frame.get('id')})")

except Exception as e:
    print(f"ERROR: {str(e)}")
