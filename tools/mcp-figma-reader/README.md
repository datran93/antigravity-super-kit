# Figma Reader MCP Server

This MCP server allows AI agents to read and interact with Figma designs directly. It can fetch raw JSON design data, metadata, and export specific components as images.

## Features

- **Read Design Structure**: Get the full JSON tree of a Figma file.
- **Node-Specific Data**: Fetch detailed data for specific elements.
- **Image Export**: Render Figma elements as PNG, SVG, JPG, or PDF URLs.
- **Auto URL Parsing**: Automatically extracts file keys and node IDs from Figma links.

## Setup

### 1. Get a Figma Personal Access Token
1. Open Figma and go to **Settings**.
2. Scroll down to the **Personal access tokens** section.
3. Type a name for your token (e.g., "AI Agent") and press Enter.
4. **Copy the token** immediately (it won't be shown again).

### 2. Configure Environment
Set the following environment variable in your system or add it to a `.env` file in your root workspace:

```bash
FIGMA_ACCESS_TOKEN=your_token_here
```

## Tools Provided

### `read_figma_design(url, depth=1)`
Fetches the JSON structure of a Figma file. Use `depth=1` for a quick overview of pages and top-level frames.

### `get_design_details(url, ids=None)`
Fetches deep JSON data for specific nodes. If the URL contains a `node-id`, it will target that automatically.

### `export_figma_images(url, format="png", scale=1.0)`
Returns a temporary URL of the rendered design element.

## Example
If you have a URL like:
`https://www.figma.com/design/ABC123XYZ/My-App?node-id=1-10`

The agent can call `read_figma_design` with this URL to understand the component at node `1:10`.
