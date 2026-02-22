import os
import sys
import json
from dotenv import load_dotenv

sys.path.append("/Users/datran/LearnDev/antigravity-kit/tools/mcp-figma-reader")
import server

load_dotenv("/Users/datran/LearnDev/antigravity-kit/.env")

target_id = "403:24078"

try:
    print(f"Searching for neighbor components around {target_id}...")
    # Fetch the page itself with depth 2 to see the frames on that page
    res = server.fetch_figma_api("files/kQtFAU9yfZsqIEILCdS0Zg", {"depth": "4"})

    ui_page = None
    for page in res['document']['children']:
        if page['name'] == "UI Design":
            ui_page = page
            break

    if ui_page:
        print(f"Found Page: {ui_page['name']}")

        # Check all frames on this page
        print("\nFrames found on 'UI Design' page:")
        for frame in ui_page.get("children", []):
            fname = frame.get("name")
            fid = frame.get("id")
            print(f"- {fname} ({fid})")
            if fid == target_id:
                print("  ^ This is the requested node.")

            # If the name is "Default", maybe it's part of a component set
            if frame.get("type") == "COMPONENT_SET":
                 print(f"  [Component Set] with variants: {[c.get('name') for c in frame.get('children', [])]}")

except Exception as e:
    print(f"ERROR: {str(e)}")
