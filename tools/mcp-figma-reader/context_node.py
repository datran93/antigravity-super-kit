import os
import sys
from dotenv import load_dotenv

sys.path.append("/Users/datran/LearnDev/antigravity-kit/tools/mcp-figma-reader")
import server

load_dotenv("/Users/datran/LearnDev/antigravity-kit/.env")

target_id = "403:24078"

try:
    # Use fetch_figma_api with specific query to get info about this node
    # Since /nodes is 404, we have to rely on /files with depth
    print(f"Fetching metadata for {target_id}...")
    res = server.fetch_figma_api("files/kQtFAU9yfZsqIEILCdS0Zg", {"depth": "4"})

    def find_node(curr, tid):
        if curr.get("id") == tid:
            return curr
        for child in curr.get("children", []):
            found = find_node(child, tid)
            if found: return found
        return None

    node = find_node(res["document"], target_id)
    if node:
        print(f"Node Type: {node.get('type')}")
        print(f"Node Name: {node.get('name')}")

        # Check if it's an instance
        if node.get("type") == "INSTANCE":
            print(f"This is an instance of component: {node.get('name')}")

        # Check siblings
        # To find siblings we need the parent
        def find_parent(curr, tid):
            for child in curr.get("children", []):
                if child.get("id") == tid:
                    return curr
                found = find_parent(child, tid)
                if found: return found
            return None

        parent = find_parent(res["document"], target_id)
        if parent:
            print(f"Parent: {parent.get('name')} ({parent.get('type')})")
            print("Siblings in the same parent:")
            for sibling in parent.get("children", []):
                print(f"- {sibling.get('name')} [{sibling.get('type')}] ({sibling.get('id')})")

except Exception as e:
    print(f"ERROR: {str(e)}")
