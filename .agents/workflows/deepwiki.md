---
description: Generate comprehensive, interactive repository wikis and deep knowledge bases, similar to Windsurf Deepwiki.
---

# 📖 Deepwiki Workflow (Automated Knowledge Base)

This workflow guides you (the AI Agent) to autonomously transform a codebase or a specific feature into a structured, highly navigable wiki. It combines high-level architectural overviews with deep implementation details, flow diagrams, and verifiable code citations.

## 🚀 Execution Phase

### Phase 1: Wiki Architecture Design 🏗️
- **Analyze**: Scan the repository/feature to detect the project type, domain layers, tech stack, and key components.
- **Catalogue**: Generate a hierarchical **Wiki Structure Plan** mapping out:
    - **Overview/Onboarding**: Senior-level insights, system purpose, and learning paths.
    - **Architecture & Data Flow**: High-level designs, state machines, and database schemas.
    - **Deep Dives**: System → Components → Methods.
- *Wait for user approval* on the proposed structure before generating the actual content.

### Phase 2: Knowledge Synthesis & Content Generation 🧠
- Act autonomously to iteratively generate content for each page defined in the approved catalogue.
- **For each page:**
    - **Traceability**: Trace actual code paths. Do NOT guess or hallucinate functionality.
    - **Visuals**: Create dark-mode Mermaid diagrams (e.g., state, sequence, class, flowchart) to visualize complex logic (minimum 1-2 per complex page).
    - **Citations**: Cite source files rigorously using `[filename:line_number](file_path)` format.
    - **Tone**: Professional, technical, concise, and structured (use bolding, lists, and tables).

### Phase 3: Repository Integration 🖇️
- Create a `WIKI/{slug}/` directory (where `{slug}` is the specific feature or topic name) to isolate and store all generated markdown pages for this execution.
- Create a local `WIKI/{slug}/INDEX.md` as the landing page for this specific topic.
- **Root Index**: Generate or update the repository's root `WIKI/INDEX.md` as the master table of contents.
- **External Linking**: Add a prominent link to the new `WIKI/{slug}/INDEX.md` in the root `WIKI/INDEX.md` so it can be accessed easily from the outside.
- **Internal Navigation**: Link all generated markdown pages together *within* the `{slug}` folder for seamless cross-navigation (e.g., adding "← Back to Index" or "Next: Component X" at the bottom of pages).

### Phase 4: Final Validation ✅
- Verify all relative links between the generated wiki pages.
- Check that all cited files and lines are accurate and actually exist in the current codebase.
- Ensure Mermaid diagrams have valid syntax and render correctly in standard markdown viewers (e.g., GitHub/GitLab).

### Phase 5: Mission Success 🏁
- Present the generated Wiki structure to the user.
- Share the exact path to the updated root `WIKI/INDEX.md`.
- Offer to explain specific "First Principles", design tradeoffs discovered, or answer follow-up questions.

## 🔴 Critical Constraints
1. **Verifiable Depth**: Every technical claim MUST have a source. No hand-waving like "this likely does X".
2. **First Principles**: Always explain *WHY* an architectural pattern or library was chosen before explaining *WHAT* it does.
3. **Structured Navigation**: The Wiki must be as easy to navigate as Wikipedia, with rich cross-linking.
4. **No Destructive Actions**: This workflow is strictly read-only for source code. Do NOT modify source code; only create/update files in the `WIKI/` directory.

---

## 📌 Usage Examples
`/deepwiki "Generate a complete onboarding wiki for this repository"`
`/deepwiki "Create a deep technical reference for the internal messaging system"`
`/deepwiki "Document the authentication flow and user session management"`
