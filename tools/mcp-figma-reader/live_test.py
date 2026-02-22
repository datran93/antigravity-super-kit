import os
import sys
import json
from dotenv import load_dotenv

# Add tools dir to path
sys.path.append("/Users/datran/LearnDev/antigravity-kit/tools/mcp-figma-reader")
import server

# Load .env
load_dotenv("/Users/datran/LearnDev/antigravity-kit/.env")

url = "https://www.figma.com/design/kQtFAU9yfZsqIEILCdS0Zg/-Seinetime-AI--UI-Design?node-id=403-23395&p=f&t=6mMnT9BLlfQ8IprP-0"

print(f"Testing Figma URL: {url}")
try:
    # Use depth 1 to avoid massive payload for the first test
    result = server.read_figma_design(url, depth=1)
    print("SUCCESS: Received data from Figma")
    # Save a sample to a file so we can inspect it without flooding console
    with open("figma_sample.json", "w") as f:
        f.write(result)
    print("Saved raw design data to figma_sample.json")

    # Also try to get image
    print("Testing image export...")
    images = server.export_figma_images(url)
    print(f"Image results: {images}")

except Exception as e:
    print(f"FAILED: {str(e)}")
