import os
from dotenv import load_dotenv
load_dotenv("/Users/datran/LearnDev/antigravity-kit/.env")

import glob
import re
import yaml
import chromadb
from chromadb.utils import embedding_functions
import hashlib
from typing import Dict

# Config
SKILLS_DIR = "/Users/datran/LearnDev/antigravity-kit/.agent/skills"
DB_DIR = "/Users/datran/LearnDev/antigravity-kit/tools/mcp-skill-router/.chroma_db"

def parse_skill_file(filepath: str) -> Dict:
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    # Extract frontmatter
    match = re.search(r'^---\n(.*?)\n---', content, re.DOTALL)
    metadata = {}
    if match:
        try:
            metadata = yaml.safe_load(match.group(1))
        except Exception:
            pass

    # Clean and extract preview
    text_content = re.sub(r'^---\n(.*?)\n---', '', content, flags=re.DOTALL).strip()
    preview = " ".join(text_content.split())[:250] + "..." if len(text_content) > 250 else text_content

    skill_name = os.path.basename(os.path.dirname(filepath))
    description = metadata.get('description', '')
    tags_list = metadata.get('tags', [])
    if isinstance(tags_list, list):
        tags_str = ', '.join(tags_list)
    else:
        tags_str = str(tags_list)

    search_text = f"Skill: {skill_name}\nTags: {tags_str}\nDescription: {description}\n\nPreview: {preview}"

    desc_clean = str(description).replace('\n', ' ')[:150] if description else ''

    # Calculate hash to detect changes
    file_hash = hashlib.md5(content.encode('utf-8')).hexdigest()

    return {
        "id": skill_name,
        "text": search_text,
        "metadata": {
            "name": skill_name,
            "description": desc_clean,
            "tags": tags_str,
            "path": filepath,
            "preview": preview,
            "hash": file_hash
        }
    }

def build_index():
    api_key = os.environ.get("OPENAI_API_KEY")
    if not api_key:
        print("❌ ERROR: OPENAI_API_KEY environment variable is not set.")
        return

    print("🔄 Initializing ChromaDB with OpenAI Embeddings (text-embedding-3-small)...")
    client = chromadb.PersistentClient(path=DB_DIR)

    emb_fn = embedding_functions.OpenAIEmbeddingFunction(
        api_key=api_key,
        model_name="text-embedding-3-small"
    )

    collection = client.get_or_create_collection(
        name="skills_collection_openai",
        embedding_function=emb_fn
    )

    print(f"📖 Scanning SKILL.md files in {SKILLS_DIR}...")
    skill_files = glob.glob(os.path.join(SKILLS_DIR, "*/SKILL.md"))

    docs = []
    ids = []
    metadatas = []

    # Delta Indexing
    try:
        existing = collection.get(include=["metadatas"])
        existing_hashes = {m['name']: m.get('hash') for m in existing['metadatas']} if existing and existing['metadatas'] else {}
    except Exception:
        existing_hashes = {}

    current_ids = []

    for filepath in skill_files:
        parsed = parse_skill_file(filepath)
        skill_id = parsed["id"]
        current_ids.append(skill_id)

        # Skip if unchanged
        if existing_hashes.get(skill_id) == parsed["metadata"]["hash"]:
            continue

        docs.append(parsed["text"])
        ids.append(parsed["id"])
        metadatas.append(parsed["metadata"])

    # Clean up stale skills
    to_delete = [id for id in existing_hashes.keys() if id not in current_ids]
    if to_delete:
        collection.delete(ids=to_delete)
        print(f"🗑️ Deleted {len(to_delete)} removed skills.")

    if not docs:
        print("✅ No skills changed. Vector DB is already up to date!")
        return

    print(f"✅ Found {len(docs)} new or modified skills. Starting vector DB ingestion...")

    batch_size = 100
    for i in range(0, len(docs), batch_size):
        end = min(i + batch_size, len(docs))
        print(f"⏳ Processing batch {i+1} to {end}...")
        try:
            collection.upsert(
                documents=docs[i:end],
                ids=ids[i:end],
                metadatas=metadatas[i:end]
            )
        except Exception as e:
            print(f"❌ Error at batch {i+1}-{end}: {str(e)}")

    print("🎉 Finished index updates!")

if __name__ == "__main__":
    build_index()
