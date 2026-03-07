import os
import requests
import sqlite3
import hashlib
import time
from datetime import datetime
import traceback
from contextlib import closing
from typing import List
from mcp.server.fastmcp import FastMCP
from duckduckgo_search import DDGS

# Khởi tạo MCP Server
mcp = FastMCP("McpDocResearcher")

CACHE_DB_PATH = os.path.join(os.path.dirname(os.path.abspath(__file__)), "research_cache.db")
CACHE_EXPIRY_SECONDS = 30 * 24 * 60 * 60 # 1 month

def get_db():
    conn = sqlite3.connect(CACHE_DB_PATH)
    conn.execute('''
        CREATE TABLE IF NOT EXISTS cache (
            key TEXT PRIMARY KEY,
            value TEXT,
            timestamp REAL
        )
    ''')
    return conn

def get_cache(key: str):
    with closing(get_db()) as conn:
        cursor = conn.cursor()
        cursor.execute("SELECT value, timestamp FROM cache WHERE key = ?", (key,))
        row = cursor.fetchone()
        if row:
            value, timestamp = row
            if time.time() - timestamp < CACHE_EXPIRY_SECONDS:
                return value
    return None

def set_cache(key: str, value: str):
    with closing(get_db()) as conn:
        conn.execute('''
            INSERT OR REPLACE INTO cache (key, value, timestamp)
            VALUES (?, ?, ?)
        ''', (key, value, time.time()))
        conn.commit()

def generate_cache_key(prefix: str, content: str) -> str:
    return f"{prefix}_{hashlib.md5(content.encode('utf-8')).hexdigest()}"

def fetch_jina_markdown(url: str) -> str:
    """Uses the free r.jina.ai API to fetch clean Markdown from any website."""
    try:
        jina_url = f"https://r.jina.ai/{url}"
        response = requests.get(jina_url, timeout=25)
        response.raise_for_status()
        return response.text
    except Exception as e:
        return f"Error fetching markdown from {url}: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def search_latest_syntax(topic: str, libraries: List[str] = []) -> str:
    """
    Search the real-time internet for the absolute latest SOTA (State Of The Art) syntax, best practices, and documentation for a specific programming topic or library.
    Use this tool before writing any logic to ensure you are not generating legacy code or using deprecated APIs.

    Args:
        topic: The specific concept to research (e.g., 'React server components data fetching', 'Next.js 14 App Router layout constraints', 'Zustand slices 2026').
        libraries: Optional list of specific libraries being used to narrow down the context.
    """
    try:
        search_query = topic
        if libraries:
            search_query += " " + " ".join(libraries)

        cache_key = generate_cache_key("search", search_query)
        cached_result = get_cache(cache_key)
        if cached_result:
            return f"⚡ (Cached - Loaded instantly from memory)\n{cached_result}"

        # 1. Search with DuckDuckGo
        ddgs = DDGS()
        results = list(ddgs.text(search_query + " tutorial OR documentation", region="us-en", backend="lite", max_results=3))

        if not results:
            return f"❌ No recent results found for: {topic}"

        final_report = [f"🔍 REAL-TIME RESEARCH RESULTS FOR: '{topic}'\n"]

        # 2. Extract quick search snippets
        final_report.append("### 1. QUICK SNIPPETS (SEARCH ENGINE RESULTS)")
        for idx, res in enumerate(results):
            final_report.append(f"{idx+1}. [{res.get('title', 'Unknown')}]({res.get('href', '')})")
            final_report.append(f"   Snippet: {res.get('body', '')}\n")

        # 3. Deep dive into the top 1 result using Jina Reader to get actual Markdown Code
        top_url = results[0].get('href', '')
        if top_url:
            final_report.append(f"### 2. DEEP DIVE (PARTIAL EXTRACTION OF TOP RESULT)")
            final_report.append(f"Source: {top_url}")

            # Cache the full version so subsequent read_website_markdown calls are instant
            cache_key_full = generate_cache_key("url_full", top_url)
            full_content = get_cache(cache_key_full)
            if not full_content:
                full_content = fetch_jina_markdown(top_url)
                if not full_content.startswith("Error fetching markdown"):
                    set_cache(cache_key_full, full_content)

            total_len = len(full_content)
            truncated = full_content[:4000]

            if total_len > 4000:
                final_report.append(f"Reading content... (Previewing first 4000/{total_len} characters)")
                final_report.append("\n```markdown\n" + truncated + "\n...\n```\n")
                final_report.append(f"\n💡 NOTICE: The full document has {total_len} characters. To read beyond this preview, use `read_website_markdown(url=\"{top_url}\", page=1)`.")
            else:
                final_report.append("Reading content...")
                final_report.append("\n```markdown\n" + truncated + "\n```\n")

        final_report.append("\n💡 ADVICE FOR AGENT: Synthesize these latest patterns and strictly apply them to your code generation. DO NOT use legacy patterns from your original training data if they conflict with these new docs.")

        final_report_str = "\n".join(final_report)
        set_cache(cache_key, final_report_str)
        return final_report_str

    except Exception as e:
        return f"❌ Error performing real-time research: {str(e)}\n{traceback.format_exc()}"

@mcp.tool()
def read_website_markdown(url: str, page: int = 1) -> str:
    """
    Scrape any specific documentation URL or website and return its content perfectly formatted as clean Markdown.
    Supports pagination for large documents. Each page returns up to 8000 characters.

    Args:
        url: The absolute URL including https:// (e.g. 'https://react.dev/reference/react/useActionState')
        page: The page number to read (default 1).
    """
    try:
        cache_key = generate_cache_key("url_full", url)
        content = get_cache(cache_key)

        if not content:
            content = fetch_jina_markdown(url)
            if not content.startswith("Error fetching markdown"):
                set_cache(cache_key, content)

        if content.startswith("Error fetching markdown"):
            return content

        chunk_size = 8000
        total_length = len(content)
        total_pages = max(1, (total_length + chunk_size - 1) // chunk_size)

        page = max(1, min(page, total_pages))

        start_idx = (page - 1) * chunk_size
        end_idx = min(start_idx + chunk_size, total_length)

        page_content = content[start_idx:end_idx]

        header = f"📄 Source: {url} | Page {page}/{total_pages}\n"
        header += "-" * 50 + "\n"

        footer = "\n" + "-" * 50 + "\n"
        if page < total_pages:
            footer += f"💡 (Page {page}/{total_pages}. There is more content. Extract the next page by calling this tool again with page={page+1})\n"
        else:
            footer += f"✅ (End of document)\n"

        return header + page_content + footer
    except Exception as e:
         return f"❌ Error scraping URL: {str(e)}\n{traceback.format_exc()}"

if __name__ == "__main__":
    mcp.run(transport='stdio')
