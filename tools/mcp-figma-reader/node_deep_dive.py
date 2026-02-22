import os
import sys
import json
from dotenv import load_dotenv

sys.path.append("/Users/datran/LearnDev/antigravity-kit/tools/mcp-figma-reader")
import server

load_dotenv("/Users/datran/LearnDev/antigravity-kit/.env")

target_id = "403:24078"

try:
    print(f"Exporting image for {target_id} to verify visual content...")
    img_res = server.export_figma_images(f"https://www.figma.com/design/kQtFAU9yfZsqIEILCdS0Zg/test?node-id={target_id.replace(':', '-')}")
    print(f"Image Export Result: {img_res}")

    # Try fetching the node with depth 0 on the file to get metadata of all pages
    res = server.fetch_figma_api("files/kQtFAU9yfZsqIEILCdS0Zg", {"depth": "4"})

    def find_and_describe(curr, tid):
        if curr.get("id") == tid:
            return curr
        for child in curr.get("children", []):
            found = find_and_describe(child, tid)
            if found: return found
        return None

    node = find_and_describe(res["document"], target_id)
    if node:
        print(f"\nNode Metadata:")
        print(json.dumps({
            "name": node.get("name"),
            "type": node.get("type"),
            "id": node.get("id"),
            "child_count": len(node.get("children", []))
        }, indent=2))

        if node.get("children"):
            print("\nImmediate Children:")
            for c in node["children"]:
                print(f"- {c.get('name')} ({c.get('type')})")

except Exception as e:
    print(f"ERROR: {str(e)}")
