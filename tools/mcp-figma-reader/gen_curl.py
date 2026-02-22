import os
import sys
import urllib.parse
from dotenv import load_dotenv

load_dotenv("/Users/datran/LearnDev/antigravity-kit/.env")
token = os.environ.get("FIGMA_ACCESS_TOKEN")

file_key = "kQtFAU9yfZsqIEILCdS0Zg"
node_id = "403:23396"

params = urllib.parse.urlencode({"ids": node_id})
url = f"https://api.figma.com/v1/nodes/{file_key}?{params}"

print(f"CURL COMMAND:")
print(f"curl -H 'X-Figma-Token: {token}' '{url}'")
