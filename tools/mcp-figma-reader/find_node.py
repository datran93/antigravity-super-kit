import os
import sys
import json
from dotenv import load_dotenv

sys.path.append("/Users/datran/LearnDev/antigravity-kit/tools/mcp-figma-reader")
import server

load_dotenv("/Users/datran/LearnDev/antigravity-kit/.env")

print(f"Fetching file structure...")
try:
    # Just the file metadata and pages
    res = server.fetch_figma_api("files/kQtFAU9yfZsqIEILCdS0Zg", {"depth": "1"})
    print(f"File: {res.get('name')}")
    for page in res.get('document', {}).get('children', []):
        print(f"Page: {page.get('name')} (ID: {page.get('id')})")
        # Check if the node ID prefix matches the page ID
        if page.get('id').split(':')[0] == "403":
             print(f"  -> Match prefix found in page {page.get('name')}")

except Exception as e:
    print(f"FAILED: {str(e)}")
