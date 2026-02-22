import os
import sys
import json
from dotenv import load_dotenv
from urllib.request import Request, urlopen
from urllib.error import HTTPError

load_dotenv("/Users/datran/LearnDev/antigravity-kit/.env")
token = os.environ.get("FIGMA_ACCESS_TOKEN")

file_key = "kQtFAU9yfZsqIEILCdS0Zg"
node_id = "403:23395"

url = f"https://api.figma.com/v1/files/{file_key}"
print(f"Testing basic file access: {url}")

req = Request(url)
req.add_header("X-Figma-Token", token)

try:
    with urlopen(req) as response:
        data = json.loads(response.read().decode("utf-8"))
        print(f"SUCCESS: File name: {data.get('name')}")
except HTTPError as e:
    print(f"FAILED (Basic File): {e.code} - {e.read().decode('utf-8')}")

url_node = f"https://api.figma.com/v1/nodes/{file_key}?ids={node_id}"
print(f"Testing node access: {url_node}")
req_node = Request(url_node)
req_node.add_header("X-Figma-Token", token)

try:
    with urlopen(req_node) as response:
        data = json.loads(response.read().decode("utf-8"))
        print(f"SUCCESS: Node info: {list(data['nodes'].keys())}")
except HTTPError as e:
    print(f"FAILED (Node): {e.code} - {e.read().decode('utf-8')}")
