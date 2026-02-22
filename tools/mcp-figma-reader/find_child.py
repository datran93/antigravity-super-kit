import os
import sys
import json
from dotenv import load_dotenv

sys.path.append("/Users/datran/LearnDev/antigravity-kit/tools/mcp-figma-reader")
import server

load_dotenv("/Users/datran/LearnDev/antigravity-kit/.env")

print(f"Fetching depth 2 children...")
try:
    res = server.fetch_figma_api("files/kQtFAU9yfZsqIEILCdS0Zg", {"depth": "2"})
    page = res['document']['children'][0] # UI Design page
    print(f"Page: {page['name']} ({page['id']})")
    child = page['children'][0]
    print(f"Child: {child['name']} ({child['id']})")

    print(f"Testing node access for child: {child['id']}")
    res_node = server.fetch_figma_api("nodes/kQtFAU9yfZsqIEILCdS0Zg", {"ids": child['id']})
    print(f"SUCCESS: Found node keys: {list(res_node.get('nodes', {}).keys())}")

except Exception as e:
    print(f"FAILED: {str(e)}")
